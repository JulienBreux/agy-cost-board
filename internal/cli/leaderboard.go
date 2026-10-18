package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLeaderboardCommand(v *viper.Viper) *cobra.Command {
	var days int
	var paygMode string

	cmd := &cobra.Command{
		Use:   "leaderboard",
		Short: "Display ranked developer leaderboard by spend, tokens, and efficiency",
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

			entries, err := engine.GetLeaderboard(ctx, days)
			if err != nil {
				return fmt.Errorf("calculate developer leaderboard: %w", err)
			}
			entries = attribution.ApplyPAYGModeToLeaderboard(entries, paygMode)

			format := flagFormat
			if cfg := config.FromContext(ctx); cfg != nil && cfg.Format != "" {
				format = cfg.Format
			}

			switch strings.ToLower(format) {
			case "json":
				out, err := FormatJSON(entries)
				if err != nil {
					return err
				}
				cmd.Println(out)

			case "csv":
				headers := []string{"RANK", "USER", "TOTAL_COST", "SPEND_SHARE_PCT", "TOTAL_TOKENS", "ACTIVE_DAYS", "COST_PER_1K_TOKENS", "TOP_MODEL", "CURRENCY"}
				var rows [][]string
				for _, e := range entries {
					rows = append(rows, e.ToCSVRow())
				}
				out, err := FormatCSV(headers, rows)
				if err != nil {
					return err
				}
				cmd.Print(out)

			default:
				headers := []string{"RANK", "USER", "TOTAL SPEND", "INCLUDED", "PAYG", "SPEND SHARE", "TOKENS", "WTU", "CACHE %", "ACTIVE DAYS", "COST / 1K TOKENS", "TOP MODEL", "EFFORT"}
				var rows [][]string
				for _, e := range entries {
					effort := e.TopEffort
					if effort == "" {
						effort = "none"
					}
					rows = append(rows, []string{
						fmt.Sprintf("#%d", e.Rank),
						e.UserID,
						domain.FormatCost(e.TotalCost, e.Currency),
						domain.FormatCost(e.IncludedCost, e.Currency),
						domain.FormatCost(e.PAYGCost, e.Currency),
						fmt.Sprintf("%.1f%%", e.ShareOfSpend),
						strconv.FormatInt(e.TotalTokens, 10),
						strconv.FormatInt(e.WTU, 10),
						fmt.Sprintf("%.1f%%", e.CacheHitRate),
						fmt.Sprintf("%dd", e.ActiveDays),
						fmt.Sprintf("%s%.4f", domain.CurrencySymbol(e.Currency), e.CostPer1KTokens),
						e.TopModel,
						effort,
					})
				}
				cmd.Print(FormatTable(headers, rows))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Time window in days")
	cmd.Flags().StringVar(&paygMode, "payg-mode", "fair_share", "PAYG attribution mode: fair_share or when_used")

	return cmd
}
