package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/julienbreux/agy-ge-board/internal/attribution"
	"github.com/julienbreux/agy-ge-board/internal/domain"
)

// SortModes
const (
	SortByCostDesc = iota
	SortByTokensDesc
	SortByUserAsc
)

// UI Theme Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a73e8")).
			MarginBottom(1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#1a73e8")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9aa0a6")).
				Background(lipgloss.Color("#2d3033")).
				Padding(0, 2)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3c4043")).
			Padding(0, 1).
			MarginRight(2)

	cardValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#137333"))

	cardWarnValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#e37400"))

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8ab4f8")).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#3c4043"))

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#303134"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5f6368")).
			MarginTop(1)
)

// Model represents the Bubbletea terminal UI state.
type Model struct {
	Engine       *attribution.Engine
	Days         int
	ActiveTab    int
	Cursor       int
	SortMode     int
	Costs        []domain.AllocatedUserCost
	Governance   *domain.LicenseGovernanceSummary
	Overview     *domain.OverviewMetrics
	SelectedUser *domain.UserCostSummary
	Loading      bool
	Err          error
	Width        int
	Height       int
}

// NewModel initializes the Bubbletea state machine.
func NewModel(engine *attribution.Engine, days int) Model {
	if days <= 0 {
		days = 30
	}
	return Model{
		Engine:    engine,
		Days:      days,
		ActiveTab: 0,
		Cursor:    0,
		SortMode:  SortByCostDesc,
		Loading:   true,
	}
}

// SetData injects loaded data directly (used in tests and after async fetch).
func (m *Model) SetData(costs []domain.AllocatedUserCost, gov *domain.LicenseGovernanceSummary, overview *domain.OverviewMetrics) {
	m.Costs = costs
	m.Governance = gov
	m.Overview = overview
	m.Loading = false
	m.sortCosts()
}

// DataLoadedMsg represents async completion of data loading.
type DataLoadedMsg struct {
	Costs    []domain.AllocatedUserCost
	Gov      *domain.LicenseGovernanceSummary
	Overview *domain.OverviewMetrics
	Err      error
}

// Init triggers initial async data loading.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		if m.Engine == nil {
			return DataLoadedMsg{Err: fmt.Errorf("no attribution engine configured")}
		}
		ctx := context.Background()
		costs, err := m.Engine.GetAttributedCosts(ctx, m.Days, "")
		if err != nil {
			return DataLoadedMsg{Err: err}
		}
		gov, err := m.Engine.GetLicenseGovernance(ctx, m.Days)
		if err != nil {
			return DataLoadedMsg{Err: err}
		}
		overview, err := m.Engine.GetOverviewMetrics(ctx, m.Days)
		if err != nil {
			return DataLoadedMsg{Err: err}
		}
		return DataLoadedMsg{
			Costs:    costs,
			Gov:      gov,
			Overview: overview,
		}
	}
}

// Update processes Bubbletea messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case DataLoadedMsg:
		if msg.Err != nil {
			m.Err = msg.Err
			m.Loading = false
			return m, nil
		}
		m.SetData(msg.Costs, msg.Gov, msg.Overview)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 3
			m.Cursor = 0
			return m, nil

		case "shift+tab":
			m.ActiveTab = (m.ActiveTab + 2) % 3
			m.Cursor = 0
			return m, nil

		case "1":
			m.ActiveTab = 0
			m.Cursor = 0
			return m, nil

		case "2":
			m.ActiveTab = 1
			m.Cursor = 0
			return m, nil

		case "3":
			m.ActiveTab = 2
			m.Cursor = 0
			return m, nil

		case "down", "j":
			if m.ActiveTab == 0 && m.Cursor < len(m.Costs)-1 {
				m.Cursor++
			}
			return m, nil

		case "up", "k":
			if m.ActiveTab == 0 && m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil

		case "s":
			m.SortMode = (m.SortMode + 1) % 3
			m.sortCosts()
			return m, nil

		case "enter":
			if m.ActiveTab == 0 && len(m.Costs) > m.Cursor {
				selectedEmail := m.Costs[m.Cursor].UserID
				if m.Engine != nil {
					summary, err := m.Engine.GetUserSummary(context.Background(), selectedEmail, m.Days)
					if err == nil {
						m.SelectedUser = summary
						m.ActiveTab = 2
					}
				}
			}
			return m, nil

		case "esc", "b":
			if m.ActiveTab == 2 {
				m.ActiveTab = 0
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) sortCosts() {
	if len(m.Costs) == 0 {
		return
	}
	switch m.SortMode {
	case SortByCostDesc:
		sort.Slice(m.Costs, func(i, j int) bool {
			return m.Costs[i].AllocatedCost > m.Costs[j].AllocatedCost
		})
	case SortByTokensDesc:
		sort.Slice(m.Costs, func(i, j int) bool {
			return m.Costs[i].UserTokens > m.Costs[j].UserTokens
		})
	case SortByUserAsc:
		sort.Slice(m.Costs, func(i, j int) bool {
			return m.Costs[i].UserID < m.Costs[j].UserID
		})
	}
}

// View renders the terminal dashboard.
func (m Model) View() string {
	var sb strings.Builder

	// Top Title Banner
	sb.WriteString(titleStyle.Render("⚡ AGY & GEMINI ENTERPRISE COST ATTRIBUTION BOARD"))
	sb.WriteString("\n")

	// Tabs Bar
	tabs := []string{"[1] ATTRIBUTED COSTS", "[2] LICENSE GOVERNANCE", "[3] DEVELOPER BREAKDOWN"}
	var renderedTabs []string
	for i, t := range tabs {
		if i == m.ActiveTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(t))
		}
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
	sb.WriteString("\n\n")

	// KPI Cards
	if m.Overview != nil {
		card1 := cardStyle.Render(fmt.Sprintf("TOTAL BILLED\n%s", cardValueStyle.Render(fmt.Sprintf("$%.2f", m.Overview.TotalBilledCost))))
		card2 := cardStyle.Render(fmt.Sprintf("TOTAL TOKENS\n%s", cardValueStyle.Render(fmt.Sprintf("%d", m.Overview.TotalTokens))))
		card3 := cardStyle.Render(fmt.Sprintf("ACTIVE USERS\n%s", cardValueStyle.Render(fmt.Sprintf("%d", m.Overview.ActiveUsersCount))))
		card4 := cardStyle.Render(fmt.Sprintf("POTENTIAL SAVINGS\n%s", cardWarnValueStyle.Render(fmt.Sprintf("$%.2f/mo", m.Overview.EstimatedMonthlySavings))))
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, card1, card2, card3, card4))
		sb.WriteString("\n\n")
	}

	// Body based on ActiveTab
	switch m.ActiveTab {
	case 0:
		sb.WriteString(m.renderCostsView())
	case 1:
		sb.WriteString(m.renderGovernanceView())
	case 2:
		sb.WriteString(m.renderUserBreakdownView())
	}

	// Footer Help
	sb.WriteString(helpStyle.Render("\n[Tab] Next View | [↑/↓] Navigate | [Enter] Drilldown | [s] Cycle Sort | [q] Quit"))
	return sb.String()
}

func (m Model) renderCostsView() string {
	var sb strings.Builder
	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-28s %-20s %-12s %-14s %-10s %-12s",
		"USER", "MODEL", "DATE", "TOKENS", "SHARE %", "COST (USD)")))
	sb.WriteString("\n")

	if len(m.Costs) == 0 {
		sb.WriteString("  No attributed cost records found.\n")
		return sb.String()
	}

	maxRows := 12
	start := 0
	if m.Cursor >= maxRows {
		start = m.Cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(m.Costs) {
		end = len(m.Costs)
	}

	for i := start; i < end; i++ {
		c := m.Costs[i]
		prefix := "  "
		if i == m.Cursor {
			prefix = "> "
		}
		rowStr := fmt.Sprintf("%s%-26s %-20s %-12s %-14d %-10s $%.2f",
			prefix, c.UserID, c.Model, c.UsageDate, c.UserTokens,
			fmt.Sprintf("%.1f%%", c.TokenShare*100), c.AllocatedCost)

		if i == m.Cursor {
			sb.WriteString(selectedRowStyle.Render(rowStr))
		} else {
			sb.WriteString(rowStr)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderGovernanceView() string {
	var sb strings.Builder
	if m.Governance == nil {
		sb.WriteString("  No license governance data available.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("GEMINI ENTERPRISE SEAT QUOTA: %d\n", m.Governance.SeatQuota))
	sb.WriteString(fmt.Sprintf("ASSIGNED SEATS: %d | ACTIVE: %d | DORMANT: %d\n",
		m.Governance.AssignedSeats, m.Governance.ActiveSeats, m.Governance.DormantSeats))
	sb.WriteString(fmt.Sprintf("SEAT UTILIZATION: %.1f%%\n\n", m.Governance.UtilizationPct))

	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-32s %-12s %-16s %-12s",
		"USER ID", "STATUS", "LAST ACTIVITY", "ACTION")))
	sb.WriteString("\n")

	if len(m.Governance.DormantUsers) == 0 {
		sb.WriteString("  All assigned seats are active! 0 dormant licenses found.\n")
	} else {
		for _, u := range m.Governance.DormantUsers {
			lastAct := "Never"
			if !u.LastActivity.IsZero() {
				lastAct = u.LastActivity.Format("2006-01-02")
			}
			sb.WriteString(fmt.Sprintf("  %-30s %-12s %-16s %-12s\n",
				u.UserID, string(u.Status), lastAct, "Reclaim Seat"))
		}
	}

	return sb.String()
}

func (m Model) renderUserBreakdownView() string {
	var sb strings.Builder
	if m.SelectedUser == nil {
		sb.WriteString("  No developer selected. Return to [1] Costs view and press [Enter] on a user row.\n")
		return sb.String()
	}

	u := m.SelectedUser
	sb.WriteString(fmt.Sprintf("DEVELOPER BREAKDOWN: %s\n", u.UserID))
	sb.WriteString(fmt.Sprintf("Status: %s | Total Tokens: %d | Total Attributed Cost: $%.2f\n\n",
		u.SeatStatus, u.TotalTokens, u.TotalCost))

	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-24s %-16s %-12s %-12s",
		"MODEL", "TOKENS", "TOKEN SHARE", "COST (USD)")))
	sb.WriteString("\n")

	for model, detail := range u.ModelBreakdown {
		sb.WriteString(fmt.Sprintf("  %-22s %-16d %-12s $%.2f\n",
			model, detail.Tokens, fmt.Sprintf("%.1f%%", detail.Share*100), detail.Cost))
	}

	return sb.String()
}
