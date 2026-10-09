package cli

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/julienbreux/agy-cost-board/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newTUICommand(v *viper.Viper) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive terminal dashboard",
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
			p := tea.NewProgram(tui.NewModel(engine, days), tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Lookback window in days")

	return cmd
}
