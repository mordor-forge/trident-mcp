package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

// addOperation registers a Tripo Operation as an MCP tool whose input schema
// is the Operation's request type.
func addOperation[R any](s *Server, op tripo.Operation[R]) {
	schema, err := op.InputSchema()
	if err != nil {
		panic(fmt.Sprintf("server: %s: %v", op.Name, err))
	}
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        op.Name,
		Description: op.Description + " Creates a Tripo task (spends credits) and returns its task_id immediately; then call wait_for_task.",
		InputSchema: schema,
		Annotations: creditSpending,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in R) (*mcp.CallToolResult, tripo.Submission, error) {
		sub, err := op.Submit(ctx, s.tripo, in)
		if err != nil {
			return nil, tripo.Submission{}, err
		}
		return text(submissionText(sub)), *sub, nil
	})
}

var nextSteps = map[tripo.Produces]string{
	tripo.ProducesMesh:         "wait_for_task saves the mesh. Its task_id is the input for convert_model, retopologize, smart_uv, texture_model, segment_mesh and rig_check.",
	tripo.ProducesImage:        "wait_for_task saves the image. Its task_id can be the input of image_to_model, image_to_multiview or image_to_image.",
	tripo.ProducesViews:        "wait_for_task saves the views. Pass the task_id as input to multiview_to_model, or to edit_multiview to fix a view.",
	tripo.ProducesSplat:        "wait_for_task saves the .splat file.",
	tripo.ProducesRigCheck:     "wait_for_task reports riggable and rig_type; then rig_model with the same input and that rig_type.",
	tripo.ProducesSegmentation: "wait_for_task saves the segmented result; complete_mesh takes this task_id to fill in the parts.",
}

func submissionText(sub *tripo.Submission) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Started %s task %s", sub.Operation, sub.TaskID)
	if sub.Model != "" {
		fmt.Fprintf(&b, " (model %s)", sub.Model)
	}
	b.WriteString(".\n")
	for _, w := range sub.Warnings {
		fmt.Fprintf(&b, "Warning: %s\n", w)
	}
	fmt.Fprintf(&b, "Next: wait_for_task with task_id %s. %s", sub.TaskID, nextSteps[sub.Produces])
	return b.String()
}
