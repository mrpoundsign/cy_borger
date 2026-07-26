# Project Rules & Guidelines

- **Git Commit Boundary**: Do not commit changes to git unless explicitly prompted by the user.
- **Go Tooling**:
  - Always run `goimports` and `golangci-lint` when saving Go files, and fix any reported issues (note: a pre-commit hook enforces this).
  - Prefer the Go standard library (`net/http`, `html/template`, `database/sql`) over heavy third-party frameworks when possible.
- **Data & Aesthetics**:
  - Maintain CY_BORG visual theme (neon yellow `#ffe600` on dark background `#050505`).
- **Testing & Playwright**:
  - All Playwright tests must be placed in the `tests/` directory to keep the project root clean.
  - Prefer running individual test files (e.g., `npx playwright test tests/test_inject.js`) instead of the full suite unless specifically requested, to save time.
  - When modifying Go files and running tests outside of a watcher like Air, always manually rebuild the application before running Playwright tests to ensure the test environment runs the latest code.
  - Be precise with Playwright selectors and text assertions, especially regarding UI elements with emojis (e.g., `💀 FLATLINE`).
- **Go Handlers & Forms**:
  - Always call `r.ParseForm()` before accessing values via `r.FormValue()` in POST requests.
