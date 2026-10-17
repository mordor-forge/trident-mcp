package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/tripo/tripotest"
)

type harness struct {
	fake      *tripotest.Server
	session   *mcp.ClientSession
	outputDir string

	mu       sync.Mutex
	progress []*mcp.ProgressNotificationParams
}

// connect starts the real server, backed by the real Tripo client talking to
// the fake API, and an in-memory MCP client.
func connect(t *testing.T) *harness {
	t.Helper()
	h := &harness{fake: tripotest.New(t), outputDir: t.TempDir()}
	srv := New(h.fake.Client(t), Options{OutputDir: h.outputDir, Version: "test", PollInterval: 10 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	go func() { _ = srv.MCPServer().Run(ctx, serverTransport) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, &mcp.ClientOptions{
		ProgressNotificationHandler: func(_ context.Context, req *mcp.ProgressNotificationClientRequest) {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.progress = append(h.progress, req.Params)
		},
	})
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	h.session = session
	return h
}

// call invokes a tool and returns its text and structured content.
func (h *harness) call(t *testing.T, name string, args map[string]any) (string, map[string]any, bool) {
	t.Helper()
	params := &mcp.CallToolParams{Name: name, Arguments: args}
	params.SetProgressToken("progress-" + name)
	res, err := h.session.CallTool(context.Background(), params)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	var structured map[string]any
	if res.StructuredContent != nil {
		data, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(data, &structured)
	}
	return b.String(), structured, res.IsError
}

func (h *harness) mustCall(t *testing.T, name string, args map[string]any) (string, map[string]any) {
	t.Helper()
	text, structured, isErr := h.call(t, name, args)
	if isErr {
		t.Fatalf("%s returned an error: %s", name, text)
	}
	return text, structured
}

var operationTools = []string{
	"text_to_model", "image_to_model", "multiview_to_model",
	"text_to_image", "image_to_image", "image_to_multiview", "edit_multiview", "image_to_splat",
	"import_model", "texture_model", "convert_model",
	"retopologize", "smart_uv", "segment_mesh", "complete_mesh", "smart_segment",
	"rig_check", "rig_model", "retarget_animation",
}

var otherTools = []string{
	"wait_for_task", "get_task", "get_tasks", "download_task",
	"get_balance", "get_usage", "upload_file", "list_models", "get_config",
}

func TestServer_ListsTheWholeSurface(t *testing.T) {
	h := connect(t)
	res, err := h.session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tool := range res.Tools {
		got = append(got, tool.Name)
		if tool.Description == "" || tool.Annotations == nil {
			t.Errorf("%s lacks a description or annotations", tool.Name)
		}
	}
	for _, tool := range res.Tools {
		spends := slices.Contains(operationTools, tool.Name) && tool.Name != "rig_check"
		destructive := tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint
		if !tool.Annotations.ReadOnlyHint && spends != destructive {
			t.Errorf("%s: destructiveHint %v, but credit-spending is %v", tool.Name, destructive, spends)
		}
	}
	want := append(append([]string(nil), operationTools...), otherTools...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tools = %v\nwant    %v", got, want)
	}
}

func TestDocs_DescribeEveryTool(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join("..", "..", "docs", "MCP_TOOLS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append(append([]string(nil), operationTools...), otherTools...) {
		if !strings.Contains(string(doc), "`"+name+"`") {
			t.Errorf("docs/MCP_TOOLS.md does not mention `%s`", name)
		}
	}
}

func TestOperationTools_ReachTripo(t *testing.T) {
	h := connect(t)
	image := filepath.Join(t.TempDir(), "ref.png")
	if err := os.WriteFile(image, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		tool string
		args map[string]any
		path string
		want map[string]any
	}{
		{"text_to_model", map[string]any{"prompt": "a lantern", "model": "p1", "face_limit": 3000},
			"/generation/text-to-model", map[string]any{"prompt": "a lantern", "model": "P1-20260311", "face_limit": 3000.0}},
		{"image_to_model", map[string]any{"input": image, "orientation": "align_image"},
			"/generation/image-to-model", map[string]any{"orientation": "align_image", "model": "v3.1-20260211"}},
		{"multiview_to_model", map[string]any{"inputs": map[string]any{"front": "https://x/f.png", "left": "https://x/l.png"}},
			"/generation/multiview-to-model", map[string]any{"inputs": []any{map[string]any{"front": "https://x/f.png"}, map[string]any{"left": "https://x/l.png"}}}},
		{"text_to_image", map[string]any{"prompt": "a sword", "template": "t_pose"},
			"/generation/text-to-image", map[string]any{"model": "seedream_v4", "template": "t_pose"}},
		{"image_to_image", map[string]any{"input": "task-img", "template": "3d_enhance"},
			"/generation/image-to-image", map[string]any{"input": "task-img", "template": "3d_enhance"}},
		{"image_to_multiview", map[string]any{"input": "task-img"}, "/generation/image-to-multiview", map[string]any{"input": "task-img"}},
		{"edit_multiview", map[string]any{"input": "task-mv", "prompts": []any{map[string]any{"view": "left", "prompt": "fix the arm"}}},
			"/generation/edit-multiview", map[string]any{"input": "task-mv"}},
		{"image_to_splat", map[string]any{"input": "https://x/a.png"}, "/generation/image-to-splat", map[string]any{"input": "https://x/a.png"}},
		{"import_model", map[string]any{"input": "https://x/a.glb"}, "/models/import", map[string]any{"input": "https://x/a.glb"}},
		{"texture_model", map[string]any{"input": "task-mesh", "texture_prompt": map[string]any{"text": "rusty iron"}},
			"/models/texture", map[string]any{"model": "v3.0-20250812", "texture_prompt": map[string]any{"text": "rusty iron"}}},
		{"convert_model", map[string]any{"input": "task-mesh", "format": "STL", "flatten_bottom": true},
			"/models/convert", map[string]any{"format": "STL", "flatten_bottom": true}},
		{"retopologize", map[string]any{"input": "task-mesh", "quad": true, "face_limit": 5000},
			"/mesh/decimate", map[string]any{"model": "v2.0", "quad": true, "face_limit": 5000.0}},
		{"smart_uv", map[string]any{"input": "task-mesh"}, "/mesh/uv", map[string]any{"input": "task-mesh"}},
		{"segment_mesh", map[string]any{"input": "task-mesh"}, "/mesh/segment", map[string]any{"model": "v1.0-20250506"}},
		{"complete_mesh", map[string]any{"input": "task-seg"}, "/mesh/complete", map[string]any{"input": "task-seg"}},
		{"smart_segment", map[string]any{"seg_type": "image", "input": "https://x/a.png"}, "/mesh/smartsegment", map[string]any{"seg_type": "image"}},
		{"rig_check", map[string]any{"input": "task-mesh"}, "/animations/rig-check", map[string]any{"input": "task-mesh"}},
		{"rig_model", map[string]any{"input": "task-mesh", "model": "v2.5", "rig_type": "quadruped"},
			"/animations/rig", map[string]any{"model": "v2.5-20260210", "rig_type": "quadruped"}},
		{"retarget_animation", map[string]any{"input": "task-rig", "animation": "preset:walk"},
			"/animations/retarget", map[string]any{"animation": "preset:walk"}},
	}
	covered := map[string]bool{}
	for _, tc := range cases {
		covered[tc.tool] = true
		t.Run(tc.tool, func(t *testing.T) {
			text, structured := h.mustCall(t, tc.tool, tc.args)
			taskID, _ := structured["task_id"].(string)
			if taskID == "" || !strings.Contains(text, "wait_for_task") {
				t.Fatalf("result = %q %v", text, structured)
			}
			body := h.fake.Last(t, tc.path).Body
			for k, v := range tc.want {
				if !reflect.DeepEqual(body[k], v) {
					t.Errorf("body[%q] = %#v, want %#v", k, body[k], v)
				}
			}
		})
	}
	for _, tool := range operationTools {
		if !covered[tool] {
			t.Errorf("no test case calls %s", tool)
		}
	}
}

func TestOperationTools_SchemaRejectsBadArguments(t *testing.T) {
	h := connect(t)
	for name, args := range map[string]map[string]any{
		"old camelCase argument": {"prompt": "x", "faceLimit": 1000},
		"enum outside the list":  {"prompt": "x", "texture_quality": "ultra"},
		"wrong type":             {"prompt": "x", "face_limit": "many"},
		"missing prompt":         {},
	} {
		t.Run(name, func(t *testing.T) {
			if text, _, isErr := h.call(t, "text_to_model", args); !isErr {
				t.Fatalf("accepted: %s", text)
			}
		})
	}
	if reqs := h.fake.Requests(); len(reqs) != 0 {
		t.Fatalf("bad arguments reached Tripo: %v", reqs)
	}
}

func TestOperationTools_SurfaceTripoErrors(t *testing.T) {
	h := connect(t)
	h.fake.Queue("/generation/text-to-model", tripotest.Response{Status: 403, Body: map[string]any{
		"code": 2010, "message": "Insufficient credits", "suggestion": "Top up at platform.tripo3d.ai",
	}})
	text, _, isErr := h.call(t, "text_to_model", map[string]any{"prompt": "x"})
	if !isErr || !strings.Contains(text, "Insufficient credits") || !strings.Contains(text, "Top up") {
		t.Fatalf("error = %q", text)
	}
}

func TestWaitForTask_ReportsProgressAndSavesOutputs(t *testing.T) {
	h := connect(t)
	model := h.fake.AddFile("pbr_model.glb", "model/gltf-binary", []byte("glTF-binary"))
	preview := h.fake.AddFile("render.webp", "image/webp", []byte("RIFF"))
	h.fake.SetTask("t1",
		map[string]any{"status": "queued", "queuing_num": 2},
		map[string]any{"status": "queued", "queuing_num": 1},
		map[string]any{"status": "running", "progress": 40},
		map[string]any{"status": "success", "type": "text_to_model", "credits_consumed": 40,
			"output": map[string]any{"model_url": model, "rendered_image_url": preview}},
	)

	text, structured := h.mustCall(t, "wait_for_task", map[string]any{"task_id": "t1"})
	if structured["done"] != true || !strings.Contains(text, "Saved 2 file(s)") {
		t.Fatalf("wait = %q %v", text, structured)
	}
	saved := filepath.Join(h.outputDir, "t1", "model.glb")
	if data, err := os.ReadFile(saved); err != nil || string(data) != "glTF-binary" {
		t.Fatalf("saved model: %q, %v", data, err)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.progress) != 3 || !strings.Contains(h.progress[0].Message, "queued (position 2)") {
		t.Fatalf("progress = %+v", h.progress)
	}
	for i := 1; i < len(h.progress); i++ {
		if h.progress[i].Progress <= h.progress[i-1].Progress {
			t.Fatalf("progress must increase: %v then %v", h.progress[i-1].Progress, h.progress[i].Progress)
		}
	}
}

func TestWaitForTask_ReturnsWhileStillRunning(t *testing.T) {
	h := connect(t)
	h.fake.SetTask("slow", map[string]any{"status": "running", "progress": 30})
	text, structured := h.mustCall(t, "wait_for_task", map[string]any{"task_id": "slow", "timeout_seconds": 1})
	if structured["done"] != false || !strings.Contains(text, "call wait_for_task again") {
		t.Fatalf("wait = %q %v", text, structured)
	}
}

func TestWaitForTask_ReportsFailureWithoutDownloading(t *testing.T) {
	h := connect(t)
	h.fake.SetTask("bad", map[string]any{"status": "banned", "error_code": 2008, "error_message": "content policy"})
	text, structured := h.mustCall(t, "wait_for_task", map[string]any{"task_id": "bad"})
	if structured["done"] != true || !strings.Contains(text, "content policy") || !strings.Contains(text, "code 2008") {
		t.Fatalf("wait = %q %v", text, structured)
	}
}

func TestTaskTools(t *testing.T) {
	h := connect(t)
	h.fake.Succeed("rc", map[string]any{"riggable": true, "rig_type": "biped"})
	h.fake.Succeed("img", map[string]any{"generated_image_url": h.fake.AddFile("img.png", "image/png", []byte("PNG"))})

	text, _ := h.mustCall(t, "get_task", map[string]any{"task_id": "rc"})
	if !strings.Contains(text, "Riggable: true (rig_type biped)") {
		t.Fatalf("get_task = %q", text)
	}

	text, structured := h.mustCall(t, "get_tasks", map[string]any{"task_ids": []any{"rc", "nope"}})
	if !strings.Contains(text, "Not found: nope") || structured["tasks"] == nil {
		t.Fatalf("get_tasks = %q %v", text, structured)
	}

	text, _ = h.mustCall(t, "download_task", map[string]any{"task_id": "img"})
	if !strings.Contains(text, filepath.Join(h.outputDir, "img", "generated_image.png")) {
		t.Fatalf("download_task = %q", text)
	}
}

func TestAccountAndConfigTools(t *testing.T) {
	h := connect(t)

	text, _ := h.mustCall(t, "get_balance", nil)
	if !strings.Contains(text, "Available credits: 1000") {
		t.Fatalf("get_balance = %q", text)
	}

	h.fake.Usage = []any{map[string]any{"task_id": "t1", "type": "rig", "credits_consumed": 25}}
	text, _ = h.mustCall(t, "get_usage", map[string]any{"limit": 5})
	if !strings.Contains(text, "1 task(s), 25 credits") {
		t.Fatalf("get_usage = %q", text)
	}

	path := filepath.Join(t.TempDir(), "mesh.glb")
	if err := os.WriteFile(path, []byte("glTF"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, structured := h.mustCall(t, "upload_file", map[string]any{"path": path})
	if token, _ := structured["file_token"].(string); !strings.HasPrefix(token, "file-") {
		t.Fatalf("upload_file = %v", structured)
	}

	text, structured = h.mustCall(t, "list_models", map[string]any{"operation": "rig_model"})
	if models, _ := structured["models"].([]any); len(models) != 3 || !strings.Contains(text, "v3.0-20260909") {
		t.Fatalf("list_models = %q", text)
	}
	text, _ = h.mustCall(t, "list_models", map[string]any{"operation": "text_to_image"})
	if !strings.Contains(text, "seedream_v4 (Seedream v4) default") || strings.Contains(text, "seedream_v5 (Seedream v5) default") {
		t.Fatalf("text_to_image defaults = %q", text)
	}
	if _, _, isErr := h.call(t, "list_models", map[string]any{"operation": "import_model"}); !isErr {
		t.Fatal("list_models for a tool without model versions should fail")
	}

	text, _ = h.mustCall(t, "get_config", nil)
	if !strings.Contains(text, h.fake.URL) || !strings.Contains(text, h.outputDir) {
		t.Fatalf("get_config = %q", text)
	}
}
