package tui

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/domain"
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
	PreviousTab  int
	Cursor       int
	SortMode     int
	PAYGMode     string
	Costs        []domain.AllocatedUserCost
	Leaderboard  []domain.LeaderboardEntry
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
		PAYGMode:  "fair_share",
		Loading:   true,
	}
}

// SetData injects loaded data directly (used in tests and after async fetch).
func (m *Model) SetData(costs []domain.AllocatedUserCost, gov *domain.LicenseGovernanceSummary, overview *domain.OverviewMetrics) {
	m.Costs = costs
	m.Governance = gov
	m.Overview = overview
	if len(m.Leaderboard) == 0 && m.Engine != nil {
		if lb, err := m.Engine.GetLeaderboard(context.Background(), m.Days); err == nil {
			m.Leaderboard = lb
		}
	}
	m.Loading = false
	m.sortCosts()
	m.sortLeaderboard()
}

// DataLoadedMsg represents async completion of data loading.
type DataLoadedMsg struct {
	Costs       []domain.AllocatedUserCost
	Leaderboard []domain.LeaderboardEntry
	Gov         *domain.LicenseGovernanceSummary
	Overview    *domain.OverviewMetrics
	Err         error
}

// Init triggers initial async data loading.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		if m.Engine == nil {
			return DataLoadedMsg{Err: errors.New("no attribution engine configured")}
		}
		ctx := context.Background()
		costs, err := m.Engine.GetAttributedCosts(ctx, m.Days, "")
		if err != nil {
			return DataLoadedMsg{Err: err}
		}
		leaderboard, err := m.Engine.GetLeaderboard(ctx, m.Days)
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
			Costs:       costs,
			Leaderboard: leaderboard,
			Gov:         gov,
			Overview:    overview,
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
		m.Leaderboard = msg.Leaderboard
		m.SetData(msg.Costs, msg.Gov, msg.Overview)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 4
			m.Cursor = 0
			return m, nil

		case "shift+tab":
			m.ActiveTab = (m.ActiveTab + 3) % 4
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

		case "4":
			m.ActiveTab = 3
			m.Cursor = 0
			return m, nil

		case "down", "j":
			if m.ActiveTab == 0 && m.Cursor < len(m.Costs)-1 {
				m.Cursor++
			} else if m.ActiveTab == 1 && m.Cursor < len(m.Leaderboard)-1 {
				m.Cursor++
			}
			return m, nil

		case "up", "k":
			if (m.ActiveTab == 0 || m.ActiveTab == 1) && m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil

		case "s":
			m.SortMode = (m.SortMode + 1) % 3
			m.sortCosts()
			m.sortLeaderboard()
			return m, nil

		case "p":
			if m.PAYGMode == "when_used" {
				m.PAYGMode = "fair_share"
			} else {
				m.PAYGMode = "when_used"
			}
			m.Costs = attribution.ApplyPAYGModeToCosts(m.Costs, m.PAYGMode)
			m.Leaderboard = attribution.ApplyPAYGModeToLeaderboard(m.Leaderboard, m.PAYGMode)
			m.sortCosts()
			m.sortLeaderboard()
			return m, nil

		case "enter":
			var selectedEmail string
			if m.ActiveTab == 0 && len(m.Costs) > m.Cursor {
				selectedEmail = m.Costs[m.Cursor].UserID
			} else if m.ActiveTab == 1 && len(m.Leaderboard) > m.Cursor {
				selectedEmail = m.Leaderboard[m.Cursor].UserID
			}
			if selectedEmail != "" && m.Engine != nil {
				summary, err := m.Engine.GetUserSummary(context.Background(), selectedEmail, m.Days)
				if err == nil {
					m.SelectedUser = summary
					m.PreviousTab = m.ActiveTab
					m.ActiveTab = 3
				}
			}
			return m, nil

		case "esc", "b":
			if m.ActiveTab == 3 {
				m.ActiveTab = m.PreviousTab
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
		slices.SortFunc(m.Costs, func(a, b domain.AllocatedUserCost) int {
			return cmp.Compare(b.AllocatedCost, a.AllocatedCost)
		})
	case SortByTokensDesc:
		slices.SortFunc(m.Costs, func(a, b domain.AllocatedUserCost) int {
			return cmp.Compare(b.UserTokens, a.UserTokens)
		})
	case SortByUserAsc:
		slices.SortFunc(m.Costs, func(a, b domain.AllocatedUserCost) int {
			return cmp.Compare(a.UserID, b.UserID)
		})
	}
}

func (m *Model) sortLeaderboard() {
	if len(m.Leaderboard) == 0 {
		return
	}
	switch m.SortMode {
	case SortByCostDesc:
		slices.SortFunc(m.Leaderboard, func(a, b domain.LeaderboardEntry) int {
			if a.TotalCost != b.TotalCost {
				return cmp.Compare(b.TotalCost, a.TotalCost)
			}
			return cmp.Compare(a.Rank, b.Rank)
		})
	case SortByTokensDesc:
		slices.SortFunc(m.Leaderboard, func(a, b domain.LeaderboardEntry) int {
			return cmp.Compare(b.TotalTokens, a.TotalTokens)
		})
	case SortByUserAsc:
		slices.SortFunc(m.Leaderboard, func(a, b domain.LeaderboardEntry) int {
			return cmp.Compare(a.UserID, b.UserID)
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
	tabs := []string{"[1] ATTRIBUTED COSTS", "[2] LEADERBOARD", "[3] LICENSE GOVERNANCE", "[4] DEVELOPER BREAKDOWN"}
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
		card1 := cardStyle.Render("TOTAL BILLED\n" + cardValueStyle.Render(domain.FormatCost(m.Overview.TotalBilledCost, m.Overview.Currency)) +
			fmt.Sprintf("\nInc: %s | PAYG: %s", domain.FormatCost(m.Overview.IncludedCost, m.Overview.Currency), domain.FormatCost(m.Overview.PAYGCost, m.Overview.Currency)))
		card2 := cardStyle.Render("TOTAL TOKENS / WTU\n" + cardValueStyle.Render(strconv.FormatInt(m.Overview.TotalTokens, 10)) +
			fmt.Sprintf("\nWTU: %d | Cache: %.1f%%", m.Overview.TotalWTU, m.Overview.CacheHitRate))
		card3 := cardStyle.Render("ACTIVE USERS\n" + cardValueStyle.Render(strconv.Itoa(m.Overview.ActiveUsersCount)))
		card4 := cardStyle.Render("POTENTIAL SAVINGS\n" + cardWarnValueStyle.Render(domain.FormatCost(m.Overview.EstimatedMonthlySavings, m.Overview.Currency) + "/mo"))
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, card1, card2, card3, card4))
		sb.WriteString("\n\n")
	}

	// Body based on ActiveTab
	switch m.ActiveTab {
	case 0:
		sb.WriteString(m.renderCostsView())
	case 1:
		sb.WriteString(m.renderLeaderboardView())
	case 2:
		sb.WriteString(m.renderGovernanceView())
	case 3:
		sb.WriteString(m.renderUserBreakdownView())
	}

	// Footer Help
	sb.WriteString(helpStyle.Render("\n[Tab] Next View | [↑/↓] Navigate | [Enter] Drilldown | [s] Cycle Sort | [p] PAYG Mode | [q] Quit"))
	return sb.String()
}

func (m Model) renderCostsView() string {
	var sb strings.Builder
	modeLabel := m.PAYGMode
	if modeLabel == "" {
		modeLabel = "fair_share"
	}
	sb.WriteString(fmt.Sprintf("PAYG Attribution Mode: %s (press 'p' to toggle)\n", modeLabel))
	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-26s %-18s %-12s %-12s %-10s %-10s %-10s %-12s",
		"USER", "MODEL", "DATE", "TOKENS", "WTU", "CACHE %", "SHARE %", "COST")))
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
	end := min(start+maxRows, len(m.Costs))

	for i := start; i < end; i++ {
		c := m.Costs[i]
		prefix := "  "
		if i == m.Cursor {
			prefix = "> "
		}
		rowStr := fmt.Sprintf("%s%-24s %-18s %-12s %-12d %-10d %-10s %-10s %s",
			prefix, c.UserID, c.Model, c.UsageDate, c.UserTokens,
			c.WTU, fmt.Sprintf("%.1f%%", c.CacheHitRate),
			fmt.Sprintf("%.1f%%", c.TokenShare*100), domain.FormatCost(c.AllocatedCost, c.Currency))

		if i == m.Cursor {
			sb.WriteString(selectedRowStyle.Render(rowStr))
		} else {
			sb.WriteString(rowStr)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderLeaderboardView() string {
	var sb strings.Builder
	modeLabel := m.PAYGMode
	if modeLabel == "" {
		modeLabel = "fair_share"
	}
	sb.WriteString(fmt.Sprintf("PAYG Attribution Mode: %s (press 'p' to toggle)\n", modeLabel))
	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-6s %-26s %-14s %-10s %-12s %-10s %-10s %-8s %-12s %-18s %-10s",
		"RANK", "DEVELOPER", "TOTAL SPEND", "SHARE %", "TOKENS", "WTU", "CACHE %", "DAYS", "COST / 1K", "TOP MODEL", "EFFORT")))
	sb.WriteString("\n")

	if len(m.Leaderboard) == 0 {
		sb.WriteString("  No developer leaderboard entries found.\n")
		return sb.String()
	}

	maxRows := 12
	start := 0
	if m.Cursor >= maxRows {
		start = m.Cursor - maxRows + 1
	}
	end := min(start+maxRows, len(m.Leaderboard))

	for i := start; i < end; i++ {
		e := m.Leaderboard[i]
		prefix := "  "
		if i == m.Cursor {
			prefix = "> "
		}
		effort := e.TopEffort
		if effort == "" {
			effort = "none"
		}
		rowStr := fmt.Sprintf("%s#%-3d %-26s %-14s %-10s %-12d %-10d %-10s %-8s %-12s %-18s %s",
			prefix, e.Rank, e.UserID, domain.FormatCost(e.TotalCost, e.Currency),
			fmt.Sprintf("%.1f%%", e.ShareOfSpend), e.TotalTokens,
			e.WTU, fmt.Sprintf("%.1f%%", e.CacheHitRate),
			fmt.Sprintf("%dd", e.ActiveDays),
			fmt.Sprintf("%s%.4f", domain.CurrencySymbol(e.Currency), e.CostPer1KTokens), e.TopModel, effort)

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
	sb.WriteString(fmt.Sprintf("ASSIGNED SEATS: %d | STANDARD: %d | PLUS: %d | TRIAL: %d\n",
		m.Governance.AssignedSeats, m.Governance.StandardSeats, m.Governance.PlusSeats, m.Governance.TrialSeats))
	sb.WriteString(fmt.Sprintf("ACTIVE: %d | DORMANT: %d | SEAT UTILIZATION: %.1f%%\n",
		m.Governance.ActiveSeats, m.Governance.DormantSeats, m.Governance.UtilizationPct))
	if m.Governance.WeeklyPool != nil {
		sb.WriteString(fmt.Sprintf("WEEKLY ALLOWANCE POOL (%s / %s): %s / %s (%.1f%% drawdown) | Resets in %dh\n",
			m.Governance.WeeklyPool.PoolMode,
			m.Governance.WeeklyPool.SeatBasis,
			domain.FormatCost(m.Governance.WeeklyPool.IncludedUsedUSD, m.Governance.Currency),
			domain.FormatCost(m.Governance.WeeklyPool.PoolLimitUSD, m.Governance.Currency),
			m.Governance.WeeklyPool.UtilizationPct,
			m.Governance.WeeklyPool.ResetInHours))
	}
	sb.WriteString("\n")

	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-30s %-14s %-10s %-12s %-16s %-12s",
		"USER ID", "TIER", "LOCATION", "STATUS", "LAST ACTIVITY", "ACTION")))
	sb.WriteString("\n")

	if len(m.Governance.DormantUsers) == 0 {
		sb.WriteString("  All assigned seats are active! 0 dormant licenses found.\n")
	} else {
		for _, u := range m.Governance.DormantUsers {
			lastAct := "Never"
			if !u.LastActivity.IsZero() {
				lastAct = u.LastActivity.Format("2006-01-02")
			}
			tier := u.Tier
			if u.IsTrial {
				tier += " (Trial)"
			}
			sb.WriteString(fmt.Sprintf("  %-28s %-14s %-10s %-12s %-16s %-12s\n",
				u.UserID, tier, u.Location, string(u.Status), lastAct, "Reclaim Seat"))
		}
	}

	return sb.String()
}

func (m Model) renderUserBreakdownView() string {
	var sb strings.Builder
	if m.SelectedUser == nil {
		sb.WriteString("  No developer selected. Return to [1] Costs or [2] Leaderboard view and press [Enter] on a user row.\n")
		return sb.String()
	}

	u := m.SelectedUser
	sb.WriteString(fmt.Sprintf("DEVELOPER BREAKDOWN: %s\n", u.UserID))
	sb.WriteString(fmt.Sprintf("Status: %s | Total Tokens: %d | WTU: %d | Cache Hit: %.1f%%\n",
		u.SeatStatus, u.TotalTokens, u.WTU, u.CacheHitRate))
	sb.WriteString(fmt.Sprintf("Total Attributed Cost: %s (Included: %s | PAYG: %s | Fair Share PAYG: %s | When Used PAYG: %s)\n\n",
		domain.FormatCost(u.TotalCost, u.Currency),
		domain.FormatCost(u.IncludedCost, u.Currency),
		domain.FormatCost(u.PAYGCost, u.Currency),
		domain.FormatCost(u.FairSharePAYGCost, u.Currency),
		domain.FormatCost(u.WhenUsedPAYGCost, u.Currency)))

	sb.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-24s %-16s %-12s %-12s",
		"MODEL", "TOKENS", "TOKEN SHARE", "COST")))
	sb.WriteString("\n")

	for model, detail := range u.ModelBreakdown {
		sb.WriteString(fmt.Sprintf("  %-22s %-16d %-12s %s\n",
			model, detail.Tokens, fmt.Sprintf("%.1f%%", detail.Share*100), domain.FormatCost(detail.Cost, u.Currency)))
	}

	return sb.String()
}


