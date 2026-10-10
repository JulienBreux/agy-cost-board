package cli

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version metadata set at build time via ldflags.
var (
	Version   = "v0.4.0"
	Commit    = "none"
	BuildDate = "unknown"
)

// Global CLI options
var (
	flagConfigFile     string
	flagDemo           bool
	flagProjectID      string
	flagTelemetryTable string
	flagBillingTable   string
	flagSeatQuota      int
	flagFormat         string
)

// ResetFlags restores default flag values across unit test runs.
func ResetFlags() {
	flagConfigFile = ""
	flagDemo = false
	flagProjectID = ""
	flagTelemetryTable = ""
	flagBillingTable = ""
	flagSeatQuota = 10
	flagFormat = "table"
	flagSetupSave = false
	flagSetupCreate = false
	flagSetupDryRun = false
	flagSetupSink = "agy-inference-sink"
	flagSetupDataset = "antigravity_telemetry"
}

// SyncFlags updates the package-level flag variables with resolved configuration values.
func SyncFlags(cfg *config.Config) {
	if cfg == nil {
		return
	}
	flagDemo = cfg.Demo
	if cfg.ProjectID != "" {
		flagProjectID = cfg.ProjectID
	}
	if cfg.TelemetryTable != "" {
		flagTelemetryTable = cfg.TelemetryTable
	}
	if cfg.BillingTable != "" {
		flagBillingTable = cfg.BillingTable
	}
	if cfg.SeatQuota > 0 {
		flagSeatQuota = cfg.SeatQuota
	}
	if cfg.Format != "" {
		flagFormat = cfg.Format
	}
	if cfg.SinkName != "" {
		flagSetupSink = cfg.SinkName
	}
	if cfg.DatasetName != "" {
		flagSetupDataset = cfg.DatasetName
	}
}

// NewRootCommand creates the top-level Cobra command with an isolated Viper instance.
func NewRootCommand() *cobra.Command {
	v := viper.New()
	return NewRootCommandWithViper(v)
}

// NewRootCommandWithViper creates the top-level Cobra command and binds flags to the provided Viper instance.
func NewRootCommandWithViper(v *viper.Viper) *cobra.Command {
	config.SetupViper(v)

	rootCmd := &cobra.Command{
		Use:           "agy-cost-board",
		Short:         "Antigravity & Gemini Enterprise Cost Attribution Board",
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate),
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `agy-cost-board reconciles Google Cloud BigQuery inference telemetry with GCP billing exports
to compute proportional, per-user AI costs and track Gemini Enterprise seat utilization.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadViper(v, flagConfigFile)
			if err != nil {
				return fmt.Errorf("configuration file error: %w", err)
			}

			// Propagate resolved config via context
			cmd.SetContext(config.WithConfig(cmd.Context(), cfg))

			// Synchronize flag variables so direct references observe resolved configuration
			SyncFlags(cfg)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&flagConfigFile, "config", "", "Path to configuration file (default .agy-cost-board.yaml)")
	rootCmd.PersistentFlags().BoolVar(&flagDemo, "demo", false, "Use realistic synthetic demo data without GCP connection")
	rootCmd.PersistentFlags().StringVar(&flagProjectID, "project", "", "Google Cloud Project ID")
	rootCmd.PersistentFlags().StringVar(&flagTelemetryTable, "telemetry-table", "", "BigQuery table for inference logs")
	rootCmd.PersistentFlags().StringVar(&flagBillingTable, "billing-table", "", "BigQuery table for GCP billing export")
	rootCmd.PersistentFlags().IntVar(&flagSeatQuota, "seat-quota", 10, "Gemini Enterprise seat quota")
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "table", "Output format: table, json, or csv")

	// Bind persistent Cobra flags directly to Viper keys
	_ = v.BindPFlag("demo", rootCmd.PersistentFlags().Lookup("demo"))
	_ = v.BindPFlag("project_id", rootCmd.PersistentFlags().Lookup("project"))
	_ = v.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	_ = v.BindPFlag("telemetry_table", rootCmd.PersistentFlags().Lookup("telemetry-table"))
	_ = v.BindPFlag("telemetry-table", rootCmd.PersistentFlags().Lookup("telemetry-table"))
	_ = v.BindPFlag("billing_table", rootCmd.PersistentFlags().Lookup("billing-table"))
	_ = v.BindPFlag("billing-table", rootCmd.PersistentFlags().Lookup("billing-table"))
	_ = v.BindPFlag("seat_quota", rootCmd.PersistentFlags().Lookup("seat-quota"))
	_ = v.BindPFlag("seat-quota", rootCmd.PersistentFlags().Lookup("seat-quota"))
	_ = v.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

	rootCmd.AddCommand(newSetupCommand(v))
	rootCmd.AddCommand(newCostCommand(v))
	rootCmd.AddCommand(newLicenseCommand(v))
	rootCmd.AddCommand(newUserCommand(v))
	rootCmd.AddCommand(newDoctorCommand(v))
	rootCmd.AddCommand(newTUICommand(v))
	rootCmd.AddCommand(newServeCommand(v))

	return rootCmd
}

// BuildEngine constructs the attribution Engine from flags/resolved configuration or falls back to demo mode.
func BuildEngine(ctx context.Context) (*attribution.Engine, string, error) {
	cfg := config.FromContext(ctx)

	demo := flagDemo
	projectID := flagProjectID
	telemetryTable := flagTelemetryTable
	billingTable := flagBillingTable
	seatQuota := flagSeatQuota

	if cfg != nil {
		demo = cfg.Demo
		projectID = cmp.Or(cfg.ProjectID, projectID)
		telemetryTable = cmp.Or(cfg.TelemetryTable, telemetryTable)
		billingTable = cmp.Or(cfg.BillingTable, billingTable)
		if cfg.SeatQuota > 0 {
			seatQuota = cfg.SeatQuota
		}
	}

	if demo || projectID == "" {
		return attribution.NewEngine(bigquery.NewDemoDataProviderWithQuota(seatQuota), 5*time.Minute), "demo (synthetic data)", nil
	}

	clientCfg := bigquery.ClientConfig{
		ProjectID:      projectID,
		TelemetryTable: telemetryTable,
		BillingTable:   billingTable,
		SeatQuota:      seatQuota,
	}

	bqClient, err := bigquery.NewBigQueryClient(ctx, clientCfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize BigQuery client: %w", err)
	}

	return attribution.NewEngine(bqClient, 15*time.Minute), "Google Cloud BigQuery (" + projectID + ")", nil
}

// Execute runs the root CLI command.
func Execute() error {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	return nil
}
