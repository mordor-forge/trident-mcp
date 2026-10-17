package server

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

const (
	defaultWaitSeconds = 120
	maxWaitSeconds     = 900
)

type taskInput struct {
	TaskID string `json:"task_id" jsonschema:"Task ID returned by a task-creating tool"`
}

type tasksInput struct {
	TaskIDs []string `json:"task_ids" jsonschema:"Up to 100 task IDs"`
}

type waitInput struct {
	TaskID         string `json:"task_id" jsonschema:"Task ID returned by a task-creating tool"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"How long to wait before returning with the task still running (default 120, max 900). Call again to keep waiting"`
	Download       *bool  `json:"download,omitempty" jsonschema:"Save the outputs when the task succeeds (default true)"`
}

type downloadInput struct {
	TaskID  string   `json:"task_id" jsonschema:"Task ID of a successful task"`
	Outputs []string `json:"outputs,omitempty" jsonschema:"Only save these outputs, e.g. model, rendered_image, generated_image, front_view (default all)"`
}

type waitResult struct {
	Task *tripo.Task `json:"task"`
	// Done is false when the timeout passed before the task finished.
	Done      bool             `json:"done"`
	Artifacts []tripo.Artifact `json:"artifacts,omitempty"`
}

type downloadResult struct {
	Task      *tripo.Task      `json:"task"`
	Artifacts []tripo.Artifact `json:"artifacts"`
}

func (s *Server) registerTaskTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "wait_for_task",
		Description: "Wait for a Tripo task to finish, sending progress notifications, then save its outputs (meshes, images, views, splats) to the output directory. Returns with done=false if timeout_seconds passes first; call again to keep waiting.",
		Annotations: savesFiles,
	}, s.handleWait)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_task",
		Description: "Get a Tripo task's status, progress, outputs and error. Output URLs are signed and expire; use download_task to save them.",
		Annotations: readOnly,
	}, s.handleGetTask)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_tasks",
		Description: "Get up to 100 Tripo tasks at once. Unknown IDs are listed under missed.",
		Annotations: readOnly,
	}, s.handleGetTasks)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "download_task",
		Description: "Save the outputs of a successful Tripo task to <output dir>/<task_id>/<output>.<ext>, overwriting earlier downloads of it. Download formats are what the task produced; use convert_model for other formats.",
		Annotations: savesFiles,
	}, s.handleDownload)
}

func (s *Server) handleWait(ctx context.Context, req *mcp.CallToolRequest, in waitInput) (*mcp.CallToolResult, waitResult, error) {
	timeout := in.TimeoutSeconds
	if timeout <= 0 {
		timeout = defaultWaitSeconds
	}
	timeout = min(timeout, maxWaitSeconds)
	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	token := req.Params.GetProgressToken()
	sent := -1.0
	task, err := s.tripo.WaitTask(waitCtx, in.TaskID, tripo.WaitOptions{Interval: s.pollInterval, OnUpdate: func(t *tripo.Task) {
		// MCP requires progress to increase with every notification.
		if progress := float64(t.Progress); token != nil && progress > sent {
			sent = progress
			_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
				ProgressToken: token, Progress: progress, Total: 100, Message: progressMessage(t),
			})
		}
	}})
	// Only our own deadline means "still running"; an HTTP timeout inside a
	// poll also matches context.DeadlineExceeded.
	if err != nil && waitCtx.Err() != nil && ctx.Err() == nil && task != nil {
		result := waitResult{Task: task}
		return text(fmt.Sprintf("%s\nStill not finished after %d s; call wait_for_task again to keep waiting.", taskText(task), timeout)), result, nil
	}
	if err != nil {
		return nil, waitResult{}, err
	}

	result := waitResult{Task: task, Done: true}
	if task.Status != tripo.StatusSuccess || (in.Download != nil && !*in.Download) {
		return text(taskText(task)), result, nil
	}
	task, artifacts, err := s.tripo.Download(ctx, in.TaskID, s.outputDir)
	if err != nil {
		return nil, waitResult{}, err
	}
	result.Task, result.Artifacts = task, artifacts
	return text(taskText(task) + "\n" + artifactsText(artifacts)), result, nil
}

func (s *Server) handleGetTask(ctx context.Context, _ *mcp.CallToolRequest, in taskInput) (*mcp.CallToolResult, tripo.Task, error) {
	task, err := s.tripo.GetTask(ctx, in.TaskID)
	if err != nil {
		return nil, tripo.Task{}, err
	}
	return text(taskText(task)), *task, nil
}

func (s *Server) handleGetTasks(ctx context.Context, _ *mcp.CallToolRequest, in tasksInput) (*mcp.CallToolResult, tripo.TaskList, error) {
	list, err := s.tripo.GetTasks(ctx, in.TaskIDs)
	if err != nil {
		return nil, tripo.TaskList{}, err
	}
	var b strings.Builder
	for _, id := range in.TaskIDs {
		if task, ok := list.Tasks[id]; ok {
			fmt.Fprintf(&b, "- %s\n", strings.ReplaceAll(taskText(task), "\n", "\n  "))
		}
	}
	if len(list.Missed) > 0 {
		fmt.Fprintf(&b, "Not found: %s\n", strings.Join(list.Missed, ", "))
	}
	return text(strings.TrimRight(b.String(), "\n")), *list, nil
}

func (s *Server) handleDownload(ctx context.Context, _ *mcp.CallToolRequest, in downloadInput) (*mcp.CallToolResult, downloadResult, error) {
	task, artifacts, err := s.tripo.Download(ctx, in.TaskID, s.outputDir, in.Outputs...)
	if err != nil {
		return nil, downloadResult{}, err
	}
	return text(artifactsText(artifacts)), downloadResult{Task: task, Artifacts: artifacts}, nil
}

func progressMessage(t *tripo.Task) string {
	switch {
	case t.QueuingNum != nil && t.Status == tripo.StatusQueued:
		return fmt.Sprintf("queued (position %d)", *t.QueuingNum)
	case t.RunningLeftTime != nil && t.Status == tripo.StatusRunning:
		return fmt.Sprintf("running, %d%% (about %d s left)", t.Progress, *t.RunningLeftTime)
	default:
		return fmt.Sprintf("%s, %d%%", t.Status, t.Progress)
	}
}

func taskText(t *tripo.Task) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task %s", t.TaskID)
	if t.Type != "" {
		fmt.Fprintf(&b, " (%s)", t.Type)
	}
	fmt.Fprintf(&b, ": %s", progressMessage(t))
	if t.ErrorMessage != "" || t.ErrorCode != "" {
		fmt.Fprintf(&b, "\nError: %s", t.ErrorMessage)
		if t.ErrorCode != "" {
			fmt.Fprintf(&b, " (code %s)", t.ErrorCode)
		}
	}
	if riggable, ok := t.Riggable(); ok {
		fmt.Fprintf(&b, "\nRiggable: %t", riggable)
		if rt := t.RigType(); rt != "" {
			fmt.Fprintf(&b, " (rig_type %s)", rt)
		}
	}
	if urls := t.OutputURLs(); len(urls) > 0 && t.Status == tripo.StatusSuccess {
		names := make([]string, 0, len(urls))
		for name := range urls {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprintf(&b, "\nOutputs: %s", strings.Join(names, ", "))
	}
	if t.CreditsConsumed > 0 {
		fmt.Fprintf(&b, "\nCredits: %g", t.CreditsConsumed)
	}
	return b.String()
}

func artifactsText(artifacts []tripo.Artifact) string {
	if len(artifacts) == 0 {
		return "The task has no files to save."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Saved %d file(s):", len(artifacts))
	for _, a := range artifacts {
		fmt.Fprintf(&b, "\n- %s: %s (%s)", a.Name, a.Path, humanBytes(a.Bytes))
	}
	return b.String()
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
