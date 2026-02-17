## [1.1.0]

### Added
- **Setup Skills**: New `loto-cli setup-skills` command to install AI skills for Claude Code and other agents
- **Auto-prompt**: On first interactive run, prompts to install AI skills (downloads latest from GitHub)
- **Homebrew**: Simplified formula, skill installation handled by the binary itself

## [1.0.0]

### Added
- Initial release
- View latest extraction results for all games (Loto 6/49, Loto 5/40, Joker, Noroc, Super Noroc)
- View purchased ticket history with win/loss status and prize amounts
- Ticket statistics (total spent, total won, net result, win rate, per-game breakdown)
- Interactive TUI with tabbed interface (Results, Tickets, Stats)
- CLI commands: `results`, `tickets`, `stats`, `config`
- Cookie persistence for faster logins
- AI skill for Claude Code and other agents
