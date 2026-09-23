package setup

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ProvisionPlan encapsulates shell commands and descriptions required to provision GCP resources.
type ProvisionPlan struct {
	Commands     []string
	Descriptions []string
}

// GenerateProvisionPlan creates the sequence of CLI commands to configure BigQuery and Cloud Logging.
func GenerateProvisionPlan(projectID, datasetID, sinkName string) ProvisionPlan {
	if datasetID == "" {
		datasetID = "antigravity_telemetry"
	}
	if sinkName == "" {
		sinkName = "agy-inference-sink"
	}
	if projectID == "" {
		projectID = "YOUR_PROJECT_ID"
	}

	cmd1 := fmt.Sprintf("bq mk --dataset --description \"Antigravity Telemetry Dataset\" %s:%s", projectID, datasetID)
	desc1 := fmt.Sprintf("Create BigQuery destination dataset '%s' in project '%s'", datasetID, projectID)

	cmd2 := fmt.Sprintf(
		"gcloud logging sinks create %s bigquery.googleapis.com/projects/%s/datasets/%s --log-filter='jsonPayload.log_type=\"InferenceResponseLog\"' --use-partitioned-tables",
		sinkName, projectID, datasetID,
	)
	desc2 := fmt.Sprintf("Create Cloud Logging sink '%s' filtering InferenceResponseLog to BigQuery", sinkName)

	return ProvisionPlan{
		Commands:     []string{cmd1, cmd2},
		Descriptions: []string{desc1, desc2},
	}
}

// Execute runs each provisioning command or simulates execution if dryRun is true.
func (p ProvisionPlan) Execute(ctx context.Context, dryRun bool) ([]string, error) {
	results := make([]string, len(p.Commands))

	for i, cmdStr := range p.Commands {
		if dryRun {
			results[i] = fmt.Sprintf("[DRY-RUN] %s", cmdStr)
			continue
		}

		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return results, fmt.Errorf("command '%s' failed: %v, output: %s", cmdStr, err, string(output))
		}
		results[i] = string(output)
	}

	return results, nil
}
