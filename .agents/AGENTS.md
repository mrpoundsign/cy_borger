# Project Rules & Guidelines

- **Git Commit Boundary**: Do not commit changes to git unless explicitly prompted by the user.
- **Go Tooling**:
  - Always run `goimports` (or `go fmt`) and `golangci-lint` when saving Go files, and fix any reported issues.
  - Prefer the Go standard library (`net/http`, `html/template`, `database/sql`) over heavy third-party frameworks when possible.
- **Data & Aesthetics**:
  - Maintain CY_BORG visual theme (neon yellow `#ffe600` on dark background `#050505`).
