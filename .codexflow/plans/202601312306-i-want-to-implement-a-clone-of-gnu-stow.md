# Plan: Initial Go Stow Clone (stow + dry-run)

## Goal
- Define a minimal, cross-platform (Linux/macOS/Windows) Go CLI that implements a GNU Stow-compatible `stow` command and supports a dry-run flag without changing the filesystem.
- Acceptance criteria:
  - CLI contract is documented and enforced for the supported subset (positional package names; `-n/--no` for dry-run; `-d/--dir`; `-t/--target`), including which messages go to stdout vs stderr.
  - Default target behavior is explicit and tested: if `--target` is omitted, the target is the parent directory of `--dir`.
  - Planning produces a deterministic, stable-ordered list of intended link operations (leaf files only), consistent across multiple packages and platforms.
  - Dry-run performs full validation/planning but makes zero filesystem changes.
  - Real mode creates any required parent directories, then symlinks for planned operations; conflicts are detected, reported, and skipped (no overwrite).
  - When an intended symlink already exists and points to the correct source, the operation is treated as a no-op (not a conflict).
  - Exit codes are documented and exercised for: success with no conflicts, conflicts present, and validation errors (exact numeric codes are defined in the plan and README).
  - Any problem is reported on stderr and includes the target path plus the error message; dry-run output remains stable for tests.
  - Tests cover path mapping, traversal order, conflict detection, existing-symlink no-op, dry-run behavior, target defaulting, and CLI output for core cases.

## Non-goals
- Implement `unstow`, `restow`, `--adopt`, or other GNU Stow features beyond the minimal `stow` + dry-run.
- Guarantee complete behavioral parity with GNU Stow in this first step.
- Add advanced conflict resolution, package priority, or ignore file parsing unless required by clarified specs.

## Constraints / assumptions
- Repository currently has no source files; we will create a new Go module (github.com/beppler/gstow) and minimal CLI layout.
- Dry-run must perform the same traversal and decision logic as real mode but must not modify the filesystem (including no directory creation).
- Cross-platform support requires path handling via `path/filepath` and avoiding POSIX-only assumptions.
- Windows symlink creation may require elevated privileges or Developer Mode; the implementation must surface clear errors.
- Deterministic traversal/order is required to keep tests stable.
- Only leaf files are symlinked; directory symlinks are not created.
- Conflicts are skipped (not overwritten) and reported.
- Planning must not follow symlinks inside the package tree.
- If `--target` is omitted, default target is the parent of `--dir`.
- Stdout is reserved for planned/created operations; stderr is reserved for errors and conflict reports.

## Proposed approach
- Define the supported CLI contract (flags, positional arguments, defaults, exit codes, and output format) and document it.
- Establish a Go module and a CLI entry point (`cmd/stow/main.go`) with GNU Stow-compatible flags (`-n/--no`, `-d/--dir`, `-t/--target`).
- Implement a core package that:
  - Validates inputs (package name list, source dir, target dir) and resolves absolute paths.
  - Walks each package directory with deterministic ordering (sorted entries), computes target link paths, and records planned operations.
  - Detects conflicts (target exists and is not the intended symlink), and treats matching symlinks as no-ops.
  - Detects duplicate target paths across packages during planning reporting them as conflicts.
  - Applies operations (parent directory creation + symlink creation) only when not in dry-run mode.
- Add unit and CLI tests using temporary directories to verify correctness and output, with Windows-aware expectations.

## Risks / edge cases
- Windows symlink permissions can cause failures; tests must avoid assuming symlink creation works without Developer Mode.
- Existing target paths that are directories, regular files, or symlinks to different destinations must be treated as conflicts.
- Package directories containing nested paths and dotfiles should map correctly without skipping entries unless explicitly specified.
- Path separator differences and relative vs absolute path handling can cause incorrect target computations.
- Walking directories must avoid following symlinks inside the package to prevent unintended recursion.
- GNU Stow output formatting is broader than this subset; tests must lock in only the supported messages.
- Handling multiple package arguments and overlap conflicts across packages needs deterministic ordering.
- Creating parent directories in real mode must not mask conflicts at the final target path.
- Exit code definitions must be consistent across CLI and internal execution paths.

## Test gaps to watch for
- Behavior when the package directory does not exist or is empty.
- Idempotency when the intended symlink already exists and points to the correct source.
- Dry-run output consistency when conflicts are present.
- Windows-specific behavior when `os.Symlink` is unavailable; tests should skip or assert error messages appropriately.
- Exit code behavior when conflicts occur alongside successful links.
- Multiple package inputs and conflict ordering across packages.
- Parent directory creation in real mode vs no changes in dry-run mode.
- Default target behavior when `--target` is omitted.
- Stdout/stderr separation for success vs error/conflict messages.

## Files to change
- `go.mod` (initialize module for the CLI and packages).
- `cmd/stow/main.go` (CLI entry point and flag parsing).
- `cmd/stow/main_test.go` (CLI output and exit code tests).
- `internal/stow/plan.go` (planning logic and data structures for operations).
- `internal/stow/execute.go` (apply planned operations; respects dry-run).
- `internal/stow/plan_test.go` (unit tests for planning and dry-run behavior).
- `internal/stow/execute_test.go` (unit tests for execution vs dry-run behavior).
- `README.md` (minimal usage documentation for `stow` and dry-run).
- `.codexflow/reviews/202601312306-i-want-to-implement-a-clone-of-gnu-stow.review.md` (update review findings after fixes).

## Step-by-step checklist
- [x] Define the CLI contract for this subset (positional packages, `-n/--no`, `-d/--dir`, `-t/--target`, default target behavior, stdout/stderr rules, numeric exit codes, and output expectations) and document it in `README.md`.
- [x] Create `go.mod` plus a basic CLI skeleton in `cmd/stow/main.go` with GNU Stow-compatible flag parsing and error handling.
- [x] Implement planning logic in `internal/stow/plan.go`: validate inputs, traverse each package dir in deterministic order, compute target paths, detect conflicts, treat matching symlinks as no-ops, and avoid following symlinks.
- [x] Implement execution logic in `internal/stow/execute.go` that creates required parent directories, applies the plan, skips conflicts, and skips changes in dry-run mode.
- [x] Add unit tests in `internal/stow/plan_test.go` for path mapping, traversal order, conflict detection, existing-symlink no-op, empty or missing package handling, multiple packages, dry-run planning behavior, and default target computation.
- [x] Add unit tests in `internal/stow/execute_test.go` to verify execution vs dry-run behavior (including parent directory creation), conflict skipping, and Windows-aware expectations.
- [x] Add CLI tests in `cmd/stow/main_test.go` for output formatting and exit codes in dry-run, success, and conflict scenarios.
- [x] Add minimal `README.md` usage examples covering real and dry-run execution and the CLI contract and duplicate target conflicts reason and Windows requisites related to symlinks.
- [x] Update the review report with resolved findings after applying fixes.

## Tests to run
- `go test ./...`

## Rollback plan
- Revert the newly created Go module and source files (`go.mod`, `cmd/`, `internal/`, `README.md`) to return the repo to its original state.

## Open questions
- None.
