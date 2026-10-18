// Package tripotest is an in-memory fake of the Tripo v3 API for tests. It
// creates Tasks for every task-creating endpoint, serves Tasks, uploads,
// account queries and CDN downloads, and records every API request.
package tripotest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

// APIKey is the key the fake accepts.
const APIKey = "tsk_test"

// Request is one recorded API call.
type Request struct {
	Method string
	Path   string
	Query  string
	// Body is the decoded JSON body, if any.
	Body map[string]any
	// FileName and FileBytes are set for multipart uploads.
	FileName  string
	FileBytes int
}

// Response is a canned reply for the next call to a path.
type Response struct {
	Status int
	Body   any
	Header http.Header
}

// Server is the fake. The zero Balance is 1000 credits.
type Server struct {
	*httptest.Server

	mu       sync.Mutex
	requests []Request
	tasks    map[string][]map[string]any // task ID → states served in order; the last repeats
	files    map[string]cdnFile
	queued   map[string][]Response
	nextID   int
	Balance  map[string]any
	Usage    any
}

type cdnFile struct {
	contentType string
	data        []byte
	// truncated files drop the connection halfway through the body.
	truncated bool
}

// New starts a fake and closes it when the test ends.
func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{
		tasks:   map[string][]map[string]any{},
		files:   map[string]cdnFile{},
		queued:  map[string][]Response{},
		Balance: map[string]any{"balance": 1000.0, "frozen": 0.0},
		Usage:   []any{},
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.Close)
	return s
}

// Client returns a tripo.Client wired to the fake, with fast retries.
func (s *Server) Client(t testing.TB) *tripo.Client {
	t.Helper()
	c, err := tripo.New(tripo.Options{APIKey: APIKey, BaseURL: s.URL, HTTPClient: s.Server.Client(), TransferClient: s.Server.Client()})
	if err != nil {
		t.Fatalf("tripotest: creating client: %v", err)
	}
	return c
}

// Requests returns every recorded API call, oldest first.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

// Last returns the most recent call to path, failing the test if there is none.
func (s *Server) Last(t testing.TB, path string) Request {
	t.Helper()
	reqs := s.Requests()
	for i := len(reqs) - 1; i >= 0; i-- {
		if reqs[i].Path == path {
			return reqs[i]
		}
	}
	t.Fatalf("tripotest: no request to %s; got %v", path, paths(reqs))
	return Request{}
}

func paths(reqs []Request) []string {
	out := make([]string, len(reqs))
	for i, r := range reqs {
		out[i] = r.Method + " " + r.Path
	}
	return out
}

// SetTask makes GET /tasks/{id} serve the given states in order, repeating
// the last one. Missing task_id and progress fields are filled in.
func (s *Server) SetTask(id string, states ...map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, st := range states {
		if _, ok := st["task_id"]; !ok {
			st["task_id"] = id
		}
		if _, ok := st["progress"]; !ok {
			st["progress"] = 0
			if st["status"] == "success" {
				st["progress"] = 100
			}
		}
	}
	s.tasks[id] = states
}

// Succeed makes a Task successful with the given output.
func (s *Server) Succeed(id string, output map[string]any) {
	s.SetTask(id, map[string]any{"status": "success", "output": output})
}

// AddFile serves data at /cdn/<name> and returns its URL.
func (s *Server) AddFile(name, contentType string, data []byte) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files["/cdn/"+name] = cdnFile{contentType: contentType, data: data}
	return s.URL + "/cdn/" + name
}

// AddTruncatedFile serves the first half of data at /cdn/<name>, then drops
// the connection, and returns its URL.
func (s *Server) AddTruncatedFile(name string, data []byte) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files["/cdn/"+name] = cdnFile{contentType: "application/octet-stream", data: data, truncated: true}
	return s.URL + "/cdn/" + name
}

// Queue makes the next call to path reply with resp instead of the default.
func (s *Server) Queue(path string, resp Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queued[path] = append(s.queued[path], resp)
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/cdn/") {
		s.serveFile(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/presigned/") {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		return
	}

	rec := Request{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if file, header, err := r.FormFile("file"); err == nil {
			data, _ := io.ReadAll(file)
			rec.FileName, rec.FileBytes = header.Filename, len(data)
		}
	} else if r.Body != nil {
		data, _ := io.ReadAll(r.Body)
		if len(data) > 0 {
			_ = json.Unmarshal(data, &rec.Body)
		}
	}

	s.mu.Lock()
	s.requests = append(s.requests, rec)
	var canned *Response
	if q := s.queued[r.URL.Path]; len(q) > 0 {
		canned = &q[0]
		s.queued[r.URL.Path] = q[1:]
	}
	s.mu.Unlock()

	if r.Header.Get("Authorization") != "Bearer "+APIKey {
		writeJSON(w, http.StatusUnauthorized, nil, map[string]any{"code": 1002, "message": "Authentication required", "suggestion": "Check if your credentials is valid"})
		return
	}
	if canned != nil {
		writeJSON(w, canned.Status, canned.Header, canned.Body)
		return
	}

	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tasks/"):
		s.serveTask(w, strings.TrimPrefix(r.URL.Path, "/tasks/"))
	case r.Method == http.MethodPost && r.URL.Path == "/tasks/list":
		s.serveTaskList(w, rec.Body)
	case r.Method == http.MethodPost && r.URL.Path == "/files":
		ok(w, map[string]any{"file_token": s.newID("file")})
	case r.Method == http.MethodPost && r.URL.Path == "/files/presign":
		id := s.newID("file")
		ok(w, map[string]any{"presigned_url": s.URL + "/presigned/" + id, "file_token": id, "expires_in": 1800})
	case r.Method == http.MethodGet && r.URL.Path == "/account/balance":
		ok(w, s.Balance)
	case r.Method == http.MethodGet && r.URL.Path == "/account/usage":
		ok(w, s.Usage)
	case r.Method == http.MethodPost && isTaskCreating(r.URL.Path):
		id := s.newID("task")
		s.SetTask(id, map[string]any{"status": "queued", "type": taskType(r.URL.Path)})
		ok(w, map[string]any{"task_id": id})
	default:
		writeJSON(w, http.StatusNotFound, nil, map[string]any{"code": 4001, "message": "endpoint does not exist: " + r.Method + " " + r.URL.Path})
	}
}

func isTaskCreating(path string) bool {
	for _, prefix := range []string{"/generation/", "/models/", "/mesh/", "/animations/"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func taskType(path string) string {
	return strings.ReplaceAll(path[strings.LastIndex(path, "/")+1:], "-", "_")
}

func (s *Server) newID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	return fmt.Sprintf("%s-%04d", prefix, s.nextID)
}

func (s *Server) nextState(id string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	states, found := s.tasks[id]
	if !found {
		return nil, false
	}
	state := states[0]
	if len(states) > 1 {
		s.tasks[id] = states[1:]
	}
	return state, true
}

func (s *Server) serveTask(w http.ResponseWriter, id string) {
	state, found := s.nextState(id)
	if !found {
		writeJSON(w, http.StatusNotFound, nil, map[string]any{"code": 2001, "message": "task not found"})
		return
	}
	ok(w, state)
}

func (s *Server) serveTaskList(w http.ResponseWriter, body map[string]any) {
	ids, _ := body["task_ids"].([]any)
	tasks := map[string]any{}
	missed := []string{}
	for _, raw := range ids {
		id, _ := raw.(string)
		if state, found := s.nextState(id); found {
			tasks[id] = state
		} else {
			missed = append(missed, id)
		}
	}
	ok(w, map[string]any{"tasks": tasks, "missed": missed})
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	f, found := s.files[r.URL.Path]
	s.mu.Unlock()
	if !found {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", f.contentType)
	if !f.truncated {
		_, _ = w.Write(f.data)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(f.data)))
	_, _ = w.Write(f.data[:len(f.data)/2])
	w.(http.Flusher).Flush()
	if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
		_ = conn.Close()
	}
}

func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, nil, map[string]any{"code": 0, "data": data})
}

func writeJSON(w http.ResponseWriter, status int, header http.Header, body any) {
	for k, vs := range header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
