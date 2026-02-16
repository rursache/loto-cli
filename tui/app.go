package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rursache/loto-cli/client"
	"github.com/rursache/loto-cli/models"
)

// tab represents which tab is currently active
type tab int

const (
	tabResults tab = iota
	tabTickets
	tabStats
	tabCount // keep last for modular arithmetic
)

// Messages for async data fetching
type resultsMsg struct {
	results []models.Extraction
	err     error
}

type ticketsMsg struct {
	tickets []models.Ticket
	err     error
}

// model is the main Bubble Tea model
type model struct {
	client *client.Client

	// UI state
	activeTab    tab
	width        int
	height       int
	ready        bool
	viewport     viewport.Model
	spinner      spinner.Model

	// Data
	results       []models.Extraction
	tickets       []models.Ticket
	resultsErr    error
	ticketsErr    error
	loadingResults bool
	loadingTickets bool
}

// Run starts the TUI application
func Run(c *client.Client) error {
	m := newModel(c)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func newModel(c *client.Client) model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(spinnerStyle),
	)

	return model{
		client:         c,
		activeTab:      tabResults,
		spinner:        s,
		loadingResults: true,
		loadingTickets: true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchResults(m.client),
		fetchTickets(m.client),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % tabCount
			m.updateViewportContent()
		case "left", "h":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			m.updateViewportContent()
		case "1":
			m.activeTab = tabResults
			m.updateViewportContent()
		case "2":
			m.activeTab = tabTickets
			m.updateViewportContent()
		case "3":
			m.activeTab = tabStats
			m.updateViewportContent()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerH := 1
		tabBarH := 1
		footerH := 1
		verticalMargin := headerH + tabBarH + footerH + 1 // +1 for spacing

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMargin)
			m.viewport.MouseWheelEnabled = true
			m.viewport.MouseWheelDelta = 3
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargin
			m.updateViewportContent()
		}

	case resultsMsg:
		m.loadingResults = false
		if msg.err != nil {
			m.resultsErr = msg.err
		} else {
			m.results = msg.results
		}
		m.updateViewportContent()

	case ticketsMsg:
		m.loadingTickets = false
		if msg.err != nil {
			m.ticketsErr = msg.err
		} else {
			m.tickets = msg.tickets
		}
		m.updateViewportContent()

	case spinner.TickMsg:
		if m.loadingResults || m.loadingTickets {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update viewport for scrolling
	if m.ready {
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return fmt.Sprintf("\n  %s Initializing...", m.spinner.View())
	}

	header := m.renderHeader()
	tabBar := m.renderTabBar()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		m.viewport.View(),
		footer,
	)
}

// renderHeader renders the top header bar
func (m model) renderHeader() string {
	title := appTitleStyle.Render(" loto-cli ")
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(title)))
	right := lipgloss.NewStyle().Foreground(colorBorder).Render(line)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, right)
}

// renderTabBar renders the tab navigation
func (m model) renderTabBar() string {
	tabs := []struct {
		label string
		id    tab
	}{
		{"Results", tabResults},
		{"Tickets", tabTickets},
		{"Stats", tabStats},
	}

	var rendered []string
	for i, t := range tabs {
		style := inactiveTabStyle
		if m.activeTab == t.id {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(" "+t.label+" "))
		if i < len(tabs)-1 {
			rendered = append(rendered, tabGapStyle.Render(" "))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, rendered...)

	fill := strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)))
	fillStyled := lipgloss.NewStyle().Background(colorBorder).Render(fill)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, fillStyled)
}

// renderFooter renders the bottom keybinding help
func (m model) renderFooter() string {
	keys := []struct{ key, desc string }{
		{"←/→/Tab", "switch tabs"},
		{"↑/↓/j/k", "scroll"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts,
			footerKeyStyle.Render(k.key)+
				footerDescStyle.Render(" "+k.desc),
		)
	}

	help := strings.Join(parts, footerDescStyle.Render("  \u2022  "))
	return footerStyle.Copy().Width(m.width).Render(help)
}

// updateViewportContent sets the viewport content based on the active tab
func (m *model) updateViewportContent() {
	if !m.ready {
		return
	}

	var content string
	switch m.activeTab {
	case tabResults:
		content = m.renderResultsContent()
	case tabTickets:
		content = m.renderTicketsContent()
	case tabStats:
		content = m.renderStatsContent()
	}

	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

// renderResultsContent renders the extraction results for the viewport
func (m model) renderResultsContent() string {
	if m.loadingResults {
		return fmt.Sprintf("\n  %s Loading results...", m.spinner.View())
	}

	if m.resultsErr != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading results: %s", m.resultsErr))
	}

	if len(m.results) == 0 {
		return emptyStyle.Render("No results available.")
	}

	var sections []string
	for _, ext := range m.results {
		section := m.renderExtraction(ext)
		sections = append(sections, section)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderExtraction renders a single game extraction
func (m model) renderExtraction(ext models.Extraction) string {
	color := gameColor(string(ext.Game))

	headerWidth := min(40, m.width-2)

	// Game header with date on the right side, same line
	gameName := gameHeaderStyle.Copy().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(color).
		Render(string(ext.Game))

	date := gameDateStyle.Copy().
		Background(color).
		Foreground(lipgloss.Color("#FFFFFFCC")).
		Render(ext.Date + " ")

	gap := headerWidth - lipgloss.Width(gameName) - lipgloss.Width(date)
	if gap < 1 {
		gap = 1
	}
	filler := lipgloss.NewStyle().Background(color).Render(strings.Repeat(" ", gap))
	headerLine := lipgloss.JoinHorizontal(lipgloss.Center, gameName, filler, date)

	// Numbers (including bonus inline for Joker)
	var balls []string
	for _, n := range ext.Numbers {
		ball := numberBallStyle.Render(fmt.Sprintf("%2d", n))
		balls = append(balls, ball)
	}

	// Append bonus numbers inline with a "+" separator
	if len(ext.Bonus) > 0 {
		balls = append(balls, bonusLabelStyle.Render("+"))
		for _, n := range ext.Bonus {
			ball := bonusBallStyle.Render(fmt.Sprintf("%2d", n))
			balls = append(balls, ball)
		}
	}

	numbersLine := numbersRowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, balls...))

	section := lipgloss.JoinVertical(lipgloss.Left, headerLine, numbersLine)
	return gameSectionStyle.Render(section)
}

// renderTicketsContent renders the ticket list for the viewport
func (m model) renderTicketsContent() string {
	if m.loadingTickets {
		return fmt.Sprintf("\n  %s Loading tickets...", m.spinner.View())
	}

	if m.ticketsErr != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading tickets: %s", m.ticketsErr))
	}

	if len(m.tickets) == 0 {
		return emptyStyle.Render("No tickets found.")
	}

	var cards []string
	for _, t := range m.tickets {
		card := m.renderTicket(t)
		cards = append(cards, card)
	}

	return lipgloss.JoinVertical(lipgloss.Left, cards...)
}

// renderTicket renders a single ticket card
func (m model) renderTicket(t models.Ticket) string {
	color := gameColor(string(t.Game))
	cardWidth := min(m.width-4, 60)

	// Game name with color
	gameName := ticketGameStyle.Copy().
		Foreground(color).
		Render(string(t.Game))

	// Status badge
	badge := renderStatusBadge(t.Status)

	// Top row: game name + status
	topRow := lipgloss.JoinHorizontal(lipgloss.Center, gameName, "  ", badge)

	// Detail rows
	idRow := ticketLabelStyle.Render("Ticket:") + "  " +
		ticketIDStyle.Render(t.TicketID)

	dateRow := ticketLabelStyle.Render("Draw:") + "  " +
		ticketDateStyle.Render(t.DrawDate)

	priceRow := ticketLabelStyle.Render("Price:") + "  " +
		ticketPriceStyle.Render(t.Price)

	rows := []string{topRow, idRow, dateRow, priceRow}

	if t.Prize != "" {
		prizeRow := ticketLabelStyle.Render("Prize:") + "  " +
			lipgloss.NewStyle().Foreground(colorStatusWon).Bold(true).Render(t.Prize)
		rows = append(rows, prizeRow)
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return ticketCardStyle.Copy().
		Width(cardWidth).
		Render(inner)
}

// renderStatusBadge renders a colored status badge
func renderStatusBadge(s models.TicketStatus) string {
	switch s {
	case models.StatusWon:
		return statusStyle(true, false, false).Render(s.String())
	case models.StatusLost:
		return statusStyle(false, true, false).Render(s.String())
	case models.StatusPending:
		return statusStyle(false, false, true).Render(s.String())
	default:
		return statusStyle(false, false, false).Render(s.String())
	}
}

// renderStatsContent renders the statistics tab
func (m model) renderStatsContent() string {
	if m.loadingTickets {
		return fmt.Sprintf("\n  %s Loading ticket data...", m.spinner.View())
	}

	if m.ticketsErr != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading tickets: %s", m.ticketsErr))
	}

	if len(m.tickets) == 0 {
		return emptyStyle.Render("No ticket data available for stats.")
	}

	tickets := m.tickets

	// --- Compute stats ---
	totalTickets := len(tickets)
	var totalSpent, totalWon float64
	var wonCount, lostCount, pendingCount int
	gameCount := make(map[models.Game]int)
	gameSpent := make(map[models.Game]float64)
	gameWon := make(map[models.Game]int)
	gameWonAmount := make(map[models.Game]float64)

	for _, t := range tickets {
		price := parsePrice(t.Price)
		totalSpent += price
		gameCount[t.Game]++
		gameSpent[t.Game] += price

		switch t.Status {
		case models.StatusWon:
			wonCount++
			gameWon[t.Game]++
			if t.Prize != "" {
				prizeVal := parsePrice(t.Prize)
				totalWon += prizeVal
				gameWonAmount[t.Game] += prizeVal
			}
		case models.StatusLost:
			lostCount++
		case models.StatusPending:
			pendingCount++
		}
	}

	// Date range
	firstDate := tickets[len(tickets)-1].DrawDate
	lastDate := tickets[0].DrawDate

	winRate := float64(0)
	decided := wonCount + lostCount
	if decided > 0 {
		winRate = float64(wonCount) / float64(decided) * 100
	}

	avgPrice := totalSpent / float64(totalTickets)
	netResult := totalWon - totalSpent

	cardWidth := min(m.width-4, 60)

	// --- Render ---
	var sections []string

	// Overview card
	overviewHeader := statsSectionHeader.Copy().Width(cardWidth).Render("Overview")

	// Format net result with color
	netLabel := fmt.Sprintf("%.2f RON", netResult)
	var netRendered string
	if netResult >= 0 {
		netRendered = lipgloss.NewStyle().Foreground(colorStatusWon).Bold(true).Render("+" + netLabel)
	} else {
		netRendered = lipgloss.NewStyle().Foreground(colorStatusLost).Bold(true).Render(netLabel)
	}

	overviewRows := []string{
		statsRow("Total Tickets", fmt.Sprintf("%d", totalTickets)),
		statsRow("Total Spent", fmt.Sprintf("%.2f RON", totalSpent)),
		statsRow("Total Won", fmt.Sprintf("%.2f RON", totalWon)),
		statsRow("Net Result", netRendered),
		statsRow("Avg Ticket Price", fmt.Sprintf("%.2f RON", avgPrice)),
		statsRow("Date Range", fmt.Sprintf("%s → %s", firstDate, lastDate)),
	}
	overviewCard := statsCardStyle.Copy().Width(cardWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, append([]string{overviewHeader}, overviewRows...)...),
	)
	sections = append(sections, overviewCard)

	// Win/Loss card
	wlHeader := statsSectionHeader.Copy().Width(cardWidth).Render("Results")
	wlRows := []string{
		statsRow("Won", statusStyle(true, false, false).Render(fmt.Sprintf(" %d ", wonCount))),
		statsRow("Lost", statusStyle(false, true, false).Render(fmt.Sprintf(" %d ", lostCount))),
	}
	if pendingCount > 0 {
		wlRows = append(wlRows, statsRow("Pending", statusStyle(false, false, true).Render(fmt.Sprintf(" %d ", pendingCount))))
	}
	wlRows = append(wlRows, statsRow("Win Rate", fmt.Sprintf("%.1f%%", winRate)))
	wlCard := statsCardStyle.Copy().Width(cardWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, append([]string{wlHeader}, wlRows...)...),
	)
	sections = append(sections, wlCard)

	// Per-game breakdown
	games := []models.Game{models.GameLoto649, models.GameLoto540, models.GameJoker}
	var gameRows []string
	bgHeader := statsSectionHeader.Copy().Width(cardWidth).Render("By Game")
	for _, g := range games {
		count := gameCount[g]
		if count == 0 {
			continue
		}
		spent := gameSpent[g]
		won := gameWon[g]
		wonAmt := gameWonAmount[g]
		color := gameColor(string(g))
		name := lipgloss.NewStyle().Foreground(color).Bold(true).Render(string(g))
		detail := fmt.Sprintf("%d tickets  •  %.2f RON spent  •  %d won (%.2f RON)", count, spent, won, wonAmt)
		gameRows = append(gameRows, name+"\n"+statsValueStyle.Render(detail))
	}
	if len(gameRows) > 0 {
		bgCard := statsCardStyle.Copy().Width(cardWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left, append([]string{bgHeader}, gameRows...)...),
		)
		sections = append(sections, bgCard)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// statsRow renders a label-value row for the stats tab
func statsRow(label, value string) string {
	return statsLabelStyle.Render(label) + "  " + statsValueStyle.Render(value)
}

// parsePrice extracts a float from "24,50 RON" format
func parsePrice(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, " RON")
	s = strings.TrimSuffix(s, "RON")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

// Async data fetching commands

func fetchResults(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		results, err := c.GetResults()
		return resultsMsg{results: results, err: err}
	}
}

func fetchTickets(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		tickets, err := c.GetAllTickets()
		return ticketsMsg{tickets: tickets, err: err}
	}
}

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
