# loto-cli

## Project Overview

loto-cli is a Go CLI and TUI application for interacting with [Loteria Romana](https://loto.ro), Romania's national lottery. It scrapes lottery extraction results from loto.ro and ticket purchase history from bilete.loto.ro. It supports both direct CLI commands (results, tickets, stats) and an interactive terminal UI built with Bubble Tea.

Repository: `github.com/rursache/loto-cli`

## Codebase Structure

```
.
├── main.go              # Entry point, CLI command routing, withClient helper, output formatting
├── skills.go            # AI skill installation logic (setup-skills command, auto-prompt)
├── config/
│   └── config.go        # Config loading/creation at ~/.config/loto-cli/config.json
├── client/
│   ├── client.go        # HTTP client with cookie jar, browser headers, geo-blocking detection
│   ├── auth.go          # Login flow: cookie restore, CSRF token fetch, POST credentials, session check
│   ├── results.go       # Scrapes extraction results from loto.ro homepage using goquery
│   ├── tickets.go       # Scrapes ticket history from bilete.loto.ro with pagination and prize fetching
│   └── cookies.go       # Cookie persistence: save/load/clear cookies to JSON on disk
├── models/
│   └── models.go        # Data types: Game, Extraction, Ticket, TicketStatus, GameFromImagePath
├── tui/
│   ├── app.go           # Bubble Tea model: tabs (Results, Tickets, Stats), async data loading, rendering
│   └── styles.go        # Lipgloss styles: colors, tab bar, game headers, number balls, ticket cards, stats
├── skill/
│   ├── SKILL.md         # AI skill definition file (downloaded by setup-skills)
│   └── references/
│       └── help-man-page.md
├── .github/
│   └── workflows/
│       └── trigger-tap-update.yml  # GitHub Actions workflow for Homebrew tap updates
├── go.mod               # Go module: go 1.25.0
└── go.sum
```

### Package Responsibilities

- **main** (root) -- CLI command dispatch. Handles `results`, `tickets`, `stats`, `config`, `setup-skills`, `tui`, `help`, `version`. The `withClient` helper handles config loading, client creation, and login for authenticated commands. Also contains `formatNumbers`, `formatNorocNumber`, and `parsePriceStr` utilities.
- **config** -- Reads and writes `~/.config/loto-cli/config.json`. Creates a prefilled template on first run. Validates that email and password are set. Provides a default User-Agent string.
- **client** -- HTTP interactions with loto.ro and bilete.loto.ro. All requests go through `newRequest` (sets browser headers) and `doRequest` (checks for 410 geo-blocking). Uses `net/http/cookiejar` for session management.
- **models** -- Pure data types with no dependencies. Defines `Game` constants (Loto 6/49, Loto 5/40, Joker, Noroc, Super Noroc), `Extraction`, `Ticket`, `TicketStatus` (Won/Lost/Pending), and `GameFromImagePath` for parsing image URLs.
- **tui** -- Interactive terminal UI using Bubble Tea and Lipgloss. Three tabs: Results, Tickets, Stats. Fetches data asynchronously on startup. Keyboard navigation: arrow keys, Tab, 1/2/3 for tabs; j/k for scrolling; q to quit.

## Key Concepts

### Authentication (Laravel CSRF Flow)

bilete.loto.ro runs Laravel. The login flow in `client/auth.go`:

1. Attempt to restore session from saved cookies (`LoadCookies`).
2. Verify session validity by requesting the ticket history page and checking if the page title contains "Biletele Mele". A 302 redirect to /login means the session expired.
3. If no valid session: GET the login page, extract the CSRF token from `<meta name="csrf-token">`.
4. POST to /login with `_token`, `email`, `password` as form-encoded data.
5. A 302 response indicates successful login (Laravel redirect). The cookie jar captures session cookies automatically.
6. Save cookies to disk for subsequent runs.

### HTML Scraping with goquery

All data is obtained by scraping HTML pages (there is no API):

- **Results** (`client/results.go`): Scrapes the loto.ro homepage. Finds the results section by a specific CSS class (`vc_custom_1643109784313`), then parses three columns (one per game pair). Each column contains Ninja Tables (footable) with draw numbers, dates, and Noroc numbers. Hidden tables (parent has "ascuns" class) are skipped.
- **Tickets** (`client/tickets.go`): Scrapes bilete.loto.ro ticket history pages. Parses `div.ticket-preview` cards for game type, ticket ID, draw date, status, and price. Pagination is handled by parsing "Showing X to Y of Z results" text. Won tickets get an additional detail page fetch to extract prize amounts from `<tfoot>` rows.

### Cookie Persistence

Session cookies are serialized to `~/.config/loto-cli/cookies.json` with 0600 permissions. On load, expired cookies are filtered out. Corrupted cookie files are silently ignored.

### Geo-Blocking

loto.ro returns HTTP 410 (Gone) for non-Romanian IP addresses. The `doRequest` method in `client/client.go` checks for this status and returns a clear error message suggesting a VPN.

### Browser Headers

All requests include browser-like headers (User-Agent, Accept, Accept-Language, Cache-Control, DNT, Pragma) to avoid being blocked. The User-Agent is configurable in the config file and defaults to a Chrome on macOS string.

## Build and Run

This is a Go project. Do NOT use Xcode or Swift tooling.

```bash
# Build
go build -o loto-cli .

# Run (interactive TUI)
./loto-cli

# Run specific commands
./loto-cli results
./loto-cli tickets
./loto-cli stats
./loto-cli config
./loto-cli setup-skills

# Install from source
go install github.com/rursache/loto-cli@latest
```

### Testing

There are currently no automated tests in this project. To verify a build works, compile and run `loto-cli results` (does not require authentication but does require a Romanian IP) or `loto-cli help`.

### Version

The `version` variable in `main.go` defaults to `"dev"` and is typically overridden at build time via `-ldflags`:

```bash
go build -ldflags "-X main.version=1.1.0" -o loto-cli .
```

## Release Process

Step by step:

1. Update `CHANGELOG.md` with the new version and changes.
2. Commit and push to the `master` branch.
3. Create a GitHub release with tag `vX.Y.Z` (e.g., `v1.1.0`). The release notes should match the CHANGELOG entry for that version.
4. When the release is published, the `trigger-tap-update.yml` GitHub Actions workflow fires automatically.
5. That workflow dispatches to the `update-formula.yml` workflow in the `rursache/homebrew-tap` repository, passing the formula name (`loto-cli`), the tag, and the repository.
6. The homebrew-tap workflow downloads the source tarball from the GitHub release, computes the SHA256 hash, updates the Homebrew formula file, and commits and pushes the changes.
7. Users receive the update via `brew update && brew upgrade loto-cli`.

## GitHub Actions

### trigger-tap-update.yml

Located at `.github/workflows/trigger-tap-update.yml`. This workflow:

- Triggers on: GitHub release `published` event, or manual `workflow_dispatch` with a `tag` input.
- Uses the `TAP_GITHUB_TOKEN` secret (a PAT with access to the `rursache/homebrew-tap` repo).
- Sends a `workflow_dispatch` event to `rursache/homebrew-tap` to run `update-formula.yml` on the `master` branch.
- Passes three inputs: `formula` ("loto-cli"), `tag` (the release tag), and `repository` (the source repo).

This is the only CI workflow in the project. There are no build or test workflows.

## AI Skills

The `setup-skills` command and the auto-prompt mechanism allow AI assistants (Claude Code, etc.) to learn how to use loto-cli.

### How It Works

- On first interactive TUI launch, `maybePromptSkillInstall()` checks if skills are already installed and if the user has been prompted before.
- If not, it asks `Install AI skills for Claude Code and other agents? [y/N]`.
- The prompt state is tracked by a `.skill-prompted` flag file in the config directory.
- Skills can also be installed explicitly with `loto-cli setup-skills`.

### What Gets Installed

The `downloadAndInstallSkills()` function downloads files from the `skill/` directory in the GitHub repo (raw.githubusercontent.com, master branch):

- `SKILL.md` -- The skill definition with command documentation.
- `references/help-man-page.md` -- Reference material.

These are installed to two locations:
- `~/.agents/skills/loto-cli/` -- For generic agent tools.
- `~/.claude/skills/loto-cli/` -- For Claude Code specifically.

## Configuration

### Config File

Path: `~/.config/loto-cli/config.json`

Created automatically on first run with empty credentials. File permissions are 0600.

```json
{
  "email": "",
  "password": "",
  "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"
}
```

- `email` (required for tickets/stats/tui) -- bilete.loto.ro login email.
- `password` (required for tickets/stats/tui) -- bilete.loto.ro login password.
- `user_agent` (optional) -- Custom HTTP User-Agent string. Defaults to Chrome on macOS.

The `results` command works without credentials.

### Cookies File

Path: `~/.config/loto-cli/cookies.json`

Stores session cookies from bilete.loto.ro in JSON format. File permissions are 0600. Expired cookies are automatically filtered out on load. This file is managed automatically and should not need manual editing.

### Skill Prompt Flag

Path: `~/.config/loto-cli/.skill-prompted`

A marker file that prevents re-prompting the user about AI skill installation.

## Dependencies

Key Go dependencies (see `go.mod`):

- `github.com/PuerkitoBio/goquery` -- HTML parsing and CSS selector queries (jQuery-like).
- `github.com/charmbracelet/bubbletea` -- TUI framework (Elm architecture).
- `github.com/charmbracelet/bubbles` -- TUI components (spinner, viewport).
- `github.com/charmbracelet/lipgloss` -- Terminal styling and layout.

## Common Patterns

- All authenticated commands use the `withClient` helper in `main.go`, which handles config loading, client creation, login, and error reporting.
- All HTTP requests go through `client.newRequest` (adds headers) and `client.doRequest` (checks geo-blocking).
- Romanian text appears in HTML parsing: "Biletele Mele" (page title check), "Necastigator"/"Castigator" (ticket status), "TOTAL CASTIG" (prize amounts), "Tragerea"/"Stare Bilet"/"ID Bilet" (ticket fields).
- Prices use Romanian format: `24,50 RON` (comma as decimal separator). The `parsePriceStr`/`parsePrice` functions handle conversion to float64.
- The TUI fetches results and tickets concurrently on startup using Bubble Tea commands.
