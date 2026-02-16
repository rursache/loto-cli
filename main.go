package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rursache/loto-cli/client"
	"github.com/rursache/loto-cli/config"
	"github.com/rursache/loto-cli/models"
	"github.com/rursache/loto-cli/tui"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		runTUI()
		return
	}

	cmd := args[0]

	switch cmd {
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		fmt.Printf("loto-cli %s\n", version)
	case "results":
		runResults()
	case "tickets":
		withClient(runTickets)
	case "stats":
		withClient(runStats)
	case "config":
		runConfig()
	case "tui":
		runTUI()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`loto-cli - Romanian Lottery CLI

Usage:
  loto-cli [command]

Commands:
  results     Print latest extraction results
  tickets     Print ticket history
  stats       Print ticket statistics
  config      Print config file path
  tui         Start interactive TUI (default when no command)

Options:
  help, -h        Show this help message
  version, -v     Show version

Config:
  Default: ~/.config/loto-cli/config.json

`)
}

func runConfig() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(configPath)
}

func runResults() {
	created, err := config.EnsureExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		os.Exit(1)
	}
	if created {
		configPath, _ := config.GetConfigPath()
		fmt.Fprintf(os.Stderr, "Config file created at: %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Please edit it with your credentials and try again.\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrCredentialsMissing) {
			configPath, _ := config.GetConfigPath()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please edit: %s\n", configPath)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	c, err := client.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	results, err := c.GetResults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching results: %v\n", err)
		os.Exit(1)
	}

	for i, ext := range results {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("=== %s ===\n", ext.Game)
		fmt.Printf("Date: %s\n", ext.Date)
		// Noroc/Super Noroc numbers are individual digits that form a single number
		if ext.Game == models.GameNoroc || ext.Game == models.GameSuperNoroc {
			fmt.Printf("Number: %s\n", formatNorocNumber(ext.Numbers))
		} else {
			fmt.Printf("Numbers: %s\n", formatNumbers(ext.Numbers))
			if len(ext.Bonus) > 0 {
				fmt.Printf("Bonus: %s\n", formatNumbers(ext.Bonus))
			}
		}
	}
}

func runTickets(c *client.Client) {
	tickets, err := c.GetAllTickets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching tickets: %v\n", err)
		os.Exit(1)
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return
	}

	fmt.Printf("%-14s %-12s %-14s %-10s %-12s %s\n", "Game", "Ticket ID", "Draw Date", "Status", "Price", "Prize")
	fmt.Println(strings.Repeat("-", 80))

	for _, t := range tickets {
		prize := "-"
		if t.Prize != "" {
			prize = t.Prize
		}
		fmt.Printf("%-14s %-12s %-14s %-10s %-12s %s\n",
			t.Game,
			t.TicketID,
			t.DrawDate,
			t.Status.String(),
			t.Price,
			prize,
		)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d ticket(s)\n", len(tickets))
}

func runStats(c *client.Client) {
	tickets, err := c.GetAllTickets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching tickets: %v\n", err)
		os.Exit(1)
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return
	}

	totalTickets := len(tickets)
	var totalSpent, totalWon float64
	var wonCount, lostCount, pendingCount int
	gameCount := make(map[models.Game]int)
	gameSpent := make(map[models.Game]float64)
	gameWon := make(map[models.Game]int)
	gameWonAmount := make(map[models.Game]float64)

	for _, t := range tickets {
		price := parsePriceStr(t.Price)
		totalSpent += price
		gameCount[t.Game]++
		gameSpent[t.Game] += price

		switch t.Status {
		case models.StatusWon:
			wonCount++
			gameWon[t.Game]++
			if t.Prize != "" {
				gameWonAmount[t.Game] += parsePriceStr(t.Prize)
				totalWon += parsePriceStr(t.Prize)
			}
		case models.StatusLost:
			lostCount++
		case models.StatusPending:
			pendingCount++
		}
	}

	firstDate := tickets[len(tickets)-1].DrawDate
	lastDate := tickets[0].DrawDate
	avgPrice := totalSpent / float64(totalTickets)
	netResult := totalWon - totalSpent

	winRate := float64(0)
	if decided := wonCount + lostCount; decided > 0 {
		winRate = float64(wonCount) / float64(decided) * 100
	}

	fmt.Println("=== Overview ===")
	fmt.Printf("  Total Tickets:    %d\n", totalTickets)
	fmt.Printf("  Total Spent:      %.2f RON\n", totalSpent)
	fmt.Printf("  Total Won:        %.2f RON\n", totalWon)
	fmt.Printf("  Net Result:       %+.2f RON\n", netResult)
	fmt.Printf("  Avg Ticket Price: %.2f RON\n", avgPrice)
	fmt.Printf("  Date Range:       %s → %s\n", firstDate, lastDate)

	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("  Won:      %d\n", wonCount)
	fmt.Printf("  Lost:     %d\n", lostCount)
	if pendingCount > 0 {
		fmt.Printf("  Pending:  %d\n", pendingCount)
	}
	fmt.Printf("  Win Rate: %.1f%%\n", winRate)

	fmt.Println()
	fmt.Println("=== By Game ===")
	for _, g := range []models.Game{models.GameLoto649, models.GameLoto540, models.GameJoker} {
		count := gameCount[g]
		if count == 0 {
			continue
		}
		fmt.Printf("  %s\n", g)
		fmt.Printf("    Tickets: %d  |  Spent: %.2f RON  |  Won: %d (%.2f RON)\n",
			count, gameSpent[g], gameWon[g], gameWonAmount[g])
	}
}

// parsePriceStr extracts a float from "24,50 RON" format
func parsePriceStr(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, " RON")
	s = strings.TrimSuffix(s, "RON")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

func runTUI() {
	created, err := config.EnsureExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		os.Exit(1)
	}
	if created {
		configPath, _ := config.GetConfigPath()
		fmt.Fprintf(os.Stderr, "Config file created at: %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Please edit it with your credentials and try again.\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrCredentialsMissing) {
			configPath, _ := config.GetConfigPath()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please edit: %s\n", configPath)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	c, err := client.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Logging in to loto.ro...")
	if err := c.Login(); err != nil {
		fmt.Fprintf(os.Stderr, "Login error: %v\n", err)
		os.Exit(1)
	}

	tui.Run(c)
}

// withClient handles config loading, client creation, login, and runs a command
func withClient(fn func(*client.Client)) {
	created, err := config.EnsureExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		os.Exit(1)
	}
	if created {
		configPath, _ := config.GetConfigPath()
		fmt.Fprintf(os.Stderr, "Config file created at: %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Please edit it with your credentials and try again.\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrCredentialsMissing) {
			configPath, _ := config.GetConfigPath()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please edit: %s\n", configPath)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	c, err := client.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Logging in to loto.ro...")
	if err := c.Login(); err != nil {
		fmt.Fprintf(os.Stderr, "Login error: %v\n", err)
		os.Exit(1)
	}

	fn(c)
}

// formatNumbers formats a slice of ints as space-separated strings
func formatNumbers(nums []int) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, " ")
}

// formatNorocNumber joins individual digits into a single number string
func formatNorocNumber(digits []int) string {
	var s strings.Builder
	for _, d := range digits {
		fmt.Fprintf(&s, "%d", d)
	}
	return s.String()
}
