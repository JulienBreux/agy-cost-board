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

func newLicenseCommand(v *viper.Viper) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "license",
		Short: "Display Gemini Enterprise seat utilization and dormant license reclamation",
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

			gov, err := engine.GetLicenseGovernance(ctx, days)
			if err != nil {
				return fmt.Errorf("calculate license governance: %w", err)
			}

			format := flagFormat
			if cfg := config.FromContext(ctx); cfg != nil && cfg.Format != "" {
				format = cfg.Format
			}

			switch strings.ToLower(format) {
			case "json":
				out, err := FormatJSON(gov)
				if err != nil {
					return err
				}
				cmd.Println(out)

			case "csv":
				headers := []string{"SEAT_QUOTA", "ASSIGNED_SEATS", "ACTIVE_SEATS", "DORMANT_SEATS", "UTILIZATION_PCT", "ESTIMATED_MONTHLY_SAVINGS", "CURRENCY"}
				rows := [][]string{{
					strconv.Itoa(gov.SeatQuota),
					strconv.Itoa(gov.AssignedSeats),
					strconv.Itoa(gov.ActiveSeats),
					strconv.Itoa(gov.DormantSeats),
					fmt.Sprintf("%.1f%%", gov.UtilizationPct),
					fmt.Sprintf("%.2f", gov.EstimatedMonthlySavings),
					gov.Currency,
				}}
				out, err := FormatCSV(headers, rows)
				if err != nil {
					return err
				}
				cmd.Print(out)

			default:
				headers := []string{"SEAT QUOTA", "ASSIGNED", "STANDARD", "PLUS", "TRIAL", "ACTIVE", "DORMANT", "UTILIZATION", "EST. MONTHLY SAVINGS"}
				rows := [][]string{{
					strconv.Itoa(gov.SeatQuota),
					strconv.Itoa(gov.AssignedSeats),
					strconv.Itoa(gov.StandardSeats),
					strconv.Itoa(gov.PlusSeats),
					strconv.Itoa(gov.TrialSeats),
					strconv.Itoa(gov.ActiveSeats),
					strconv.Itoa(gov.DormantSeats),
					fmt.Sprintf("%.1f%%", gov.UtilizationPct),
					domain.FormatCost(gov.EstimatedMonthlySavings, gov.Currency),
				}}
				cmd.Print(FormatTable(headers, rows))

				if gov.WeeklyPool != nil {
					cmd.Printf("\nWEEKLY ALLOWANCE POOL (%s / %s): %s / %s (%.1f%% drawdown) | Resets in %dh\n",
						gov.WeeklyPool.PoolMode,
						gov.WeeklyPool.SeatBasis,
						domain.FormatCost(gov.WeeklyPool.IncludedUsedUSD, gov.Currency),
						domain.FormatCost(gov.WeeklyPool.PoolLimitUSD, gov.Currency),
						gov.WeeklyPool.UtilizationPct,
						gov.WeeklyPool.ResetInHours,
					)
				}

				if len(gov.DormantUsers) > 0 {
					cmd.Println("\nDormant Licenses (Inactive for >30 days):")
					dormantHeaders := []string{"USER ID", "TIER", "LOCATION", "STATUS", "LAST ACTIVITY"}
					var dormantRows [][]string
					for _, u := range gov.DormantUsers {
						lastAct := "Never"
						if !u.LastActivity.IsZero() {
							lastAct = u.LastActivity.Format("2006-01-02")
						}
						tier := u.Tier
						if u.IsTrial {
							tier += " (Trial)"
						}
						dormantRows = append(dormantRows, []string{
							u.UserID,
							tier,
							u.Location,
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
