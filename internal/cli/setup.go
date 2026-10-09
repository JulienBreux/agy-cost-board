package cli

import (
	"fmt"
	"strings"

	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/julienbreux/agy-cost-board/internal/setup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagSetupSave    bool
	flagSetupCreate  bool
	flagSetupDryRun  bool
	flagSetupSink    string
	flagSetupDataset string
)

func newSetupCommand(v *viper.Viper) *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Verify GCP prerequisites and optionally provision BigQuery & Logging sinks",
		Args:  cobra.NoArgs,
		Long: `setup validates the Google Cloud Application Default Credentials, BigQuery datasets,
Cloud Logging sinks, and billing export prerequisites needed for agy-cost-board to compute AI cost attribution.
Use --create or --dry-run to generate or provision the required telemetry sink.
Use --save to persist the configuration locally to .agy-cost-board.yaml for subsequent commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appCfg := config.FromContext(ctx)

			projID := flagProjectID
			telemetryTbl := flagTelemetryTable
			billingTbl := flagBillingTable
			seatQuota := flagSeatQuota
			demo := flagDemo
			format := flagFormat
			sinkName := flagSetupSink
			datasetName := flagSetupDataset

			if appCfg != nil {
				if appCfg.ProjectID != "" {
					projID = appCfg.ProjectID
				}
				if appCfg.TelemetryTable != "" {
					telemetryTbl = appCfg.TelemetryTable
				}
				if appCfg.BillingTable != "" {
					billingTbl = appCfg.BillingTable
				}
				if appCfg.SeatQuota > 0 {
					seatQuota = appCfg.SeatQuota
				}
				demo = appCfg.Demo
				if appCfg.Format != "" {
					format = appCfg.Format
				}
				if !cmd.Flags().Changed("sink-name") && appCfg.SinkName != "" {
					sinkName = appCfg.SinkName
				}
				if !cmd.Flags().Changed("dataset") && appCfg.DatasetName != "" {
					datasetName = appCfg.DatasetName
				}
			}

			// 1. Provisioning plan if requested (--create or --dry-run)
			if flagSetupCreate || flagSetupDryRun {
				plan := setup.GenerateProvisionPlan(projID, datasetName, sinkName)
				out, err := plan.Execute(ctx, flagSetupDryRun)
				if err != nil {
					return fmt.Errorf("resource provisioning error: %w", err)
				}
				for _, line := range out {
					fmt.Fprintln(cmd.OutOrStdout(), line)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "")
			}

			// 2. Build configuration for diagnostics
			if demo {
				if projID == "" {
					projID = "demo-project"
				}
				if telemetryTbl == "" {
					telemetryTbl = "demo-project.antigravity_telemetry.inference_logs"
				}
				if billingTbl == "" {
					billingTbl = "demo-project.billing_export.gcp_billing_export_v1_000"
				}
			}

			cfg := setup.Config{
				ProjectID:      projID,
				TelemetryTable: telemetryTbl,
				BillingTable:   billingTbl,
				SinkName:       sinkName,
				DatasetName:    datasetName,
				SeatQuota:      seatQuota,
				Demo:           demo,
			}

			// 3. Run all diagnostic checks
			runner := setup.NewRunner()
			report := runner.RunAll(ctx, cfg)

			// 4. Output rendering
			switch strings.ToLower(format) {
			case "json":
				out, err := FormatJSON(report)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), out)
			default:
				fmt.Fprintln(cmd.OutOrStdout(), FormatTable(domain.DiagnosticTableHeaders(), report.ToTableRows()))
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"\nOVERALL STATUS: %s | PASSED: %d | WARNINGS: %d | ERRORS: %d\n",
					report.OverallStatus, report.PassedCount, report.WarningCount, report.ErrorCount,
				)

				// Print remediation commands for any warnings or errors
				var remediations []domain.CheckResult
				for _, c := range report.Checks {
					if c.Status != domain.StatusOK && c.RemediationCommand != "" {
						remediations = append(remediations, c)
					}
				}

				if len(remediations) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "\nRecommended Action Items:")
					for _, r := range remediations {
						fmt.Fprintf(cmd.OutOrStdout(), "  • [%s] %s\n    Command: %s\n", r.Status, r.Name, r.RemediationCommand)
					}
				}
			}

			// 5. Persist configuration if --save was specified
			if flagSetupSave {
				savePath := flagConfigFile
				if savePath == "" {
					savePath = config.DefaultConfigFileName
				}
				fileCfg := &config.Config{
					ProjectID:      projID,
					TelemetryTable: telemetryTbl,
					BillingTable:   billingTbl,
					SinkName:       sinkName,
					DatasetName:    datasetName,
					SeatQuota:      seatQuota,
				}
				if err := config.Save(savePath, fileCfg); err != nil {
					return fmt.Errorf("failed to save config to %s: %w", savePath, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nConfiguration successfully persisted to %s\n", savePath)
			}

			return nil
		},
	}

	setupCmd.Flags().BoolVar(&flagSetupSave, "save", false, "Save configuration to .agy-cost-board.yaml")
	setupCmd.Flags().BoolVar(&flagSetupCreate, "create", false, "Provision missing BigQuery dataset and Cloud Logging sink")
	setupCmd.Flags().BoolVar(&flagSetupDryRun, "dry-run", false, "Simulate provisioning and print CLI commands")
	setupCmd.Flags().StringVar(&flagSetupSink, "sink-name", "agy-inference-sink", "Cloud Logging sink name")
	setupCmd.Flags().StringVar(&flagSetupDataset, "dataset", "antigravity_telemetry", "BigQuery telemetry dataset")

	if v != nil {
		_ = v.BindPFlag("sink_name", setupCmd.Flags().Lookup("sink-name"))
		_ = v.BindPFlag("dataset", setupCmd.Flags().Lookup("dataset"))
	}

	return setupCmd
}
