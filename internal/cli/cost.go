package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCostCommand(v *viper.Viper) *cobra.Command {
	var days int
	var modelFilter string

	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Display proportional per-user cost attribution",
		Args:  cobra.NoArgs,
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

			format := flagFormat
			if cfg := config.FromContext(ctx); cfg != nil && cfg.Format != "" {
				format = cfg.Format
			}

			switch strings.ToLower(format) {
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
						strconv.FormatInt(c.UserTokens, 10),
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
