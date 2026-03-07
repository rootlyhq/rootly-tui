# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.3] - 2026-03-06

### Fixed
- Enable editing of API Key and Endpoint text fields on setup panel

### Changed
- Switch all GitHub Actions runners to Blacksmith
- Bump goreleaser/goreleaser-action from 6 to 7

## [0.2.2] - 2026-02-24

### Changed
- Upgrade rootly-go from v0.7.0 to v0.8.0

## [0.2.1] - 2026-02-20

### Changed
- Bump charmbracelet/bubbles from 0.21.1 to 1.0.0
- Upgrade rootly-go from v0.4.0 to v0.7.0
- Bump golang.org/x/text (minor-and-patch group)
- Bump charmbracelet/bubbles and rootly-go dependencies

## [0.2.0] - 2026-01-06

### Added
- Sorting logic for incidents (API-based sorting)
- Extended detail fields for incidents and alerts
- Clickable URLs in label values
- Duration metrics, section icons, and UI improvements
- Split setup screen into two independent panels
- Relative time column to incidents and alerts tables
- i18n-check to lint step
- Custom User-Agent header to API client
- Copy to clipboard shortcut (c) for detail panel
- JSON syntax highlighting for alert data payload
- Pagination metadata display and boundary protection

### Fixed
- Missing newline before labels section in alerts view
- Use correct i18n keys for alerts table columns and tab titles
- Sync i18n locale files and fix broken keys
- Align bullet points in alerts detail view
- Show Source before Status in alerts detail view
- Remove unnecessary nil checks in tests (staticcheck SA5011)
- Disable linux/arm builds and fix LICENSE path in release config
- Disable CGO for cross-platform builds

### Changed
- Refactor i18n to use nested Rails-style YAML for locale files
- Rename LICENSE to LICENSE.txt and update copyright year to 2026
- Change copy shortcut from 'y' to 'c'
- Refactor: extract parseIncidentData to reduce cyclomatic complexity
- Enable CGO for clipboard support
- Improve detail view formatting

## [0.1.0] - 2024-12-31

### Added
- Terminal UI for viewing Rootly incidents and alerts
- Setup screen for API key, endpoint, timezone, and language configuration
- Tabbed navigation between Incidents and Alerts views
- Split-pane layout with list and detail panels
- Keyboard-driven navigation (j/k, Tab, arrows, g/G)
- Pagination support with h/l or [ ] keys
- Detail view with Enter key to fetch extended information
- Open incidents/alerts in browser with 'o' key
- Clickable URLs and emails using OSC 8 hyperlinks
- Markdown rendering for descriptions and summaries
- Scrollable detail panes with focus mode (f key)
- In-app debug logs viewer (l key)
- Help overlay with keyboard shortcuts (? key)
- About screen with version and documentation links (a key)
- Internationalization support for 12 languages
  - English (US/UK), Spanish, French, German, Chinese, Hindi
  - Arabic, Bengali, Portuguese, Russian, Japanese
- Auto-detect system locale from environment variables
- Debug logging with --debug flag or --log file option
- Persistent cache using BoltDB (~/.rootly-tui/cache.db)
- In-memory cache with configurable TTL (5 minutes default)
- Manual refresh with 'r' key to clear cache
- Emoji icons for alert sources (Datadog, PagerDuty, etc.)
- Severity signal bars in incident list
- Pastel color scheme for status indicators
- Arrow indicator for selected rows in table views
- Clipboard support for copying from logs viewer

### Technical
- Built with Bubble Tea TUI framework (Elm Architecture)
- Uses Bubbles for UI components and Lip Gloss for styling
- Uses rootly-go SDK for API communication
- GitHub Actions CI/CD with multi-platform testing
- GoReleaser for binary releases and Homebrew tap
- golangci-lint for code quality
- 80%+ test coverage

[Unreleased]: https://github.com/rootlyhq/rootly-tui/compare/v0.2.3...HEAD
[0.2.3]: https://github.com/rootlyhq/rootly-tui/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/rootlyhq/rootly-tui/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/rootlyhq/rootly-tui/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/rootlyhq/rootly-tui/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/rootlyhq/rootly-tui/releases/tag/v0.1.0
