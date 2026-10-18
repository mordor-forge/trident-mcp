package tripo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Status is where a Task is in its lifecycle.
type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
	StatusBanned    Status = "banned"
	StatusExpired   Status = "expired"
	StatusUnknown   Status = "unknown"
)

// Terminal reports whether the Task will not change status again.
func (s Status) Terminal() bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled, StatusBanned, StatusExpired:
		return true
	default:
		return false
	}
}

// Task is Tripo's record of one Operation run, as returned by GET /tasks/{id}.
// Output is kept verbatim so new Tripo output fields survive without code
// changes; use OutputURLs, Riggable and RigType for the well-known ones.
type Task struct {
	TaskID          string         `json:"task_id"`
	Type            string         `json:"type,omitempty"`
	Status          Status         `json:"status"`
	Progress        int            `json:"progress"`
	Output          map[string]any `json:"output,omitempty"`
	ErrorCode       FlexString     `json:"error_code,omitempty"`
	ErrorMessage    string         `json:"error_message,omitempty"`
	CreditsConsumed float64        `json:"credits_consumed,omitempty"`
	RunningLeftTime *int           `json:"running_left_time,omitempty" jsonschema:"Estimated seconds until the task finishes, when Tripo reports it"`
	QueuingNum      *int           `json:"queuing_num,omitempty" jsonschema:"Position in Tripo's queue, when Tripo reports it"`
	CreatedAt       FlexString     `json:"created_at,omitempty"`
	CompletedAt     FlexString     `json:"completed_at,omitempty"`
}

// UnmarshalJSON accepts the v3 task object, falling back to the older
// error_msg / error.message fields for the failure reason.
func (t *Task) UnmarshalJSON(data []byte) error {
	type plain Task
	var decoded struct {
		plain
		ErrorMsg string `json:"error_msg"`
		Error    *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*t = Task(decoded.plain)
	if t.ErrorMessage == "" {
		t.ErrorMessage = decoded.ErrorMsg
	}
	if t.ErrorMessage == "" && decoded.Error != nil {
		t.ErrorMessage = decoded.Error.Message
	}
	return nil
}

// FlexString decodes a JSON string or number as a string. Tripo is not
// consistent about timestamp and error-code types.
type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("expected string or number, got %s", data)
	}
	*f = FlexString(n.String())
	return nil
}

// OutputURLs returns every URL in the Task's Output, keyed by its field name
// without the "_url" suffix ("model", "rendered_image", "front_view", ...).
// Nested objects are searched too, since some Operations group their URLs.
func (t *Task) OutputURLs() map[string]string {
	urls := map[string]string{}
	collectURLs(t.Output, urls)
	return urls
}

func collectURLs(m map[string]any, into map[string]string) {
	for key, value := range m {
		switch v := value.(type) {
		case string:
			if name, ok := strings.CutSuffix(key, "_url"); ok && isHTTPURL(v) {
				into[name] = v
			}
		case []any:
			name, ok := strings.CutSuffix(key, "_urls")
			if !ok {
				continue
			}
			for i, item := range v {
				if s, ok := item.(string); ok && isHTTPURL(s) {
					into[fmt.Sprintf("%s_%d", name, i+1)] = s
				}
			}
		case map[string]any:
			collectURLs(v, into)
		}
	}
}

// Riggable reports a rig_check Task's verdict; ok is false when the Output
// has none.
func (t *Task) Riggable() (riggable, ok bool) {
	riggable, ok = t.Output["riggable"].(bool)
	return riggable, ok
}

// RigType reports the skeleton type a rig_check Task detected.
func (t *Task) RigType() string {
	for _, key := range []string{"rig_type", "topology"} {
		if s, ok := t.Output[key].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

var validTaskID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func checkTaskID(id string) error {
	if !validTaskID.MatchString(id) {
		return fmt.Errorf("tripo: invalid task ID %q: expected letters, digits, '-' or '_'", id)
	}
	return nil
}

// GetTask fetches a Task. Output URLs in the result are signed and expire;
// Download re-fetches the Task for fresh ones.
func (c *Client) GetTask(ctx context.Context, id string) (*Task, error) {
	if err := checkTaskID(id); err != nil {
		return nil, err
	}
	var task Task
	if err := c.call(ctx, http.MethodGet, "/tasks/"+id, nil, &task); err != nil {
		return nil, err
	}
	if task.TaskID == "" {
		task.TaskID = id
	}
	return &task, nil
}

// TaskList is the result of GetTasks.
type TaskList struct {
	Tasks  map[string]*Task `json:"tasks"`
	Missed []string         `json:"missed"`
}

// GetTasks fetches up to 100 Tasks in one call. IDs Tripo does not know are
// listed in Missed.
func (c *Client) GetTasks(ctx context.Context, ids []string) (*TaskList, error) {
	if len(ids) == 0 || len(ids) > 100 {
		return nil, fmt.Errorf("tripo: between 1 and 100 task IDs are required, got %d", len(ids))
	}
	for _, id := range ids {
		if err := checkTaskID(id); err != nil {
			return nil, err
		}
	}
	var list TaskList
	if err := c.call(ctx, http.MethodPost, "/tasks/list", map[string]any{"task_ids": ids}, &list); err != nil {
		return nil, err
	}
	if list.Tasks == nil {
		list.Tasks = map[string]*Task{}
	}
	if list.Missed == nil {
		list.Missed = []string{}
	}
	sort.Strings(list.Missed)
	return &list, nil
}

// WaitOptions tunes WaitTask.
type WaitOptions struct {
	// Interval between polls. Defaults to 2 s.
	Interval time.Duration
	// OnUpdate, when set, is called with every polled Task.
	OnUpdate func(*Task)
}

// WaitTask polls a Task until it reaches a terminal Status or ctx ends. When
// ctx ends first it returns the last Task seen together with ctx.Err(), so
// callers can report progress and wait again.
func (c *Client) WaitTask(ctx context.Context, id string, opts WaitOptions) (*Task, error) {
	if err := checkTaskID(id); err != nil {
		return nil, err
	}
	interval := opts.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	var last *Task
	failures := 0
	for {
		task, err := c.GetTask(ctx, id)
		switch {
		case err == nil:
			failures, last = 0, task
			if opts.OnUpdate != nil {
				opts.OnUpdate(task)
			}
			if task.Status.Terminal() {
				return task, nil
			}
		case ctx.Err() != nil:
			return last, ctx.Err()
		case !transient(err) || failures >= maxPollFailures:
			return last, err
		default:
			failures++
		}
		if !sleepCtx(ctx, interval) {
			return last, ctx.Err()
		}
	}
}

// maxPollFailures is how many consecutive transient errors WaitTask rides out.
const maxPollFailures = 3

// transient reports whether a failed poll is worth repeating: network errors
// and Tripo 5xx responses, but not 4xx such as an unknown task.
func transient(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatus >= 500
	}
	return true
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
