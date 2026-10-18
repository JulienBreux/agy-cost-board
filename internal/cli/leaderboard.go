package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/julienbreux/agy-cost-board/internal/config"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLeaderboardCommand(v *viper.Viper) *cobra.Command {
	var days int

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
				headers := []string{"RANK", "USER", "TOTAL SPEND", "SPEND SHARE", "TOKENS", "ACTIVE DAYS", "COST / 1K TOKENS", "TOP MODEL"}
				var rows [][]string
				for _, e := range entries {
					rows = append(rows, []string{
						fmt.Sprintf("#%d", e.Rank),
						e.UserID,
						domain.FormatCost(e.TotalCost, e.Currency),
						fmt.Sprintf("%.1f%%", e.ShareOfSpend),
						strconv.FormatInt(e.TotalTokens, 10),
						fmt.Sprintf("%dd", e.ActiveDays),
						fmt.Sprintf("%s%.4f", domain.CurrencySymbol(e.Currency), e.CostPer1KTokens),
						e.TopModel,
					})
				}
				cmd.Print(FormatTable(headers, rows))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Time window in days")

	return cmd
}
