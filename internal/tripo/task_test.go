package tripo_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mordor-forge/trident-mcp/internal/tripo"
	"github.com/mordor-forge/trident-mcp/internal/tripo/tripotest"
)

func TestGetTask_DecodesV3Task(t *testing.T) {
	fake := tripotest.New(t)
	fake.SetTask("t1", map[string]any{
		"status": "failed", "type": "text_to_model", "error_code": 2008, "error_message": "content policy",
		"credits_consumed": 0, "created_at": 1760000000, "running_left_time": 12,
	})
	task, err := fake.Client(t).GetTask(context.Background(), "t1")
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != tripo.StatusFailed || !task.Status.Terminal() || task.ErrorMessage != "content policy" ||
		task.ErrorCode != "2008" || task.CreatedAt != "1760000000" || *task.RunningLeftTime != 12 {
		t.Fatalf("task = %+v", task)
	}
}

func TestTask_OutputAccessors(t *testing.T) {
	fake := tripotest.New(t)
	fake.Succeed("mv", map[string]any{
		"generate_multiview_image": map[string]any{"front_view_url": "https://cdn/f.png", "back_view_url": "https://cdn/b.png"},
		"rendered_image_url":       "https://cdn/r.webp",
		"not_a_url":                "x",
	})
	fake.Succeed("rc", map[string]any{"riggable": true, "topology": "quadruped"})
	c := fake.Client(t)

	mv, err := c.GetTask(context.Background(), "mv")
	if err != nil {
		t.Fatal(err)
	}
	urls := mv.OutputURLs()
	if len(urls) != 3 || urls["front_view"] != "https://cdn/f.png" || urls["rendered_image"] != "https://cdn/r.webp" {
		t.Fatalf("urls = %v", urls)
	}

	rc, err := c.GetTask(context.Background(), "rc")
	if err != nil {
		t.Fatal(err)
	}
	if riggable, ok := rc.Riggable(); !ok || !riggable || rc.RigType() != "quadruped" {
		t.Fatalf("rig check = %v %v %q", riggable, ok, rc.RigType())
	}
}

func TestGetTask_RejectsUnsafeIDs(t *testing.T) {
	fake := tripotest.New(t)
	if _, err := fake.Client(t).GetTask(context.Background(), "../account/balance"); err == nil {
		t.Fatal("expected an error")
	}
	if len(fake.Requests()) != 0 {
		t.Fatal("unsafe ID reached Tripo")
	}
}

func TestGetTasks_ReturnsMapAndMissed(t *testing.T) {
	fake := tripotest.New(t)
	fake.Succeed("a", map[string]any{})
	list, err := fake.Client(t).GetTasks(context.Background(), []string{"a", "zz"})
	if err != nil {
		t.Fatal(err)
	}
	if list.Tasks["a"].Status != tripo.StatusSuccess || len(list.Missed) != 1 || list.Missed[0] != "zz" {
		t.Fatalf("list = %+v", list)
	}
}

func TestWaitTask_PollsUntilTerminal(t *testing.T) {
	fake := tripotest.New(t)
	fake.SetTask("w",
		map[string]any{"status": "queued", "queuing_num": 3},
		map[string]any{"status": "running", "progress": 40},
		map[string]any{"status": "success", "output": map[string]any{}},
	)
	var seen []tripo.Status
	task, err := fake.Client(t).WaitTask(context.Background(), "w", tripo.WaitOptions{
		Interval: time.Millisecond,
		OnUpdate: func(t *tripo.Task) { seen = append(seen, t.Status) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != tripo.StatusSuccess || len(seen) != 3 {
		t.Fatalf("task = %+v, seen = %v", task, seen)
	}
}

func TestWaitTask_RidesOutTransientPollErrors(t *testing.T) {
	fake := tripotest.New(t)
	fake.SetTask("w", map[string]any{"status": "success", "output": map[string]any{}})
	bad := tripotest.Response{Status: http.StatusBadGateway, Body: "<html>502</html>"}
	fake.Queue("/tasks/w", bad)
	fake.Queue("/tasks/w", bad)
	task, err := fake.Client(t).WaitTask(context.Background(), "w", tripo.WaitOptions{Interval: time.Millisecond})
	if err != nil || task.Status != tripo.StatusSuccess {
		t.Fatalf("task = %+v, err = %v", task, err)
	}

	fake.Queue("/tasks/w", tripotest.Response{Status: http.StatusNotFound, Body: map[string]any{"code": 2001, "message": "task not found"}})
	if _, err := fake.Client(t).WaitTask(context.Background(), "w", tripo.WaitOptions{Interval: time.Millisecond}); err == nil {
		t.Fatal("a 4xx poll error should end the wait")
	}
}

func TestWaitTask_ReturnsLastTaskOnTimeout(t *testing.T) {
	fake := tripotest.New(t)
	fake.SetTask("slow", map[string]any{"status": "running", "progress": 10})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	task, err := fake.Client(t).WaitTask(ctx, "slow", tripo.WaitOptions{Interval: 5 * time.Millisecond})
	if !errors.Is(err, context.DeadlineExceeded) || task == nil || task.Progress != 10 {
		t.Fatalf("task = %+v, err = %v", task, err)
	}
}

func TestDownload_SavesEveryOutput(t *testing.T) {
	fake := tripotest.New(t)
	fake.Succeed("d", map[string]any{
		"model_url":          fake.AddFile("pbr_model.glb", "model/gltf-binary", []byte("glTF")),
		"rendered_image_url": fake.AddFile("preview", "image/webp", []byte("RIFF")),
		"splat_url":          fake.AddFile("scene.splat", "application/octet-stream", []byte("SPLAT")),
	})
	dir := t.TempDir()

	task, artifacts, err := fake.Client(t).Download(context.Background(), "d", dir)
	if err != nil {
		t.Fatal(err)
	}
	if task.TaskID != "d" || len(artifacts) != 3 {
		t.Fatalf("artifacts = %+v", artifacts)
	}
	want := map[string]string{"model": "glb", "rendered_image": "webp", "splat": "splat"}
	for _, a := range artifacts {
		if want[a.Name] != a.Format || a.Path != filepath.Join(dir, "d", a.Name+"."+a.Format) {
			t.Errorf("artifact %+v", a)
		}
		if _, err := os.Stat(a.Path); err != nil {
			t.Errorf("artifact file: %v", err)
		}
	}

	_, only, err := fake.Client(t).Download(context.Background(), "d", dir, "model")
	if err != nil || len(only) != 1 || only[0].Name != "model" {
		t.Fatalf("filtered download = %+v, %v", only, err)
	}
	if _, _, err := fake.Client(t).Download(context.Background(), "d", dir, "depth_image"); err == nil {
		t.Fatal("expected an error for a missing output")
	}
}

func TestDownload_RefusesUnfinishedTasks(t *testing.T) {
	fake := tripotest.New(t)
	fake.SetTask("f", map[string]any{"status": "failed", "error_message": "model too complex"})
	_, _, err := fake.Client(t).Download(context.Background(), "f", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "model too complex") {
		t.Fatalf("err = %v", err)
	}
}

func TestDownload_LeavesNoPartialFiles(t *testing.T) {
	for name, url := range map[string]func(*tripotest.Server) string{
		"expired URL": func(f *tripotest.Server) string { return f.URL + "/cdn/expired.glb" },
		"dropped mid-body": func(f *tripotest.Server) string {
			return f.AddTruncatedFile("big.glb", make([]byte, 1<<20))
		},
	} {
		t.Run(name, func(t *testing.T) {
			fake := tripotest.New(t)
			fake.Succeed("x", map[string]any{"model_url": url(fake)})
			dir := t.TempDir()
			if _, _, err := fake.Client(t).Download(context.Background(), "x", dir); err == nil {
				t.Fatal("expected an error")
			}
			entries, _ := os.ReadDir(filepath.Join(dir, "x"))
			if len(entries) != 0 {
				t.Fatalf("left behind: %v", entries)
			}
		})
	}
}

func TestDownload_FilesAreReadableByOthers(t *testing.T) {
	fake := tripotest.New(t)
	fake.Succeed("p", map[string]any{"model_url": fake.AddFile("m.glb", "model/gltf-binary", []byte("glTF"))})
	_, artifacts, err := fake.Client(t).Download(context.Background(), "p", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(artifacts[0].Path)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("mode = %v, %v", info.Mode(), err)
	}
}

func TestClient_ErrorsCarryTripoEnvelope(t *testing.T) {
	fake := tripotest.New(t)
	c := fake.Client(t)

	fake.Queue("/account/balance", tripotest.Response{Status: http.StatusForbidden, Body: map[string]any{
		"code": 2010, "message": "insufficient credits", "suggestion": "top up", "request_id": "req_1",
	}})
	_, err := c.GetBalance(context.Background())
	var apiErr *tripo.Error
	if !errors.As(err, &apiErr) || apiErr.Code != 2010 || apiErr.HTTPStatus != 403 || apiErr.RequestID != "req_1" {
		t.Fatalf("err = %#v", err)
	}
	for _, part := range []string{"insufficient credits", "code 2010", "request req_1", "Suggestion: top up"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error %q lacks %q", err, part)
		}
	}

	fake.Queue("/account/balance", tripotest.Response{Status: http.StatusOK, Body: map[string]any{"code": 1004, "message": "bad parameter"}})
	if _, err := c.GetBalance(context.Background()); err == nil || !strings.Contains(err.Error(), "bad parameter") {
		t.Fatalf("HTTP 200 error envelope: %v", err)
	}

	fake.Queue("/account/balance", tripotest.Response{Status: http.StatusBadRequest, Body: "<html>400 Bad Request</html>"})
	if _, err := c.GetBalance(context.Background()); err == nil || !strings.Contains(err.Error(), "400 Bad Request") {
		t.Fatalf("non-JSON error body: %v", err)
	}
}

func TestClient_RetriesRateLimits(t *testing.T) {
	fake := tripotest.New(t)
	limited := tripotest.Response{Status: http.StatusTooManyRequests, Header: http.Header{"Retry-After": {"0"}},
		Body: map[string]any{"code": 2000, "message": "concurrency limit"}}
	fake.Queue("/generation/text-to-image", limited)
	fake.Queue("/generation/text-to-image", limited)

	submit(t, fake, tripo.TextToImage, tripo.TextToImageRequest{Prompt: "x"})
	if n := len(fake.Requests()); n != 3 {
		t.Fatalf("requests = %d, want 2 rate-limited + 1 accepted", n)
	}
}

func TestClient_RejectsBadKey(t *testing.T) {
	fake := tripotest.New(t)
	c, err := tripo.New(tripo.Options{APIKey: "tsk_wrong", BaseURL: fake.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetBalance(context.Background()); err == nil || !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("err = %v", err)
	}
	if _, err := tripo.New(tripo.Options{}); err == nil {
		t.Fatal("expected an error without an API key")
	}
}
