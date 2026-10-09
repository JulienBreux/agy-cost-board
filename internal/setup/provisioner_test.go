package setup_test

import (
	"strings"
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/setup"
)

func TestProvisionPlanGenerator(t *testing.T) {
	t.Run("generates valid bq mk and gcloud logging sink commands", func(t *testing.T) {
		plan := setup.GenerateProvisionPlan("my-gcp-project", "antigravity_telemetry", "agy-inference-sink")

		if len(plan.Commands) < 2 {
			t.Fatalf("expected at least 2 commands, got %d", len(plan.Commands))
		}

		bqCmd := plan.Commands[0]
		if !strings.Contains(bqCmd, "bq mk --dataset") || !strings.Contains(bqCmd, "my-gcp-project:antigravity_telemetry") {
			t.Errorf("expected bq mk command for dataset, got: %s", bqCmd)
		}

		sinkCmd := plan.Commands[1]
		if !strings.Contains(sinkCmd, "gcloud logging sinks create agy-inference-sink") ||
			!strings.Contains(sinkCmd, "bigquery.googleapis.com/projects/my-gcp-project/datasets/antigravity_telemetry") ||
			!strings.Contains(sinkCmd, `jsonPayload.log_type="InferenceResponseLog"`) {
			t.Errorf("expected gcloud logging sink command with filter, got: %s", sinkCmd)
		}
	})

	t.Run("uses defaults when sinkName or datasetID is empty", func(t *testing.T) {
		plan := setup.GenerateProvisionPlan("my-gcp-project", "", "")
		if !strings.Contains(plan.Commands[0], "antigravity_telemetry") {
			t.Errorf("expected default dataset name, got: %s", plan.Commands[0])
		}
		if !strings.Contains(plan.Commands[1], "agy-inference-sink") {
			t.Errorf("expected default sink name, got: %s", plan.Commands[1])
		}
	})

	t.Run("dry-run execution returns command strings without calling shell", func(t *testing.T) {
		plan := setup.GenerateProvisionPlan("test-project", "ds", "sink")
		outputs, err := plan.Execute(t.Context(), true)
		if err != nil {
			t.Fatalf("unexpected error on dry run: %v", err)
		}
		if len(outputs) != len(plan.Commands) {
			t.Errorf("expected %d outputs, got %d", len(plan.Commands), len(outputs))
		}
		for i, out := range outputs {
			if !strings.HasPrefix(out, "[DRY-RUN]") {
				t.Errorf("expected [DRY-RUN] prefix in output %d, got: %s", i, out)
			}
		}
	})

	t.Run("live execution runs commands and handles failures", func(t *testing.T) {
		plan := setup.ProvisionPlan{
			Commands: []string{
				"",
				"echo provision-test-ok",
			},
		}
		outputs, err := plan.Execute(t.Context(), false)
		if err != nil {
			t.Fatalf("unexpected error running echo command: %v", err)
		}
		if len(outputs) != 2 {
			t.Fatalf("expected 2 outputs, got %d", len(outputs))
		}
		if !strings.Contains(outputs[1], "provision-test-ok") {
			t.Errorf("expected output to contain provision-test-ok, got: %s", outputs[1])
		}

		failPlan := setup.ProvisionPlan{
			Commands: []string{"non_existent_binary_for_test_12345"},
		}
		_, err = failPlan.Execute(t.Context(), false)
		if err == nil {
			t.Error("expected error for non-existent binary, got nil")
		}
	})
}
