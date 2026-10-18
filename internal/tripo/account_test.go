package tripo_test

import (
	"context"
	"testing"

	"github.com/mordor-forge/trident-mcp/internal/tripo"
	"github.com/mordor-forge/trident-mcp/internal/tripo/tripotest"
)

func TestGetBalance(t *testing.T) {
	fake := tripotest.New(t)
	fake.Balance = map[string]any{"balance": 48.5, "frozen": 20}
	b, err := fake.Client(t).GetBalance(context.Background())
	if err != nil || b.Balance != 48.5 || b.Frozen != 20 {
		t.Fatalf("balance = %+v, %v", b, err)
	}
}

func TestGetUsage_AcceptsBothShapes(t *testing.T) {
	record := map[string]any{"task_id": "t1", "type": "text_to_model", "credits_consumed": 40}
	for name, usage := range map[string]any{
		"bare array":    []any{record},
		"wrapped items": map[string]any{"items": []any{record}},
	} {
		t.Run(name, func(t *testing.T) {
			fake := tripotest.New(t)
			fake.Usage = usage
			u, err := fake.Client(t).GetUsage(context.Background(), 10, 20)
			if err != nil || len(u.Records) != 1 || u.Records[0].CreditsConsumed != 40 {
				t.Fatalf("usage = %+v, %v", u, err)
			}
			if q := fake.Last(t, "/account/usage").Query; q != "limit=10&offset=20" {
				t.Fatalf("query = %q", q)
			}
		})
	}
}

func TestCatalog_DefaultsAndFilters(t *testing.T) {
	rigs := tripo.Catalog("rig_model")
	if len(rigs) != 3 {
		t.Fatalf("rig versions = %+v", rigs)
	}
	for _, op := range []string{"text_to_model", "image_to_model", "multiview_to_model", "text_to_image", "image_to_image",
		"texture_model", "retopologize", "smart_uv", "segment_mesh", "complete_mesh", "rig_model"} {
		defaults := 0
		for _, m := range tripo.Catalog(op) {
			for _, d := range m.DefaultFor {
				if d == op {
					defaults++
				}
			}
		}
		if defaults != 1 {
			t.Errorf("%s has %d default model versions, want 1", op, defaults)
		}
	}
	for _, m := range tripo.Catalog("text_to_image") {
		if m.ID == "chat_image_1" {
			t.Error("chat_image_1 retired on 2026-10-23")
		}
	}
}
