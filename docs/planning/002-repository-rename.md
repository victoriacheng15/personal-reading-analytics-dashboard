# Planning: Repository and Package Rename

## Objective

Rename the repository from `personal-reading-analytics` to `personal-reading-analytics`. Additionally, rename the `cmd/dashboard` directory to `cmd/analytics` to better reflect its function and align with the "Analytics" page name.

## Rationale

- **Architectural Clarity:** Aligns the project's public and internal names with its true scope as a data platform.
- **Consistency:** Ensures naming is consistent from the repository level down to the application binaries.

## Impact Analysis

The following files require updates to the new repository and directory names:

### 1. Repository Name (`personal-reading-analytics`)

- **Go Module:** `go.mod` must be updated.
- **Go Imports:** All `import` statements across `cmd/**/*.go` must be updated.
- **Documentation:** `README.md` and `cmd/internal/dashboard/templates/index.html` contain links to the GitHub repository.

### 2. Directory Name (`cmd/analytics`)

- **Directory:** `cmd/dashboard/` will be renamed to `cmd/analytics/`.
- **Makefile:** The build command for the dashboard binary must be updated from `./cmd/dashboard` to `./cmd/analytics`.
- **Documentation:** `AGENTS.md`, `docs/architecture/dashboard.md`, and `docs/architecture/schemas.md` reference the old path.

## Implementation Plan

### 1. File System & Content Updates

1. **Rename Directory:** Rename `cmd/dashboard` to `cmd/analytics`.
2. **Global Replace (Repo Name):** Find `personal-reading-analytics` and replace with `personal-reading-analytics`.
3. **Targeted Replace (Directory Name):** Find `cmd/dashboard` and replace with `cmd/analytics` in the `Makefile` and documentation.

### 2. Go Dependency & Build Refresh

1. Run `go mod tidy` to synchronize the `go.mod` and `go.sum` files.
2. Run `nix-shell --run "make go-test"` to confirm all tests pass with the new import paths.
3. Run `nix-shell --run "make run-dashboard"` (which will use the updated `cmd/analytics` path) to verify the application builds and runs successfully.

## Next Steps

- Approval of this plan.
- Create an issue in `blog-drafts` repo: `[personal-reading-analytics] - Repository and Package Rename`.
