package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	colorPrimary   = lipgloss.Color("#2ECC71") // green - lottery theme
	colorSecondary = lipgloss.Color("#27AE60") // darker green
	colorAccent    = lipgloss.Color("#F1C40F") // gold/yellow
	colorBg        = lipgloss.Color("#1A1A2E") // dark background
	colorBgLight   = lipgloss.Color("#16213E") // slightly lighter bg
	colorText      = lipgloss.Color("#ECF0F1") // light text
	colorTextDim   = lipgloss.Color("#7F8C8D") // dimmed text
	colorBorder    = lipgloss.Color("#34495E") // border color

	// Status colors
	colorStatusWon     = lipgloss.Color("#2ECC71") // green
	colorStatusLost    = lipgloss.Color("#E74C3C") // red
	colorStatusPending = lipgloss.Color("#F39C12") // orange/yellow
	colorStatusUnknown = lipgloss.Color("#95A5A6") // gray

	// Game header colors
	colorLoto649   = lipgloss.Color("#E74C3C") // red
	colorLoto540   = lipgloss.Color("#3498DB") // blue
	colorJoker     = lipgloss.Color("#9B59B6") // purple
	colorNoroc     = lipgloss.Color("#F39C12") // orange
	colorSuperNoroc = lipgloss.Color("#1ABC9C") // teal

	// Number ball colors
	colorBall      = lipgloss.Color("#2C3E50") // dark ball background
	colorBallText  = lipgloss.Color("#FFFFFF") // white ball text
	colorBonusBall = lipgloss.Color("#E74C3C") // red for bonus
)

// Header styles
var (
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Background(colorPrimary).
		Padding(0, 1).
		Align(lipgloss.Center)

	appTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colorPrimary).
		Padding(0, 1)
)

// Tab styles
var (
	activeTabStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colorSecondary).
		Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
		Foreground(colorTextDim).
		Background(colorBorder).
		Padding(0, 2)

	tabGapStyle = lipgloss.NewStyle().
		Background(colorBorder).
		Padding(0, 0)
)

// Footer style
var (
	footerStyle = lipgloss.NewStyle().
		Foreground(colorTextDim).
		Padding(0, 1)

	footerKeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent)

	footerDescStyle = lipgloss.NewStyle().
		Foreground(colorTextDim)
)

// Game section styles
var (
	gameSectionStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)

	gameHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginBottom(0)

	gameDateStyle = lipgloss.NewStyle().
		Foreground(colorTextDim).
		Italic(true).
		PaddingLeft(2)

	numbersRowStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingTop(0)

	numberBallStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBallText).
		Background(colorBall).
		Padding(0, 1).
		MarginRight(1).
		Align(lipgloss.Center)

	bonusBallStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBallText).
		Background(colorBonusBall).
		Padding(0, 1).
		MarginRight(1).
		Align(lipgloss.Center)

	bonusLabelStyle = lipgloss.NewStyle().
		Foreground(colorBonusBall).
		Bold(true).
		PaddingLeft(1).
		MarginRight(1)
)

// Ticket styles
var (
	ticketCardStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		MarginBottom(1)

	ticketGameStyle = lipgloss.NewStyle().
		Bold(true).
		MarginRight(1)

	ticketIDStyle = lipgloss.NewStyle().
		Foreground(colorTextDim)

	ticketDateStyle = lipgloss.NewStyle().
		Foreground(colorText)

	ticketPriceStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	ticketLabelStyle = lipgloss.NewStyle().
		Foreground(colorTextDim).
		Width(10)
)

// Status badge styles
func statusStyle(won bool, lost bool, pending bool) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	switch {
	case won:
		return base.Foreground(lipgloss.Color("#FFFFFF")).Background(colorStatusWon)
	case lost:
		return base.Foreground(lipgloss.Color("#FFFFFF")).Background(colorStatusLost)
	case pending:
		return base.Foreground(lipgloss.Color("#000000")).Background(colorStatusPending)
	default:
		return base.Foreground(lipgloss.Color("#FFFFFF")).Background(colorStatusUnknown)
	}
}

// Loading/spinner style
var (
	spinnerStyle = lipgloss.NewStyle().
		Foreground(colorPrimary)

	loadingTextStyle = lipgloss.NewStyle().
		Foreground(colorTextDim).
		PaddingLeft(1)
)

// Error style
var errorStyle = lipgloss.NewStyle().
	Foreground(colorStatusLost).
	Bold(true).
	Padding(1, 2)

// Empty state style
var emptyStyle = lipgloss.NewStyle().
	Foreground(colorTextDim).
	Italic(true).
	Padding(1, 2)

// Separator
var separatorStyle = lipgloss.NewStyle().
	Foreground(colorBorder)

// Stats styles
var (
	statsCardStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		MarginBottom(1)

	statsSectionHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorBorder)

	statsLabelStyle = lipgloss.NewStyle().
		Foreground(colorTextDim).
		Width(18)

	statsValueStyle = lipgloss.NewStyle().
		Foreground(colorText)
)

// gameColor returns the appropriate color for a game type
func gameColor(game string) lipgloss.Color {
	switch game {
	case "Loto 6/49":
		return colorLoto649
	case "Loto 5/40":
		return colorLoto540
	case "Joker":
		return colorJoker
	case "Noroc":
		return colorNoroc
	case "Super Noroc":
		return colorSuperNoroc
	default:
		return colorPrimary
	}
}
