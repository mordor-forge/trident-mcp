package server

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/provider"
)

// --- Mock providers ---

type mockGenerator struct {
	textResult           *provider.ModelOperation
	textErr              error
	imageResult          *provider.ModelOperation
	imageErr             error
	multiviewResult      *provider.ModelOperation
	multiviewErr         error
	textImageResult      *provider.ModelOperation
	textImageErr         error
	imageImageResult     *provider.ModelOperation
	imageImageErr        error
	imageMultiviewResult *provider.ModelOperation
	imageMultiviewErr    error
	editMultiviewResult  *provider.ModelOperation
	editMultiviewErr     error
	imageSplatResult     *provider.ModelOperation
	imageSplatErr        error
}

func (m *mockGenerator) TextToModel(_ context.Context, _ provider.TextToModelRequest) (*provider.ModelOperation, error) {
	return m.textResult, m.textErr
}

func (m *mockGenerator) ImageToModel(_ context.Context, _ provider.ImageToModelRequest) (*provider.ModelOperation, error) {
	return m.imageResult, m.imageErr
}

func (m *mockGenerator) MultiviewToModel(_ context.Context, _ provider.MultiviewToModelRequest) (*provider.ModelOperation, error) {
	return m.multiviewResult, m.multiviewErr
}

func (m *mockGenerator) TextToImage(_ context.Context, _ provider.TextToImageRequest) (*provider.ModelOperation, error) {
	return m.textImageResult, m.textImageErr
}

func (m *mockGenerator) ImageToImage(_ context.Context, _ provider.ImageToImageRequest) (*provider.ModelOperation, error) {
	return m.imageImageResult, m.imageImageErr
}

func (m *mockGenerator) ImageToMultiview(_ context.Context, _ provider.ImageToMultiviewRequest) (*provider.ModelOperation, error) {
	return m.imageMultiviewResult, m.imageMultiviewErr
}

func (m *mockGenerator) EditMultiview(_ context.Context, _ provider.EditMultiviewRequest) (*provider.ModelOperation, error) {
	return m.editMultiviewResult, m.editMultiviewErr
}

func (m *mockGenerator) ImageToSplat(_ context.Context, _ provider.ImageToSplatRequest) (*provider.ModelOperation, error) {
	return m.imageSplatResult, m.imageSplatErr
}

type mockStatus struct {
	statusResult   *provider.ModelTaskStatus
	statusErr      error
	downloadResult *provider.ModelResult
	downloadErr    error
	batchResult    *provider.BatchTasksResult
	batchErr       error
	uploadResult   *provider.FileUpload
	uploadErr      error
	presignResult  *provider.FileUpload
	presignErr     error
	balanceResult  *provider.AccountBalance
	balanceErr     error
	usageResult    *provider.AccountUsageResult
	usageErr       error
}

func (m *mockStatus) Status(_ context.Context, _ string) (*provider.ModelTaskStatus, error) {
	return m.statusResult, m.statusErr
}

func (m *mockStatus) Download(_ context.Context, _ string, _ string) (*provider.ModelResult, error) {
	return m.downloadResult, m.downloadErr
}

func (m *mockStatus) BatchTasks(_ context.Context, _ []string) (*provider.BatchTasksResult, error) {
	return m.batchResult, m.batchErr
}

func (m *mockStatus) UploadFile(_ context.Context, _ provider.UploadFileRequest) (*provider.FileUpload, error) {
	return m.uploadResult, m.uploadErr
}

func (m *mockStatus) CreateFileUpload(_ context.Context, _ provider.CreateFileUploadRequest) (*provider.FileUpload, error) {
	return m.presignResult, m.presignErr
}

func (m *mockStatus) GetBalance(_ context.Context) (*provider.AccountBalance, error) {
	return m.balanceResult, m.balanceErr
}

func (m *mockStatus) GetUsage(_ context.Context) (*provider.AccountUsageResult, error) {
	return m.usageResult, m.usageErr
}

type mockCommon struct {
	batchResult   *provider.BatchTasksResult
	batchErr      error
	uploadResult  *provider.FileUpload
	uploadErr     error
	presignResult *provider.FileUpload
	presignErr    error
	balanceResult *provider.AccountBalance
	balanceErr    error
	usageResult   *provider.AccountUsageResult
	usageErr      error
}

func (m *mockCommon) BatchTasks(_ context.Context, _ []string) (*provider.BatchTasksResult, error) {
	return m.batchResult, m.batchErr
}

func (m *mockCommon) UploadFile(_ context.Context, _ provider.UploadFileRequest) (*provider.FileUpload, error) {
	return m.uploadResult, m.uploadErr
}

func (m *mockCommon) CreateFileUpload(_ context.Context, _ provider.CreateFileUploadRequest) (*provider.FileUpload, error) {
	return m.presignResult, m.presignErr
}

func (m *mockCommon) GetBalance(_ context.Context) (*provider.AccountBalance, error) {
	return m.balanceResult, m.balanceErr
}

func (m *mockCommon) GetUsage(_ context.Context) (*provider.AccountUsageResult, error) {
	return m.usageResult, m.usageErr
}

type mockGeneratorWithCommon struct {
	mockGenerator
	mockCommon
}

type mockPostProc struct {
	retopResult    *provider.ModelOperation
	retopErr       error
	convertResult  *provider.ModelOperation
	convertErr     error
	stylizeResult  *provider.ModelOperation
	stylizeErr     error
	importResult   *provider.ModelOperation
	importErr      error
	refineResult   *provider.ModelOperation
	refineErr      error
	textureResult  *provider.ModelOperation
	textureErr     error
	segmentResult  *provider.ModelOperation
	segmentErr     error
	completeResult *provider.ModelOperation
	completeErr    error
	rigCheckResult *provider.RigCheckResult
	rigCheckErr    error
	rigResult      *provider.ModelOperation
	rigErr         error
	retargetResult *provider.ModelOperation
	retargetErr    error
}

func (m *mockPostProc) Retopologize(_ context.Context, _ provider.RetopologyRequest) (*provider.ModelOperation, error) {
	return m.retopResult, m.retopErr
}

func (m *mockPostProc) ConvertFormat(_ context.Context, _ provider.ConvertRequest) (*provider.ModelOperation, error) {
	return m.convertResult, m.convertErr
}

func (m *mockPostProc) Stylize(_ context.Context, _ provider.StylizeRequest) (*provider.ModelOperation, error) {
	return m.stylizeResult, m.stylizeErr
}

func (m *mockPostProc) ImportModel(_ context.Context, _ provider.ImportModelRequest) (*provider.ModelOperation, error) {
	return m.importResult, m.importErr
}

func (m *mockPostProc) RefineModel(_ context.Context, _ provider.RefineModelRequest) (*provider.ModelOperation, error) {
	return m.refineResult, m.refineErr
}

func (m *mockPostProc) TextureModel(_ context.Context, _ provider.TextureModelRequest) (*provider.ModelOperation, error) {
	return m.textureResult, m.textureErr
}

func (m *mockPostProc) SegmentMesh(_ context.Context, _ provider.SegmentMeshRequest) (*provider.ModelOperation, error) {
	return m.segmentResult, m.segmentErr
}

func (m *mockPostProc) CompleteMesh(_ context.Context, _ provider.CompleteMeshRequest) (*provider.ModelOperation, error) {
	return m.completeResult, m.completeErr
}

func (m *mockPostProc) RigCheck(_ context.Context, _ provider.RigCheckRequest) (*provider.RigCheckResult, error) {
	return m.rigCheckResult, m.rigCheckErr
}

func (m *mockPostProc) RigModel(_ context.Context, _ provider.RigModelRequest) (*provider.ModelOperation, error) {
	return m.rigResult, m.rigErr
}

func (m *mockPostProc) RetargetAnimation(_ context.Context, _ provider.RetargetAnimationRequest) (*provider.ModelOperation, error) {
	return m.retargetResult, m.retargetErr
}

type mockModelLister struct {
	models []provider.ModelInfo
	err    error
}

func (m *mockModelLister) ListModels(_ context.Context) ([]provider.ModelInfo, error) {
	return m.models, m.err
}

// --- Test client helpers ---

func connectTestClient(t *testing.T, gen *mockGenerator, stat *mockStatus, pp *mockPostProc, lister *mockModelLister) *mcp.ClientSession {
	t.Helper()

	var g provider.ModelGenerator
	var s provider.ModelStatus
	var p provider.ModelPostProcessor
	var l provider.ModelLister

	if gen != nil {
		g = gen
	}
	if stat != nil {
		s = stat
	}
	if pp != nil {
		p = pp
	}
	if lister != nil {
		l = lister
	}

	srv := NewWithOptions(g, s, p, l, Options{
		Backend:   "tripo",
		OutputDir: t.TempDir(),
		Version:   "test-version",
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	return session
}

// --- Constructor tests ---

func TestNew_CreatesServer(t *testing.T) {
	srv := New(&mockGenerator{}, &mockStatus{}, &mockPostProc{}, nil, t.TempDir())
	if srv == nil {
		t.Fatal("New returned nil")
		return
	}
	if srv.mcp == nil {
		t.Fatal("underlying MCP server is nil")
	}
}

func TestNew_NilProviders(t *testing.T) {
	srv := New(nil, nil, nil, nil, t.TempDir())
	if srv == nil {
		t.Fatal("New returned nil with all nil providers")
	}
}

func TestMCPServer_ReturnsUnderlyingServer(t *testing.T) {
	srv := New(&mockGenerator{}, nil, nil, nil, t.TempDir())
	if srv.MCPServer() != srv.mcp {
		t.Fatal("MCPServer() did not return the underlying server")
	}
}

// --- Tool registration tests ---

func TestToolsRegistered_AllProviders(t *testing.T) {
	session := connectTestClient(t,
		&mockGenerator{textResult: &provider.ModelOperation{}},
		&mockStatus{statusResult: &provider.ModelTaskStatus{}},
		&mockPostProc{retopResult: &provider.ModelOperation{}},
		&mockModelLister{models: []provider.ModelInfo{}},
	)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	wantTools := map[string]bool{
		"text_to_3d":         false,
		"image_to_3d":        false,
		"multiview_to_3d":    false,
		"text_to_image":      false,
		"image_to_image":     false,
		"image_to_multiview": false,
		"edit_multiview":     false,
		"image_to_splat":     false,
		"task_status":        false,
		"get_task":           false,
		"get_tasks":          false,
		"get_balance":        false,
		"get_usage":          false,
		"download_model":     false,
		"upload_file":        false,
		"create_file_upload": false,
		"retopologize":       false,
		"convert_format":     false,
		"stylize":            false,
		"import_model":       false,
		"refine_model":       false,
		"texture_model":      false,
		"segment_mesh":       false,
		"complete_mesh":      false,
		"rig_check":          false,
		"rig_model":          false,
		"retarget_animation": false,
		"list_models":        false,
		"get_config":         false,
	}

	for _, tool := range result.Tools {
		if _, ok := wantTools[tool.Name]; ok {
			wantTools[tool.Name] = true
		}
	}
	for name, found := range wantTools {
		if !found {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestToolsNotRegistered_WhenProvidersNil(t *testing.T) {
	srv := New(nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	// Only get_config should be registered.
	if len(result.Tools) != 1 {
		names := make([]string, len(result.Tools))
		for i, tool := range result.Tools {
			names[i] = tool.Name
		}
		t.Errorf("expected 1 tool (get_config) when all providers nil, got %d: %v", len(result.Tools), names)
	}
}

func TestGenerationToolsNotRegistered_WithoutStatusProvider(t *testing.T) {
	session := connectTestClient(t, &mockGenerator{}, nil, nil, nil)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == "text_to_3d" || tool.Name == "image_to_3d" || tool.Name == "multiview_to_3d" {
			t.Fatalf("generation tool %q should not be registered without status support", tool.Name)
		}
	}
}

func TestCommonToolsRegistered_WithoutStatusProvider(t *testing.T) {
	var gen provider.ModelGenerator = &mockGeneratorWithCommon{}
	srv := NewWithOptions(gen, nil, nil, nil, Options{
		Backend:   "tripo",
		OutputDir: t.TempDir(),
		Version:   "test-version",
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	wantCommon := map[string]bool{
		"get_tasks":          false,
		"get_balance":        false,
		"get_usage":          false,
		"upload_file":        false,
		"create_file_upload": false,
	}
	for _, tool := range result.Tools {
		if _, ok := wantCommon[tool.Name]; ok {
			wantCommon[tool.Name] = true
		}
		if tool.Name == "task_status" || tool.Name == "download_model" {
			t.Fatalf("status tool %q should not be registered without status support", tool.Name)
		}
	}
	for name, found := range wantCommon {
		if !found {
			t.Errorf("common tool %q not registered", name)
		}
	}
}

func TestPostProcessToolsNotRegistered_WithoutStatusProvider(t *testing.T) {
	session := connectTestClient(t, nil, nil, &mockPostProc{}, nil)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == "retopologize" || tool.Name == "convert_format" || tool.Name == "stylize" {
			t.Fatalf("post-process tool %q should not be registered without status support", tool.Name)
		}
	}
}

// --- Generation tool tests ---

func TestTextTo3D_Success(t *testing.T) {
	gen := &mockGenerator{
		textResult: &provider.ModelOperation{
			TaskID: "task-txt-1",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, gen, &mockStatus{}, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "text_to_3d",
		Arguments: map[string]any{
			"prompt": "a red dragon",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "3D generation started!")
	assertContentContains(t, res, "task-txt-1")
	assertStructuredField(t, res, "taskId", "task-txt-1")
	assertStructuredField(t, res, "status", "submitted")
}

func TestTextTo3D_Error(t *testing.T) {
	gen := &mockGenerator{
		textErr: errors.New("credit limit exceeded"),
	}

	session := connectTestClient(t, gen, &mockStatus{}, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "text_to_3d",
		Arguments: map[string]any{
			"prompt": "anything",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "credit limit exceeded")
}

func TestImageTo3D_Success(t *testing.T) {
	gen := &mockGenerator{
		imageResult: &provider.ModelOperation{
			TaskID: "task-img-1",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, gen, &mockStatus{}, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "image_to_3d",
		Arguments: map[string]any{
			"imageUrl": "https://example.com/photo.jpg",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "3D generation from image started!")
	assertStructuredField(t, res, "taskId", "task-img-1")
}

func TestMultiviewTo3D_Success(t *testing.T) {
	gen := &mockGenerator{
		multiviewResult: &provider.ModelOperation{
			TaskID: "task-mv-1",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, gen, &mockStatus{}, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "multiview_to_3d",
		Arguments: map[string]any{
			"imageUrls": []string{
				"https://example.com/front.jpg",
				"https://example.com/side.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Multi-view 3D generation started!")
	assertStructuredField(t, res, "taskId", "task-mv-1")
}

func TestTextToImage_Success(t *testing.T) {
	gen := &mockGenerator{
		textImageResult: &provider.ModelOperation{
			TaskID: "task-text-image",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, gen, &mockStatus{}, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "text_to_image",
		Arguments: map[string]any{
			"prompt": "a product photo",
			"model":  "seedream_v5",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatal("expected success, got error result")
	}
	assertContentContains(t, res, "Image generation started")
	assertStructuredField(t, res, "taskId", "task-text-image")
}

// --- Status tool tests ---

func TestTaskStatus_Success(t *testing.T) {
	stat := &mockStatus{
		statusResult: &provider.ModelTaskStatus{
			TaskID:   "task-abc",
			Status:   "running",
			Progress: 65,
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "task_status",
		Arguments: map[string]any{
			"taskId": "task-abc",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "task-abc")
	assertContentContains(t, res, "running")
	assertContentContains(t, res, "65%")
	assertStructuredField(t, res, "taskId", "task-abc")
	assertStructuredField(t, res, "status", "running")
}

func TestTaskStatus_WithError(t *testing.T) {
	stat := &mockStatus{
		statusResult: &provider.ModelTaskStatus{
			TaskID:   "task-fail",
			Status:   "failed",
			Progress: 0,
			Error:    "content moderation rejected",
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "task_status",
		Arguments: map[string]any{
			"taskId": "task-fail",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	assertContentContains(t, res, "content moderation rejected")
}

func TestGetTasks_Success(t *testing.T) {
	stat := &mockStatus{
		batchResult: &provider.BatchTasksResult{
			Tasks: map[string]provider.ModelTaskStatus{
				"task-a": {TaskID: "task-a", Status: "success"},
			},
			Missed: []string{"task-b"},
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_tasks",
		Arguments: map[string]any{
			"taskIds": []string{"task-a", "task-b"},
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatal("expected success, got error result")
	}
	assertContentContains(t, res, "task-a")
	assertContentContains(t, res, "task-b")
}

func TestCreateFileUpload_Success(t *testing.T) {
	stat := &mockStatus{
		presignResult: &provider.FileUpload{
			PresignedURL: "https://storage.example/upload",
			FileToken:    "file_presigned",
			ExpiresIn:    1800,
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "create_file_upload",
		Arguments: map[string]any{
			"format": "glb",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatal("expected success, got error result")
	}
	assertContentContains(t, res, "file_presigned")
	assertStructuredField(t, res, "fileToken", "file_presigned")
}

func TestUploadFile_Success(t *testing.T) {
	stat := &mockStatus{
		uploadResult: &provider.FileUpload{
			FileToken: "file_uploaded",
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "upload_file",
		Arguments: map[string]any{
			"filePath": "/tmp/model.glb",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatal("expected success, got error result")
	}
	assertContentContains(t, res, "file_uploaded")
	assertStructuredField(t, res, "fileToken", "file_uploaded")
}

func TestGetBalance_Success(t *testing.T) {
	stat := &mockStatus{
		balanceResult: &provider.AccountBalance{
			Balance: 10000.5,
			Frozen:  200.25,
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_balance",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatal("expected success, got error result")
	}
	assertContentContains(t, res, "10000.5")
	assertStructuredField(t, res, "balance", 10000.5)
	assertStructuredField(t, res, "frozen", 200.25)
}

func TestGetUsage_Success(t *testing.T) {
	stat := &mockStatus{
		usageResult: &provider.AccountUsageResult{
			Records: []provider.AccountUsageRecord{
				{
					TaskID:          "task-usage",
					Type:            "text_to_model",
					CreditsConsumed: 5.25,
					CreatedAt:       "2026-07-08T10:00:00Z",
				},
			},
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_usage",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatal("expected success, got error result")
	}
	assertContentContains(t, res, "task-usage")
	assertStructuredField(t, res, "records", []any{
		map[string]any{
			"taskId":          "task-usage",
			"type":            "text_to_model",
			"creditsConsumed": 5.25,
			"createdAt":       "2026-07-08T10:00:00Z",
		},
	})
}

func TestDownloadModel_Success(t *testing.T) {
	stat := &mockStatus{
		downloadResult: &provider.ModelResult{
			FilePath: "/tmp/test/model-abc.glb",
			Format:   "GLTF",
			TaskID:   "task-dl-1",
		},
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "download_model",
		Arguments: map[string]any{
			"taskId": "task-dl-1",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Model downloaded!")
	assertContentContains(t, res, "/tmp/test/model-abc.glb")
	assertContentContains(t, res, "GLTF")
	assertStructuredField(t, res, "filePath", "/tmp/test/model-abc.glb")
	assertStructuredField(t, res, "format", "GLTF")
}

func TestDownloadModel_Error(t *testing.T) {
	stat := &mockStatus{
		downloadErr: errors.New("task not complete"),
	}

	session := connectTestClient(t, nil, stat, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "download_model",
		Arguments: map[string]any{
			"taskId": "task-dl-2",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "task not complete")
}

// --- Post-processing tool tests ---

func TestRetopologize_Success(t *testing.T) {
	pp := &mockPostProc{
		retopResult: &provider.ModelOperation{
			TaskID: "retopo-1",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, nil, &mockStatus{}, pp, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "retopologize",
		Arguments: map[string]any{
			"originalTaskId": "orig-task",
			"quad":           true,
			"targetFaces":    4000,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Retopology started")
	assertContentContains(t, res, "quad mesh")
	assertStructuredField(t, res, "taskId", "retopo-1")
}

func TestRetopologize_TriangleMode(t *testing.T) {
	pp := &mockPostProc{
		retopResult: &provider.ModelOperation{
			TaskID: "retopo-2",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, nil, &mockStatus{}, pp, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "retopologize",
		Arguments: map[string]any{
			"originalTaskId": "orig-task",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	assertContentContains(t, res, "triangle mesh")
}

func TestConvertFormat_Success(t *testing.T) {
	pp := &mockPostProc{
		convertResult: &provider.ModelOperation{
			TaskID: "conv-1",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, nil, &mockStatus{}, pp, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "convert_format",
		Arguments: map[string]any{
			"originalTaskId": "orig-task",
			"format":         "FBX",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Format conversion to FBX started!")
	assertStructuredField(t, res, "taskId", "conv-1")
}

func TestStylize_Success(t *testing.T) {
	pp := &mockPostProc{
		stylizeResult: &provider.ModelOperation{
			TaskID: "style-1",
			Status: "submitted",
		},
	}

	session := connectTestClient(t, nil, &mockStatus{}, pp, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "stylize",
		Arguments: map[string]any{
			"originalTaskId": "orig-task",
			"style":          "lego",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Stylization (lego) started!")
	assertStructuredField(t, res, "taskId", "style-1")
}

func TestStylize_Error(t *testing.T) {
	pp := &mockPostProc{
		stylizeErr: errors.New("unsupported style"),
	}

	session := connectTestClient(t, nil, &mockStatus{}, pp, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "stylize",
		Arguments: map[string]any{
			"originalTaskId": "task-1",
			"style":          "bad",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "unsupported style")
}

func TestSegmentMeshAndRigCheck_Success(t *testing.T) {
	pp := &mockPostProc{
		segmentResult: &provider.ModelOperation{TaskID: "segment-1", Status: "submitted"},
		rigCheckResult: &provider.RigCheckResult{
			Riggable: true,
			RigType:  "biped",
		},
	}
	session := connectTestClient(t, nil, &mockStatus{}, pp, nil)

	segment, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "segment_mesh",
		Arguments: map[string]any{
			"input": "task-model",
		},
	})
	if err != nil {
		t.Fatalf("segment_mesh CallTool: %v", err)
	}
	if segment.IsError {
		t.Fatal("expected segment_mesh success")
	}
	assertStructuredField(t, segment, "taskId", "segment-1")

	rigCheck, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "rig_check",
		Arguments: map[string]any{
			"input": "task-model",
		},
	})
	if err != nil {
		t.Fatalf("rig_check CallTool: %v", err)
	}
	if rigCheck.IsError {
		t.Fatal("expected rig_check success")
	}
	assertContentContains(t, rigCheck, "biped")
	assertStructuredField(t, rigCheck, "rigType", "biped")
}

// --- Config tool tests ---

func TestListModels_Success(t *testing.T) {
	lister := &mockModelLister{
		models: []provider.ModelInfo{
			{
				Name:         "Tripo v3.1",
				ID:           "v3.1",
				Description:  "Latest model",
				Capabilities: []string{"text_to_3d", "image_to_3d", "multiview_to_3d"},
			},
		},
	}

	session := connectTestClient(t, nil, nil, nil, lister)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_models",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Built-in Model Catalog")
	assertContentContains(t, res, "Tripo v3.1")
	assertContentContains(t, res, "Latest model")
}

func TestGetConfig_ReturnsBackendInfo(t *testing.T) {
	session := connectTestClient(t, nil, nil, nil, nil)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_config",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Backend: tripo")
	assertContentContains(t, res, "Version: test-version")
	assertStructuredField(t, res, "backend", "tripo")
	assertStructuredField(t, res, "version", "test-version")
}

func TestListModelsNotRegistered_WhenListerNil(t *testing.T) {
	session := connectTestClient(t, nil, nil, nil, nil)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == "list_models" {
			t.Error("list_models should not be registered when model lister is nil")
		}
	}
}

// --- Test helpers ---

func assertContentContains(t *testing.T, res *mcp.CallToolResult, substr string) {
	t.Helper()
	for _, c := range res.Content {
		data, err := json.Marshal(c)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), substr) {
			return
		}
	}
	t.Errorf("no content entry contains %q", substr)
}

func assertStructuredField(t *testing.T, res *mcp.CallToolResult, key string, want any) {
	t.Helper()
	if res.StructuredContent == nil {
		t.Fatalf("structured content is nil, expected field %q=%#v", key, want)
	}

	data, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}

	got, ok := m[key]
	if !ok {
		t.Errorf("structured content missing field %q", key)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("structured content[%q] = %#v, want %#v", key, got, want)
	}
}
