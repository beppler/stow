# gstow (initial stow clone)

This repository provides a minimal, cross-platform Go implementation of a GNU Stow-compatible `stow` command. The initial scope supports only the `stow` command and a dry-run mode.

## Supported CLI contract (subset)

Usage:

```
stow [flags] <package> [<package> ...]
```

Flags:
- `-n`, `--no`: dry-run; plan and validate but do not change the filesystem.
- `-d`, `--dir`: stow directory (default `.`).
- `-t`, `--target`: target directory. If omitted, the default target is the parent directory of `--dir`.

Behavior:
- Packages are processed in sorted order to guarantee deterministic output.
- Only leaf files are linked. Directories are traversed; symlinked directories are treated as leaf entries (they are not traversed).
- Symlinks inside the package tree are not followed.
- Existing targets that are already the correct symlink are treated as no-ops.
- Conflicts (existing non-matching targets) are reported and skipped; there is no overwrite behavior. Duplicate target paths planned across packages are reported as conflicts with reason `duplicate target planned`.
- Dry-run performs full validation and planning but makes zero filesystem changes (no directory creation, no symlink creation).
- On Windows, creating symlinks may require Developer Mode or elevated privileges; failures are reported as errors.

Output:
- Stdout is reserved for planned/created operations:
  - `LINK <target> -> <source>`
- Stderr is reserved for conflicts and errors:
  - `CONFLICT <target>: <reason>`
  - `ERROR <path>: <message>`

Exit codes:
- `0`: success with no conflicts.
- `1`: conflicts detected.
- `2`: validation or execution error.

## Examples

Dry-run:
```
stow -n -d ./dotfiles -t $HOME vim zsh
```

Real execution:
```
stow -d ./dotfiles -t $HOME vim
```
