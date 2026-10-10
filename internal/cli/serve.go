package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/julienbreux/agy-cost-board/internal/server"
	"github.com/julienbreux/agy-cost-board/internal/setup"
	"github.com/julienbreux/agy-cost-board/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newServeCommand(v *viper.Viper) *cobra.Command {
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
		Args:  cobra.NoArgs,
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

			cfg := config.FromContext(ctx)
			projectID := flagProjectID
			telemetryTable := flagTelemetryTable
			billingTable := flagBillingTable
			seatQuota := flagSeatQuota
			demo := flagDemo

			if cfg != nil {
				projectID = cfg.ProjectID
				telemetryTable = cfg.TelemetryTable
				billingTable = cfg.BillingTable
				seatQuota = cfg.SeatQuota
				demo = cfg.Demo
				if !cmd.Flags().Changed("port") && cfg.Port > 0 {
					port = cfg.Port
				}
				if !cmd.Flags().Changed("host") && cfg.Host != "" {
					host = cfg.Host
				}
			}

			setupCfg := setup.Config{
				ProjectID:      projectID,
				TelemetryTable: telemetryTable,
				BillingTable:   billingTable,
				SeatQuota:      seatQuota,
				Demo:           demo || projectID == "",
			}
			srv := server.NewServerWithSetup(engine, staticFS, setupCfg, nil)
			srv.SetVersionInfo(server.VersionInfo{
				Version:   Version,
				Commit:    Commit,
				BuildDate: BuildDate,
			})
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

			// Setup graceful shutdown channel using signal.NotifyContext
			shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			serverErr := make(chan error, 1)
			go func() {
				cmd.Printf("Starting HTTP server on %s...\n", addr)
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					serverErr <- err
				}
			}()

			select {
			case <-shutdownCtx.Done():
				cmd.Printf("\nReceived shutdown signal: shutting down gracefully...\n")
				stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				return httpServer.Shutdown(stopCtx)
			case err := <-serverErr:
				return fmt.Errorf("HTTP server error: %w", err)
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port to listen on (defaults to $PORT or 8080)")
	cmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host address to bind to")

	if v != nil {
		_ = v.BindPFlag("port", cmd.Flags().Lookup("port"))
		_ = v.BindPFlag("host", cmd.Flags().Lookup("host"))
	}

	return cmd
}
