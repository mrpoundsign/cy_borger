# Project Rules & Guidelines

- **Git Commit Boundary**: Do not commit changes to git unless explicitly prompted by the user.
- **Go Tooling**:
  - Always run `goimports` and `golangci-lint` when saving Go files, and fix any reported issues (note: a pre-commit hook enforces this).
  - Prefer the Go standard library (`net/http`, `html/template`, `database/sql`) over heavy third-party frameworks when possible.
  - Use `go vet` and not `go build` for testing compiling. Users should be running the server under `air` and `go vet` will show you any build errors.
- **Data & Aesthetics**:
  - Maintain CY_BORG visual theme (neon yellow `#ffe600` on dark background `#050505`).
- **Testing & Playwright**:
  - All Playwright tests must be placed in the `tests/` directory to keep the project root clean.
  - Prefer running individual test files (e.g., `npx playwright test tests/test_inject.js`) instead of the full suite unless specifically requested, to save time.
  - Be precise with Playwright selectors and text assertions, especially regarding UI elements with emojis (e.g., `💀 FLATLINE`).
- **Go Handlers & Forms**:
  - Always call `r.ParseForm()` before accessing values via `r.FormValue()` in POST requests.
- **Error Handling**:
  - Never ignore errors using `_ = err` or `_, _ = func()`. Explicitly handle all errors (log them, return HTTP 500s, or handle them gracefully).

- **Frontend & UI Guidelines**:
  - **HTMX over Page Reloads**: Prefer HTMX (`hx-get`, `hx-post`, etc.) for forms and interactive elements to prevent full page reloads and maintain SPA-like performance.
  - **Micro-updates**: Favor swapping specific, minimal DOM elements or using `hx-swap-oob` for targeted updates rather than replacing entire containers (e.g. don't redraw an entire character sheet just to change a status badge).
  - **CSS Consolidation**: Avoid inline `style="..."` attributes. Use utility classes from the central stylesheet (`style.css`) and rely on predefined CSS variables for colors (e.g., `var(--color-danger)`).
  - **Unique Element IDs**: Ensure interactive elements have unique, descriptive IDs containing entity IDs (e.g., `btn-flatline-{{.ID}}-modal`) to differentiate clicks between contexts (like cards vs modals) for Playwright testing.
- **Consistent Terminology**: Use uniform terminology across the UI for actions (e.g., always use "FLATLINE", do not mix "Kill" and "Flatline").

- **Temporary Files & Cleanup**: Always place temporary files (like scratch scripts or data files) in the `tmp/` directory. Delete these files as soon as they are no longer needed to keep the repository clean.
