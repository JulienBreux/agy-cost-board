package cli

import (
	"context"
	"errors"

	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/julienbreux/agy-cost-board/internal/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newMCPCommand(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the Model Context Protocol (MCP) server over stdio",
		Long: `Starts a Model Context Protocol (MCP) JSON-RPC 2.0 server over standard input/output (stdio).
Designed for integration with Antigravity IDE/CLI and MCP-compatible agents.
Can be disabled via AGY_COST_BOARD_MCP_ENABLED=false or --mcp-enabled=false.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			mcpEnabled := flagMCPEnabled
			if cfg := config.FromContext(ctx); cfg != nil {
				mcpEnabled = cfg.MCPEnabled
			}
			if !mcpEnabled {
				return errors.New("MCP server is disabled (set AGY_COST_BOARD_MCP_ENABLED=true or --mcp-enabled=true to enable)")
			}

			engine, _, err := BuildEngine(ctx)
			if err != nil {
				return err
			}

			srv := mcp.NewServer(engine, Version)
			return srv.ServeStdio(ctx, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	return cmd
}
