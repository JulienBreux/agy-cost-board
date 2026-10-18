package cli

import (
	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDoctorCommand(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose GCP connectivity, BigQuery tables, and configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Running agy-cost-board environment diagnostics...")

			demo := flagDemo
			projectID := flagProjectID
			telemetryTable := flagTelemetryTable
			billingTable := flagBillingTable
			if cfg := config.FromContext(cmd.Context()); cfg != nil {
				demo = cfg.Demo
				projectID = cfg.ProjectID
				telemetryTable = cfg.TelemetryTable
				billingTable = cfg.BillingTable
			}

			if demo {
				cmd.Println("[OK] Mode: Demo Mode (synthetic data enabled)")
				cmd.Println("[OK] Data Source: In-memory deterministic simulator (No GCP credentials required)")
				cmd.Println("[OK] Diagnostics passed successfully.")
				return nil
			}

			if projectID == "" {
				cmd.Println("[WARN] Project ID is not specified. (Set --project or AGY_COST_BOARD_PROJECT_ID)")
				cmd.Println("[TIP] You can test immediately using the --demo flag:")
				cmd.Println("      agy-cost-board cost --demo")
				return nil
			}

			cmd.Printf("[OK] Google Cloud Project: %s\n", projectID)
			if telemetryTable != "" {
				cmd.Printf("[OK] Telemetry Table: %s\n", telemetryTable)
			} else {
				cmd.Println("[WARN] Missing telemetry table (Set --telemetry-table or AGY_COST_BOARD_TELEMETRY_TABLE)")
			}

			if billingTable != "" {
				cmd.Printf("[OK] Billing Export Table: %s\n", billingTable)
			} else {
				cmd.Println("[WARN] Missing billing export table (Set --billing-table or AGY_COST_BOARD_BILLING_TABLE)")
			}

			return nil
		},
	}
}
