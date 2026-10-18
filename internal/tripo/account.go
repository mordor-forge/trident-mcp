package tripo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Balance is the account's credit balance. Frozen credits are held by
// running Tasks until they succeed (charged) or fail (refunded).
type Balance struct {
	Balance float64 `json:"balance"`
	Frozen  float64 `json:"frozen"`
}

// GetBalance returns the account's credits.
func (c *Client) GetBalance(ctx context.Context) (*Balance, error) {
	var b Balance
	if err := c.call(ctx, http.MethodGet, "/account/balance", nil, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// UsageRecord is the credit cost of one Task.
type UsageRecord struct {
	TaskID          string     `json:"task_id"`
	Type            string     `json:"type,omitempty"`
	Status          string     `json:"status,omitempty"`
	CreditsConsumed float64    `json:"credits_consumed"`
	CreatedAt       FlexString `json:"created_at,omitempty"`
}

// Usage is a page of credit usage, newest first.
type Usage struct {
	Records []UsageRecord `json:"records"`
}

// GetUsage returns per-Task credit usage. limit and offset page through the
// history; zero leaves them to Tripo's defaults.
func (c *Client) GetUsage(ctx context.Context, limit, offset int) (*Usage, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	path := "/account/usage"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var raw json.RawMessage
	if err := c.call(ctx, http.MethodGet, path, nil, &raw); err != nil {
		return nil, err
	}

	// Tripo documents a bare array; newer responses wrap it in {items: [...]}.
	usage := &Usage{Records: []UsageRecord{}}
	if len(raw) == 0 {
		return usage, nil
	}
	if err := json.Unmarshal(raw, &usage.Records); err == nil {
		return usage, nil
	}
	var wrapped struct {
		Items   []UsageRecord `json:"items"`
		Records []UsageRecord `json:"records"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("tripo: decoding account usage: %w", err)
	}
	if wrapped.Items != nil {
		usage.Records = wrapped.Items
	} else if wrapped.Records != nil {
		usage.Records = wrapped.Records
	}
	return usage, nil
}
