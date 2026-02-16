---
name: loto-cli
description: Monitor and interact with Loteria Romana (loto.ro) via CLI (results, tickets, stats). Use when a user asks to check lottery extraction results, view their ticket history, check ticket statistics (spent, won, win rate), or translate a task into safe loto-cli commands with correct flags, output format, and confirmations.
---

# LOTO-CLI SKILL

## Overview

loto-cli is a terminal tool for interacting with [Loteria Romana](https://loto.ro). It scrapes extraction results from loto.ro and ticket history from bilete.loto.ro. It supports both CLI commands and an interactive TUI mode.

## Installation

```bash
# macOS (Homebrew)
brew install rursache/tap/loto-cli

# Go (all platforms)
go install github.com/rursache/loto-cli@latest

# From source
git clone https://github.com/rursache/loto-cli.git
cd loto-cli && go build -o loto-cli .
```

## Defaults and safety

- Config file: `~/.config/loto-cli/config.json` (created automatically on first run with empty credentials)
- Cookie cache: `~/.config/loto-cli/cookies.json` (session persistence, avoids re-login)
- Config file permissions: `0600` (user-only read/write)
- Credentials are stored in plaintext in the config file — handle with care
- The site requires a Romanian IP address. Non-Romanian IPs will get a clear error message
- Only won tickets trigger detail page fetches (for prize amounts). Lost/pending tickets are not fetched individually

## Quick start

- `loto-cli results`: latest extraction results for all games (no auth required)
- `loto-cli tickets`: purchased ticket history with win/loss status and prize amounts
- `loto-cli stats`: ticket statistics — total spent, total won, net result, win rate, per-game breakdown
- `loto-cli config`: print config file path
- `loto-cli version`: print version
- `loto-cli help`: show usage
- `loto-cli` (no args): launch interactive TUI

## Configuration

On first run, loto-cli creates `~/.config/loto-cli/config.json`:

```json
{
  "email": "your_email@example.com",
  "password": "your_password",
  "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| email | Yes | bilete.loto.ro login email |
| password | Yes | bilete.loto.ro password |
| user_agent | No | Custom HTTP user agent string (defaults to Chrome macOS) |

The `results` command works without credentials. Only `tickets`, `stats`, and `tui` require authentication.

## Commands

### results

Print the latest lottery extraction results for all games. No authentication required.

```bash
loto-cli results
```

Output: One section per game showing game name, draw date, and winning numbers. Games included:
- **Loto 6/49** — 6 numbers
- **Loto 5/40** — 6 numbers
- **Joker** — 5 numbers + 1 bonus number
- **Noroc** — single multi-digit number (tied to 6/49)
- **Super Noroc** — single multi-digit number (tied to 5/40)

Example output:
```
=== Loto 6/49 ===
Date: 15-02-2026
Numbers: 1 23 33 48 2 35

=== Noroc ===
Date: 15-02-2026
Number: 5386535

=== Joker ===
Date: 15-02-2026
Numbers: 26 43 5 18 7
Bonus: 18
```

### tickets

Print purchased ticket history with status and prize amounts. Requires authentication.

```bash
loto-cli tickets
```

Output: Table with columns — Game, Ticket ID, Draw Date, Status (Won/Lost/Pending), Price, Prize. Prize amounts are fetched from ticket detail pages for won tickets only.

Example output:
```
Game           Ticket ID    Draw Date      Status     Price        Prize
--------------------------------------------------------------------------------
Loto 6/49      669235       03.10.2024     Won        21,50 RON    30,00 RON
Loto 6/49      646221       29.09.2024     Lost       21,50 RON    -
```

### stats

Print ticket statistics computed from ticket history. Requires authentication.

```bash
loto-cli stats
```

Output: Three sections:
- **Overview** — total tickets, total spent, total won, net result, average ticket price, date range
- **Results** — won/lost/pending counts, win rate percentage
- **By Game** — per-game breakdown (ticket count, amount spent, wins, amount won)

Example output:
```
=== Overview ===
  Total Tickets:    81
  Total Spent:      2194.50 RON
  Total Won:        922.31 RON
  Net Result:       -1272.19 RON
  Avg Ticket Price: 27.09 RON
  Date Range:       22.08.2024 → 15.02.2026

=== Results ===
  Won:      5
  Lost:     76
  Win Rate: 6.2%

=== By Game ===
  Loto 6/49
    Tickets: 81  |  Spent: 2194.50 RON  |  Won: 5 (922.31 RON)
```

### config

Print the path to the config file.

```bash
loto-cli config
```

Output: The full path, e.g. `/Users/you/.config/loto-cli/config.json`

### version

Print the current version.

```bash
loto-cli version
```

### help

Print usage information.

```bash
loto-cli help
```

## Global options

| Option | Short | Description |
|--------|-------|-------------|
| `help` | `-h` | Show help message |
| `version` | `-v` | Show version |

## Examples

```bash
# Check latest lottery results (no login needed)
loto-cli results

# View all your tickets
loto-cli tickets

# See spending and win statistics
loto-cli stats

# Pipe tickets to find wins
loto-cli tickets | grep "Won"

# Check where config is stored
loto-cli config
```

## Authentication

loto-cli authenticates with bilete.loto.ro using email/password from the config file. The authentication flow:

1. Checks for saved session cookies in `~/.config/loto-cli/cookies.json`
2. If cookies are valid (session not expired), uses them directly
3. If expired, performs a fresh login (GET CSRF token → POST credentials)
4. Saves new session cookies to disk for next run

Session cookies persist between runs, so most invocations don't require a fresh login.

## Troubleshooting

- **"loto.ro requires a Romanian IP address"** — The site geo-blocks non-Romanian IPs. Connect from Romania or use a VPN with a Romanian server.
- **"credentials missing"** — Edit `~/.config/loto-cli/config.json` and fill in your email and password.
- **Login fails** — Verify your credentials work at https://bilete.loto.ro in a browser first.
- **Empty results** — The results page structure may have changed. File an issue at https://github.com/rursache/loto-cli/issues.
