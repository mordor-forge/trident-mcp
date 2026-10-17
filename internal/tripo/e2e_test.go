//go:build e2e

package tripo_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

// Live tests against the real Tripo API. The default run is free (balance,
// usage, upload, task lookup). TRIPO_E2E_SPEND=1 runs the credit-spending
// lifecycle, about 80 credits, after checking the balance covers it.
//
//	TRIPO_API_KEY=tsk_... go test -tags=e2e -run TestE2E_ ./internal/tripo/ -v -timeout 30m

const spendBudget = 150

const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+tmZ8AAAAASUVORK5CYII="

func liveClient(t *testing.T) *tripo.Client {
	t.Helper()
	key := os.Getenv("TRIPO_API_KEY")
	if key == "" {
		t.Skip("TRIPO_API_KEY is not set")
	}
	c, err := tripo.New(tripo.Options{APIKey: key, BaseURL: os.Getenv("TRIPO_BASE_URL")})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestE2E_FreeCalls(t *testing.T) {
	c := liveClient(t)
	ctx := context.Background()

	b, err := c.GetBalance(ctx)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	t.Logf("balance %.2f, frozen %.2f", b.Balance, b.Frozen)

	u, err := c.GetUsage(ctx, 5, 0)
	if err != nil {
		t.Fatalf("usage: %v", err)
	}
	t.Logf("usage: %d records", len(u.Records))

	data, _ := base64.StdEncoding.DecodeString(tinyPNG)
	path := filepath.Join(t.TempDir(), "tiny.png")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	up, err := c.UploadFile(ctx, path)
	if err != nil || up.FileToken == "" {
		t.Fatalf("upload: %+v, %v", up, err)
	}

	list, err := c.GetTasks(ctx, []string{"00000000-0000-0000-0000-000000000000"})
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}
	t.Logf("unknown task lookup: %d tasks, missed %v", len(list.Tasks), list.Missed)
}

type liveRun struct {
	t   *testing.T
	c   *tripo.Client
	dir string
}

func (r liveRun) wait(name, taskID string) *tripo.Task {
	r.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	task, err := r.c.WaitTask(ctx, taskID, tripo.WaitOptions{Interval: 3 * time.Second})
	if err != nil {
		r.t.Fatalf("%s: waiting for %s: %v", name, taskID, err)
	}
	out, _ := json.Marshal(task.Output)
	r.t.Logf("%s %s: %s, %g credits, output %s", name, taskID, task.Status, task.CreditsConsumed, out)
	if task.Status != tripo.StatusSuccess {
		r.t.Fatalf("%s failed: %s %s", name, task.ErrorCode, task.ErrorMessage)
	}
	return task
}

func (r liveRun) download(name, taskID string) {
	r.t.Helper()
	_, artifacts, err := r.c.Download(context.Background(), taskID, r.dir)
	if err != nil {
		r.t.Fatalf("%s: download: %v", name, err)
	}
	for _, a := range artifacts {
		r.t.Logf("%s saved %s (%s, %d bytes)", name, a.Path, a.Format, a.Bytes)
	}
}

func mustSubmit[R any](t *testing.T, c *tripo.Client, op tripo.Operation[R], req R) string {
	t.Helper()
	sub, err := op.Submit(context.Background(), c, req)
	if err != nil {
		t.Fatalf("%s: %v", op.Name, err)
	}
	t.Logf("%s started %s (model %s) %v", op.Name, sub.TaskID, sub.Model, sub.Warnings)
	return sub.TaskID
}

func TestE2E_SpendingLifecycle(t *testing.T) {
	if os.Getenv("TRIPO_E2E_SPEND") != "1" {
		t.Skip("set TRIPO_E2E_SPEND=1 to spend about 80 credits")
	}
	c := liveClient(t)
	b, err := c.GetBalance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if b.Balance < spendBudget {
		t.Skipf("balance %.0f is below the %d-credit budget", b.Balance, spendBudget)
	}
	dir := os.Getenv("TRIPO_E2E_OUTPUT_DIR")
	if dir == "" {
		dir = t.TempDir()
	}
	r := liveRun{t: t, c: c, dir: dir}
	f := false

	// Dropped model names should be refused by Tripo without charging.
	_, err = tripo.TextToModel.Submit(context.Background(), c, tripo.TextToModelRequest{Prompt: "a cube",
		GenerationOptions: tripo.GenerationOptions{Model: "tripo-turbo", Texture: &f, PBR: &f}})
	var apiErr *tripo.Error
	t.Logf("tripo-turbo: %v (tripo.Error: %v)", err, errors.As(err, &apiErr))

	image := mustSubmit(t, c, tripo.TextToImage, tripo.TextToImageRequest{
		Prompt: "a stylized cartoon knight character standing in a T-pose, full body, plain white background"})
	banana := mustSubmit(t, c, tripo.TextToImage, tripo.TextToImageRequest{
		Prompt: "a small wooden treasure chest, product photo, plain background", ImageOptions: tripo.ImageOptions{Model: "banana", Size: "1024x1024"}})
	mesh := mustSubmit(t, c, tripo.TextToModel, tripo.TextToModelRequest{
		Prompt:            "a stylized cartoon knight character standing in a T-pose, full body",
		GenerationOptions: tripo.GenerationOptions{Texture: &f, PBR: &f}})

	r.wait("text_to_image", image)
	r.download("text_to_image", image)
	r.wait("text_to_image banana", banana)
	views := mustSubmit(t, c, tripo.ImageToMultiview, tripo.ImageToMultiviewRequest{Input: tripo.Input(image)})

	r.wait("text_to_model", mesh)
	r.download("text_to_model", mesh)
	in := tripo.Input(mesh)
	check := mustSubmit(t, c, tripo.RigCheck, tripo.RigCheckRequest{Input: in})
	texture := mustSubmit(t, c, tripo.TextureModel, tripo.TextureModelRequest{Input: in, Model: "v3.5", TextureQuality: "fast",
		TexturePrompt: &tripo.TexturePrompt{Text: "polished steel armor with a blue cloth tabard"}})
	convert := mustSubmit(t, c, tripo.ConvertModel, tripo.ConvertModelRequest{Input: in, Format: "FBX"})
	rig := mustSubmit(t, c, tripo.RigModel, tripo.RigModelRequest{Input: in, Model: "v3.0-20260909"})

	var mu sync.Mutex
	var wg sync.WaitGroup
	results := map[string]*tripo.Task{}
	for name, id := range map[string]string{"image_to_multiview": views, "rig_check": check, "texture_model": texture, "convert_model": convert} {
		wg.Add(1)
		go func() {
			// t.Fatalf must run on the test goroutine, so record failures here.
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
			defer cancel()
			task, err := c.WaitTask(ctx, id, tripo.WaitOptions{Interval: 3 * time.Second})
			if err != nil || task.Status != tripo.StatusSuccess {
				t.Errorf("%s %s: %+v, %v", name, id, task, err)
				return
			}
			out, _ := json.Marshal(task.Output)
			t.Logf("%s %s: %s, %g credits, output %s", name, id, task.Status, task.CreditsConsumed, out)
			mu.Lock()
			results[name] = task
			mu.Unlock()
		}()
	}
	r.wait("rig_model", rig)
	retarget := mustSubmit(t, c, tripo.RetargetAnimation, tripo.RetargetAnimationRequest{Input: tripo.TaskRef(rig), Animation: "preset:walk"})
	r.wait("retarget_animation", retarget)
	r.download("retarget_animation", retarget)
	wg.Wait()
	if t.Failed() {
		t.FailNow()
	}

	riggable, ok := results["rig_check"].Riggable()
	t.Logf("rig_check: riggable=%v (present %v), rig_type %q", riggable, ok, results["rig_check"].RigType())
	keys := make([]string, 0)
	for k := range results["image_to_multiview"].OutputURLs() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	t.Logf("multiview outputs: %s", strings.Join(keys, ", "))
	r.download("convert_model", convert)

	after, err := c.GetBalance(context.Background())
	if err == nil {
		t.Logf("spent %.2f credits", b.Balance-after.Balance)
	}
}
