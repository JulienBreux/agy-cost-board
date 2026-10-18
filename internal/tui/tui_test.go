package tui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/tui"
)

func setupTestModel(t *testing.T) tui.Model {
	t.Helper()
	provider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(provider, 5*time.Minute)
	model := tui.NewModel(engine, 30)

	// Trigger initial load command synchronously for testing
	ctx := t.Context()
	costs, err := engine.GetAttributedCosts(ctx, 30, "")
	if err != nil {
		t.Fatalf("failed to fetch demo costs: %v", err)
	}
	gov, err := engine.GetLicenseGovernance(ctx, 30)
	if err != nil {
		t.Fatalf("failed to fetch demo governance: %v", err)
	}
	overview, err := engine.GetOverviewMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("failed to fetch demo overview: %v", err)
	}

	model.SetData(costs, gov, overview)
	return model
}

func TestTUIModelNavigation(t *testing.T) {
	m := setupTestModel(t)

	t.Run("Tab switching cycles through views", func(t *testing.T) {
		if m.ActiveTab != 0 {
			t.Errorf("expected initial tab 0, got %d", m.ActiveTab)
		}

		// Send tab key message
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(tui.Model)
		if m.ActiveTab != 1 {
			t.Errorf("expected tab 1 after tab press, got %d", m.ActiveTab)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(tui.Model)
		if m.ActiveTab != 2 {
			t.Errorf("expected tab 2 after second tab, got %d", m.ActiveTab)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(tui.Model)
		if m.ActiveTab != 3 {
			t.Errorf("expected tab 3 after third tab, got %d", m.ActiveTab)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(tui.Model)
		if m.ActiveTab != 0 {
			t.Errorf("expected wrap-around to tab 0, got %d", m.ActiveTab)
		}

		// Test shift-tab
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		m = updated.(tui.Model)
		if m.ActiveTab != 3 {
			t.Errorf("expected shift-tab to wrap to tab 3, got %d", m.ActiveTab)
		}

		// Test numeric tab switching (1, 2, 3, 4)
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
		m = updated.(tui.Model)
		if m.ActiveTab != 0 {
			t.Errorf("expected tab 0 on '1', got %d", m.ActiveTab)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
		m = updated.(tui.Model)
		if m.ActiveTab != 1 {
			t.Errorf("expected tab 1 on '2', got %d", m.ActiveTab)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
		m = updated.(tui.Model)
		if m.ActiveTab != 2 {
			t.Errorf("expected tab 2 on '3', got %d", m.ActiveTab)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
		m = updated.(tui.Model)
		if m.ActiveTab != 3 {
			t.Errorf("expected tab 3 on '4', got %d", m.ActiveTab)
		}
	})

	t.Run("Cursor navigation moves through cost rows", func(t *testing.T) {
		m.ActiveTab = 0
		initialCursor := m.Cursor
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(tui.Model)
		if m.Cursor != initialCursor+1 {
			t.Errorf("expected cursor to increment to %d, got %d", initialCursor+1, m.Cursor)
		}

		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = updated.(tui.Model)
		if m.Cursor != initialCursor {
			t.Errorf("expected cursor to decrement back to %d, got %d", initialCursor, m.Cursor)
		}
	})

	t.Run("CycleSort changes sort column", func(t *testing.T) {
		initialSort := m.SortMode
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m = updated.(tui.Model)
		if m.SortMode == initialSort {
			t.Errorf("expected sort mode to change upon 's' keypress")
		}

		// Cycle again
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m = updated.(tui.Model)
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m = updated.(tui.Model)
		if m.SortMode != initialSort {
			t.Errorf("expected sort mode to wrap back to initial")
		}
	})

	t.Run("Enter key selects developer on Costs and Leaderboard and opens breakdown tab", func(t *testing.T) {
		m.ActiveTab = 0
		viewCosts := m.View()
		if !strings.Contains(viewCosts, "ATTRIBUTED COSTS") {
			t.Errorf("expected tab 0 view to contain ATTRIBUTED COSTS")
		}

		m.ActiveTab = 1
		viewLeaderboard := m.View()
		if !strings.Contains(viewLeaderboard, "LEADERBOARD") || !strings.Contains(viewLeaderboard, "#1") {
			t.Errorf("expected tab 1 view to contain LEADERBOARD and #1, got: %s", viewLeaderboard)
		}

		m.ActiveTab = 2
		viewGov := m.View()
		if !strings.Contains(viewGov, "LICENSE GOVERNANCE") {
			t.Errorf("expected tab 2 view to contain LICENSE GOVERNANCE")
		}

		m.ActiveTab = 0
		m.Cursor = 0
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = updated.(tui.Model)

		if m.ActiveTab != 3 {
			t.Errorf("expected enter to switch to tab 3 (User Breakdown), got %d", m.ActiveTab)
		}
		if m.SelectedUser == nil {
			t.Fatalf("expected SelectedUser to be populated")
		}

		view := m.View()
		if !strings.Contains(view, m.SelectedUser.UserID) {
			t.Errorf("view missing selected user email: %s", view)
		}

		// Press esc to return to cost tab
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(tui.Model)
		if m.ActiveTab != 0 {
			t.Errorf("expected esc to return to tab 0, got %d", m.ActiveTab)
		}

		// Navigate and drill down from Leaderboard tab (tab 1)
		m.ActiveTab = 1
		m.Cursor = 0
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(tui.Model)
		if m.Cursor != 1 {
			t.Errorf("expected cursor 1 on Leaderboard tab, got %d", m.Cursor)
		}
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = updated.(tui.Model)
		if m.ActiveTab != 3 {
			t.Errorf("expected enter on Leaderboard tab to switch to tab 3, got %d", m.ActiveTab)
		}
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(tui.Model)
		if m.ActiveTab != 1 {
			t.Errorf("expected esc from Breakdown to return to Leaderboard tab 1, got %d", m.ActiveTab)
		}
	})

	t.Run("WindowSizeMsg updates dimensions", func(t *testing.T) {
		updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		m = updated.(tui.Model)
		if m.Width != 120 || m.Height != 40 {
			t.Errorf("expected 120x40 dimensions, got %dx%d", m.Width, m.Height)
		}
	})

	t.Run("Init cmd loads data via message", func(t *testing.T) {
		initCmd := m.Init()
		if initCmd == nil {
			t.Fatalf("expected non-nil Init command")
		}
		msg := initCmd()
		loadedMsg, ok := msg.(tui.DataLoadedMsg)
		if !ok {
			t.Fatalf("expected DataLoadedMsg, got %T", msg)
		}
		if loadedMsg.Err != nil {
			t.Fatalf("unexpected error in loaded msg: %v", loadedMsg.Err)
		}
		if len(loadedMsg.Costs) == 0 {
			t.Errorf("expected loaded costs, got 0")
		}

		// Update with loaded msg
		updated, _ := m.Update(loadedMsg)
		m = updated.(tui.Model)
		if m.Loading {
			t.Errorf("expected loading=false after DataLoadedMsg")
		}
	})

	t.Run("Quit keys return quit command", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Errorf("expected quit cmd on 'q'")
		}
	})

	t.Run("TUI displays dynamic currency symbol in overview and costs", func(t *testing.T) {
		view := m.View()
		if !strings.Contains(view, "€") {
			t.Errorf("expected TUI overview to contain € in demo mode, got: %s", view)
		}
	})

	t.Run("TUI displays WTU, Cache Hit Rate, Included and PAYG coexistence metrics", func(t *testing.T) {
		m.ActiveTab = 0
		view0 := m.View()
		if !strings.Contains(view0, "WTU") || !strings.Contains(view0, "CACHE %") || !strings.Contains(view0, "Inc:") || !strings.Contains(view0, "PAYG:") {
			t.Errorf("expected TUI tab 0 to display WTU, CACHE %%, Inc:, and PAYG:, got: %s", view0)
		}

		m.ActiveTab = 1
		view1 := m.View()
		if !strings.Contains(view1, "WTU") || !strings.Contains(view1, "CACHE %") || !strings.Contains(view1, "EFFORT") {
			t.Errorf("expected TUI Leaderboard tab 1 to display WTU, CACHE %%, and EFFORT, got: %s", view1)
		}

		m.ActiveTab = 0
		m.Cursor = 0
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = updated.(tui.Model)
		view3 := m.View()
		if !strings.Contains(view3, "WTU:") || !strings.Contains(view3, "Cache Hit:") || !strings.Contains(view3, "Included:") || !strings.Contains(view3, "PAYG:") {
			t.Errorf("expected TUI Developer Breakdown tab 3 to display WTU:, Cache Hit:, Included:, and PAYG:, got: %s", view3)
		}
	})
}

