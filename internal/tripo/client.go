// Package tripo is trident-mcp's single module for the Tripo 3D API (v3).
//
// It owns everything Tripo-specific: the HTTP transport and error envelope,
// the Operations that create Tasks, Input resolution (local files are
// uploaded), Task polling and Artifact downloads, the model-version Catalog,
// and account queries. Callers depend on *Client; the only seam is HTTP
// (Options.BaseURL and the http.Clients), which tests fill with tripotest.
package tripo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the global Tripo v3 API host.
	DefaultBaseURL = "https://openapi.tripo3d.ai/v3"
	// ChinaBaseURL is the mainland-China host. API keys are valid in one region only.
	ChinaBaseURL = "https://openapi.tripo3d.com/v3"

	defaultMaxRetries = 5
	maxRetryWait      = 32 * time.Second
	maxResponseBytes  = 32 << 20
)

// Options configures a Client.
type Options struct {
	// APIKey is the Tripo key (tsk_...). Required.
	APIKey string
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// HTTPClient sends API calls. Defaults to a client with a 60 s timeout.
	HTTPClient *http.Client
	// TransferClient downloads Artifacts and uploads presigned files. It has no
	// timeout by default; transfers are bounded by their context instead.
	TransferClient *http.Client
	// MaxRetries bounds retries of rate-limited (HTTP 429) calls. Zero means
	// the default of 5; negative disables retries.
	MaxRetries int
}

// Client talks to the Tripo v3 API.
type Client struct {
	apiKey     string
	baseURL    string
	http       *http.Client
	transfer   *http.Client
	maxRetries int
}

// New creates a Client.
func New(opts Options) (*Client, error) {
	if strings.TrimSpace(opts.APIKey) == "" {
		return nil, fmt.Errorf("tripo: API key is required")
	}
	c := &Client{
		apiKey:     strings.TrimSpace(opts.APIKey),
		baseURL:    strings.TrimRight(opts.BaseURL, "/"),
		http:       opts.HTTPClient,
		transfer:   opts.TransferClient,
		maxRetries: opts.MaxRetries,
	}
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: 60 * time.Second}
	}
	if c.transfer == nil {
		// No overall timeout, so large transfers finish; but bound the phases
		// before the body so a stalled host cannot hang a call forever.
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.ResponseHeaderTimeout = 60 * time.Second
		t.ExpectContinueTimeout = 5 * time.Second
		c.transfer = &http.Client{Transport: t}
	}
	switch {
	case c.maxRetries == 0:
		c.maxRetries = defaultMaxRetries
	case c.maxRetries < 0:
		c.maxRetries = 0
	}
	return c, nil
}

// BaseURL reports the API host the Client talks to.
func (c *Client) BaseURL() string { return c.baseURL }

// Error is a failed Tripo API call, carrying Tripo's error envelope.
type Error struct {
	HTTPStatus int
	Code       int
	Message    string
	Suggestion string
	RequestID  string
	// RetryAfter is Tripo's requested wait on rate-limit and
	// concurrency-limit (HTTP 429) errors.
	RetryAfter time.Duration

	hasRetryAfter bool
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.HTTPStatus)
	}
	var details []string
	if e.Code != 0 {
		details = append(details, "code "+strconv.Itoa(e.Code))
	}
	if e.HTTPStatus != 0 {
		details = append(details, "HTTP "+strconv.Itoa(e.HTTPStatus))
	}
	if e.RequestID != "" {
		details = append(details, "request "+e.RequestID)
	}
	out := "tripo: " + msg
	if len(details) > 0 {
		out += " (" + strings.Join(details, ", ") + ")"
	}
	if e.Suggestion != "" {
		out += ". Suggestion: " + e.Suggestion
	}
	return out
}

type envelope struct {
	Code       int             `json:"code"`
	Data       json.RawMessage `json:"data"`
	Message    string          `json:"message"`
	Suggestion string          `json:"suggestion"`
	RequestID  string          `json:"request_id"`
}

// call sends a JSON request and decodes the envelope's data into out.
func (c *Client) call(ctx context.Context, method, path string, body, out any) error {
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return fmt.Errorf("tripo: encoding %s request: %w", path, err)
		}
	}
	return c.send(ctx, c.http, func() (*http.Request, error) {
		var r io.Reader
		if payload != nil {
			r = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, r)
		if err == nil && payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, err
	}, out)
}

// send performs an authenticated request built by build through client,
// retrying HTTP 429 responses. A 429 means Tripo did not accept the call, so
// retrying a task-creating request cannot bill twice; nothing else is retried.
func (c *Client) send(ctx context.Context, client *http.Client, build func() (*http.Request, error), out any) error {
	for attempt := 0; ; attempt++ {
		req, err := build()
		if err != nil {
			return fmt.Errorf("tripo: building request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("tripo: %s %s: %w", req.Method, req.URL.Path, err)
		}
		apiErr, err := decodeEnvelope(resp, out)
		if err != nil || apiErr == nil {
			return err
		}
		if apiErr.HTTPStatus != http.StatusTooManyRequests || attempt >= c.maxRetries {
			return apiErr
		}
		wait := apiErr.RetryAfter
		if !apiErr.hasRetryAfter {
			wait = time.Second << min(attempt, 5)
		}
		if wait > maxRetryWait {
			return apiErr // Tripo asked for longer than a tool call should block
		}
		select {
		case <-ctx.Done():
			return apiErr
		case <-time.After(wait):
		}
	}
}

func decodeEnvelope(resp *http.Response, out any) (*Error, error) {
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("tripo: reading response: %w", err)
	}

	var env envelope
	jsonErr := json.Unmarshal(raw, &env)
	if resp.StatusCode >= 400 || jsonErr != nil || env.Code != 0 {
		apiErr := &Error{
			HTTPStatus: resp.StatusCode,
			Code:       env.Code,
			Message:    env.Message,
			Suggestion: env.Suggestion,
			RequestID:  env.RequestID,
		}
		apiErr.RetryAfter, apiErr.hasRetryAfter = retryAfter(resp.Header.Get("Retry-After"))
		if apiErr.Message == "" && jsonErr != nil {
			apiErr.Message = summarizeBody(raw)
		}
		return apiErr, nil
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return nil, fmt.Errorf("tripo: decoding response data: %w", err)
		}
	}
	return nil, nil
}

func retryAfter(header string) (time.Duration, bool) {
	seconds, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || seconds < 0 {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

// summarizeBody turns a non-JSON error body (such as an nginx HTML page) into
// a short single-line message.
func summarizeBody(raw []byte) string {
	text := strings.Join(strings.Fields(string(raw)), " ")
	if len(text) > 200 {
		text = text[:200] + "…"
	}
	if text == "" {
		return "empty response body"
	}
	return text
}
