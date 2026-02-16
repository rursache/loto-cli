loto-cli - Romanian Lottery CLI

Usage:
  loto-cli [command]

Commands:
  results     Print latest extraction results (no auth required)
  tickets     Print ticket history
  stats       Print ticket statistics
  config      Print config file path
  tui         Start interactive TUI (default when no command)

Options:
  help, -h        Show this help message
  version, -v     Show version

Config:
  Default: ~/.config/loto-cli/config.json

  On first run, a config file is created with empty credentials.
  Fill in your bilete.loto.ro email and password to use authenticated commands.

  The "results" command works without credentials.
  The "tickets", "stats", and "tui" commands require authentication.

Examples:
  loto-cli results          # View latest lottery numbers
  loto-cli tickets          # View your ticket history
  loto-cli stats            # View spending and win statistics
  loto-cli                  # Launch interactive TUI
