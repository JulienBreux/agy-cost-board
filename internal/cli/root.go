package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/julienbreux/agy-cost-board/internal/server"
	"github.com/julienbreux/agy-cost-board/internal/setup"
	"github.com/julienbreux/agy-cost-board/internal/tui"
	"github.com/julienbreux/agy-cost-board/web"
	"github.com/spf13/cobra"
)

// Version metadata set at build time via ldflags.
var (
	Version   = "dev"
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

// NewRootCommand creates the top-level Cobra command with subcommands.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "agy-cost-board",
		Short:   "Antigravity & Gemini Enterprise Cost Attribution Board",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate),
		Long: `agy-cost-board reconciles Google Cloud BigQuery inference telemetry with GCP billing exports
to compute proportional, per-user AI costs and track Gemini Enterprise seat utilization.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			fileCfg, err := config.Load(flagConfigFile)
			if err != nil {
				return fmt.Errorf("configuration file error: %w", err)
			}
			if fileCfg != nil {
				if flagProjectID == "" && fileCfg.ProjectID != "" {
					flagProjectID = fileCfg.ProjectID
				}
				if flagTelemetryTable == "" && fileCfg.TelemetryTable != "" {
					flagTelemetryTable = fileCfg.TelemetryTable
				}
				if flagBillingTable == "" && fileCfg.BillingTable != "" {
					flagBillingTable = fileCfg.BillingTable
				}
				if !cmd.Flags().Changed("seat-quota") && fileCfg.SeatQuota > 0 {
					flagSeatQuota = fileCfg.SeatQuota
				}
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&flagConfigFile, "config", "", "Path to configuration file (default .agy-cost-board.yaml)")
	rootCmd.PersistentFlags().BoolVar(&flagDemo, "demo", false, "Use realistic synthetic demo data without GCP connection")
	rootCmd.PersistentFlags().StringVar(&flagProjectID, "project", os.Getenv("GCP_PROJECT"), "Google Cloud Project ID")
	rootCmd.PersistentFlags().StringVar(&flagTelemetryTable, "telemetry-table", os.Getenv("TELEMETRY_TABLE"), "BigQuery table for inference logs")
	rootCmd.PersistentFlags().StringVar(&flagBillingTable, "billing-table", os.Getenv("BILLING_TABLE"), "BigQuery table for GCP billing export")
	rootCmd.PersistentFlags().IntVar(&flagSeatQuota, "seat-quota", 10, "Gemini Enterprise seat quota")
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "table", "Output format: table, json, or csv")

	rootCmd.AddCommand(newSetupCommand())
	rootCmd.AddCommand(newCostCommand())
	rootCmd.AddCommand(newLicenseCommand())
	rootCmd.AddCommand(newUserCommand())
	rootCmd.AddCommand(newDoctorCommand())
	rootCmd.AddCommand(newTUICommand())
	rootCmd.AddCommand(newServeCommand())

	return rootCmd
}

// BuildEngine constructs the attribution Engine from flags or falls back to demo mode.
func BuildEngine(ctx context.Context) (*attribution.Engine, string, error) {
	if flagDemo || flagProjectID == "" {
		return attribution.NewEngine(bigquery.NewDemoDataProviderWithQuota(flagSeatQuota), 5*time.Minute), "demo (synthetic data)", nil
	}

	cfg := bigquery.ClientConfig{
		ProjectID:      flagProjectID,
		TelemetryTable: flagTelemetryTable,
		BillingTable:   flagBillingTable,
		SeatQuota:      flagSeatQuota,
	}

	bqClient, err := bigquery.NewBigQueryClient(ctx, cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize BigQuery client: %w", err)
	}

	return attribution.NewEngine(bqClient, 15*time.Minute), "Google Cloud BigQuery (" + flagProjectID + ")", nil
}

func newCostCommand() *cobra.Command {
	var days int
	var modelFilter string

	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Display proportional per-user cost attribution",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			engine, _, err := BuildEngine(ctx)
			if err != nil {
				return err
			}

			costs, err := engine.GetAttributedCosts(ctx, days, modelFilter)
			if err != nil {
				return fmt.Errorf("calculate attributed costs: %w", err)
			}

			switch strings.ToLower(flagFormat) {
			case "json":
				out, err := FormatJSON(costs)
				if err != nil {
					return err
				}
				cmd.Println(out)

			case "csv":
				headers := []string{"USER", "MODEL", "USAGE_DATE", "USER_TOKENS", "TOTAL_MODEL_TOKENS", "TOKEN_SHARE_PCT", "ALLOCATED_COST_USD"}
				var rows [][]string
				for _, c := range costs {
					rows = append(rows, c.ToCSVRow())
				}
				out, err := FormatCSV(headers, rows)
				if err != nil {
					return err
				}
				cmd.Print(out)

			default:
				headers := []string{"USER", "MODEL", "DATE", "TOKENS", "SHARE %", "ALLOCATED COST"}
				var rows [][]string
				for _, c := range costs {
					rows = append(rows, []string{
						c.UserID,
						c.Model,
						c.UsageDate,
						fmt.Sprintf("%d", c.UserTokens),
						fmt.Sprintf("%.1f%%", c.TokenShare*100),
						fmt.Sprintf("$%.2f", c.AllocatedCost),
					})
				}
				cmd.Print(FormatTable(headers, rows))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Time window in days")
	cmd.Flags().StringVar(&modelFilter, "model", "", "Filter by AI model (e.g. gemini-1.5-pro)")
	return cmd
}

func newLicenseCommand() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "license",
		Short: "Display Gemini Enterprise seat utilization and dormant license reclamation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			engine, _, err := BuildEngine(ctx)
			if err != nil {
				return err
			}

			gov, err := engine.GetLicenseGovernance(ctx, days)
			if err != nil {
				return fmt.Errorf("calculate license governance: %w", err)
			}

			switch strings.ToLower(flagFormat) {
			case "json":
				out, err := FormatJSON(gov)
				if err != nil {
					return err
				}
				cmd.Println(out)

			case "csv":
				headers := []string{"SEAT_QUOTA", "ASSIGNED_SEATS", "ACTIVE_SEATS", "DORMANT_SEATS", "UTILIZATION_PCT", "ESTIMATED_MONTHLY_SAVINGS_USD"}
				rows := [][]string{{
					fmt.Sprintf("%d", gov.SeatQuota),
					fmt.Sprintf("%d", gov.AssignedSeats),
					fmt.Sprintf("%d", gov.ActiveSeats),
					fmt.Sprintf("%d", gov.DormantSeats),
					fmt.Sprintf("%.1f%%", gov.UtilizationPct),
					fmt.Sprintf("$%.2f", gov.EstimatedMonthlySavings),
				}}
				out, err := FormatCSV(headers, rows)
				if err != nil {
					return err
				}
				cmd.Print(out)

			default:
				headers := []string{"SEAT QUOTA", "ASSIGNED", "ACTIVE", "DORMANT", "UTILIZATION", "EST. MONTHLY SAVINGS"}
				rows := [][]string{{
					fmt.Sprintf("%d", gov.SeatQuota),
					fmt.Sprintf("%d", gov.AssignedSeats),
					fmt.Sprintf("%d", gov.ActiveSeats),
					fmt.Sprintf("%d", gov.DormantSeats),
					fmt.Sprintf("%.1f%%", gov.UtilizationPct),
					fmt.Sprintf("$%.2f", gov.EstimatedMonthlySavings),
				}}
				cmd.Print(FormatTable(headers, rows))

				if len(gov.DormantUsers) > 0 {
					cmd.Println("\nDormant Licenses (Inactive for >30 days):")
					dormantHeaders := []string{"USER ID", "STATUS", "LAST ACTIVITY"}
					var dormantRows [][]string
					for _, u := range gov.DormantUsers {
						lastAct := "Never"
						if !u.LastActivity.IsZero() {
							lastAct = u.LastActivity.Format("2006-01-02")
						}
						dormantRows = append(dormantRows, []string{
							u.UserID,
							string(u.Status),
							lastAct,
						})
					}
					cmd.Print(FormatTable(dormantHeaders, dormantRows))
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Lookback window in days")
	return cmd
}

func newUserCommand() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "user <email>",
		Short: "Display detailed token volume and cost history for a single developer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			email := args[0]
			engine, _, err := BuildEngine(ctx)
			if err != nil {
				return err
			}

			summary, err := engine.GetUserSummary(ctx, email, days)
			if err != nil {
				return fmt.Errorf("fetch user summary: %w", err)
			}

			switch strings.ToLower(flagFormat) {
			case "json":
				out, err := FormatJSON(summary)
				if err != nil {
					return err
				}
				cmd.Println(out)

			default:
				headers := []string{"USER", "STATUS", "TOTAL TOKENS", "TOTAL COST", "LAST ACTIVE"}
				lastAct := "Never"
				if !summary.LastActive.IsZero() {
					lastAct = summary.LastActive.Format("2006-01-02 15:04")
				}
				rows := [][]string{{
					summary.UserID,
					string(summary.SeatStatus),
					fmt.Sprintf("%d", summary.TotalTokens),
					fmt.Sprintf("$%.2f", summary.TotalCost),
					lastAct,
				}}
				cmd.Print(FormatTable(headers, rows))

				if len(summary.ModelBreakdown) > 0 {
					cmd.Println("\nModel Usage Breakdown:")
					mHeaders := []string{"MODEL", "TOKENS", "SHARE %", "COST"}
					var mRows [][]string
					for m, detail := range summary.ModelBreakdown {
						mRows = append(mRows, []string{
							m,
							fmt.Sprintf("%d", detail.Tokens),
							fmt.Sprintf("%.1f%%", detail.Share*100),
							fmt.Sprintf("$%.2f", detail.Cost),
						})
					}
					cmd.Print(FormatTable(mHeaders, mRows))
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Lookback window in days")
	return cmd
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose GCP connectivity, BigQuery tables, and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Running agy-cost-board environment diagnostics...")

			if flagDemo {
				cmd.Println("[OK] Mode: Demo Mode (synthetic data enabled)")
				cmd.Println("[OK] Data Source: In-memory deterministic simulator (No GCP credentials required)")
				cmd.Println("[OK] Diagnostics passed successfully.")
				return nil
			}

			if flagProjectID == "" {
				cmd.Println("[WARN] Project ID is not specified. (Set --project or GCP_PROJECT)")
				cmd.Println("[TIP] You can test immediately using the --demo flag:")
				cmd.Println("      agy-cost-board cost --demo")
				return nil
			}

			cmd.Printf("[OK] Google Cloud Project: %s\n", flagProjectID)
			if flagTelemetryTable != "" {
				cmd.Printf("[OK] Telemetry Table: %s\n", flagTelemetryTable)
			} else {
				cmd.Println("[WARN] Missing telemetry table (Set --telemetry-table or TELEMETRY_TABLE)")
			}

			if flagBillingTable != "" {
				cmd.Printf("[OK] Billing Export Table: %s\n", flagBillingTable)
			} else {
				cmd.Println("[WARN] Missing billing export table (Set --billing-table or BILLING_TABLE)")
			}

			return nil
		},
	}
}

func newTUICommand() *cobra.Command {
	var days int
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive terminal dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			engine, _, err := BuildEngine(ctx)
			if err != nil {
				return err
			}
			p := tea.NewProgram(tui.NewModel(engine, days), tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
	cmd.Flags().IntVar(&days, "days", 30, "Lookback window in days")
	return cmd
}

func newServeCommand() *cobra.Command {
	var port int
	var host string

	// Cloud Run sets PORT environment variable
	defaultPort := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			defaultPort = p
		}
	}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web dashboard HTTP server with embedded React SPA",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			engine, source, err := BuildEngine(ctx)
			if err != nil {
				return err
			}

			staticFS, err := web.FS()
			if err != nil {
				return fmt.Errorf("failed to load embedded web assets: %w", err)
			}

			setupCfg := setup.Config{
				ProjectID:      flagProjectID,
				TelemetryTable: flagTelemetryTable,
				BillingTable:   flagBillingTable,
				SeatQuota:      flagSeatQuota,
				Demo:           flagDemo || flagProjectID == "",
			}
			srv := server.NewServerWithSetup(engine, staticFS, setupCfg, nil)
			addr := fmt.Sprintf("%s:%d", host, port)

			httpServer := &http.Server{
				Addr:         addr,
				Handler:      srv.Router(),
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			cmd.Printf("┌────────────────────────────────────────────────────────┐\n")
			cmd.Printf("│ AGY & Gemini Enterprise Cost Attribution Board         │\n")
			cmd.Printf("├────────────────────────────────────────────────────────┤\n")
			cmd.Printf("│ Data Source : %-41s│\n", source)
			cmd.Printf("│ Dashboard   : http://localhost:%-25d│\n", port)
			cmd.Printf("│ Healthz     : http://localhost:%-25s│\n", fmt.Sprintf("%d/healthz", port))
			cmd.Printf("└────────────────────────────────────────────────────────┘\n\n")

			// Setup graceful shutdown channel
			shutdownChan := make(chan os.Signal, 1)
			signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

			serverErr := make(chan error, 1)
			go func() {
				cmd.Printf("Starting HTTP server on %s...\n", addr)
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					serverErr <- err
				}
			}()

			select {
			case sig := <-shutdownChan:
				cmd.Printf("\nReceived signal %s: shutting down gracefully...\n", sig)
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				return httpServer.Shutdown(shutdownCtx)
			case err := <-serverErr:
				return fmt.Errorf("HTTP server error: %w", err)
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port to listen on (defaults to $PORT or 8080)")
	cmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host address to bind to")

	return cmd
}

// Execute runs the root CLI command.
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
