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

func newUserCommand(v *viper.Viper) *cobra.Command {
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

			format := flagFormat
			if cfg := config.FromContext(ctx); cfg != nil && cfg.Format != "" {
				format = cfg.Format
			}

			switch strings.ToLower(format) {
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
					strconv.FormatInt(summary.TotalTokens, 10),
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
							strconv.FormatInt(detail.Tokens, 10),
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
