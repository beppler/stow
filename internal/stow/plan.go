package stow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Operation describes a planned link from Source to Target.
type Operation struct {
	Source string
	Target string
}

// Conflict describes a target path that cannot be linked.
type Conflict struct {
	Target string
	Reason string
}

// PlanResult contains the planned operations and any conflicts found.
type PlanResult struct {
	Operations []Operation
	Conflicts  []Conflict
}

type planState struct {
	result      PlanResult
	seenTargets map[string]struct{}
}

// Options describes inputs for planning.
type Options struct {
	Dir      string
	Target   string
	Packages []string
}

// PathError carries a path context for errors.
type PathError struct {
	Path string
	Err  error
}

func (e *PathError) Error() string {
	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

func (e *PathError) Unwrap() error {
	return e.Err
}

// DefaultTarget returns the default target directory for a given stow dir.
func DefaultTarget(dir string) (string, error) {
	if dir == "" {
		return "", errors.New("dir is required")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Dir(absDir), nil
}

// BuildPlan validates inputs and returns the planned operations.
func BuildPlan(opts Options) (PlanResult, error) {
	if len(opts.Packages) == 0 {
		return PlanResult{}, errors.New("at least one package is required")
	}

	absDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return PlanResult{}, &PathError{Path: opts.Dir, Err: err}
	}
	info, err := os.Stat(absDir)
	if err != nil {
		return PlanResult{}, &PathError{Path: absDir, Err: err}
	}
	if !info.IsDir() {
		return PlanResult{}, &PathError{Path: absDir, Err: errors.New("dir is not a directory")}
	}

	absTarget, err := filepath.Abs(opts.Target)
	if err != nil {
		return PlanResult{}, &PathError{Path: opts.Target, Err: err}
	}

	packages := append([]string(nil), opts.Packages...)
	sort.Strings(packages)

	state := planState{
		result:      PlanResult{},
		seenTargets: make(map[string]struct{}),
	}
	for _, pkg := range packages {
		pkgPath := filepath.Join(absDir, pkg)
		pkgInfo, err := os.Stat(pkgPath)
		if err != nil {
			return PlanResult{}, &PathError{Path: pkgPath, Err: err}
		}
		if !pkgInfo.IsDir() {
			return PlanResult{}, &PathError{Path: pkgPath, Err: errors.New("package is not a directory")}
		}
		if err := walkPackage(pkgPath, absTarget, &state); err != nil {
			return PlanResult{}, err
		}
	}

	return state.result, nil
}

func walkPackage(pkgPath, targetRoot string, state *planState) error {
	return walkDir(pkgPath, "", targetRoot, state)
}

func walkDir(root, rel, targetRoot string, state *planState) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return &PathError{Path: root, Err: err}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(root, name)
		relPath := filepath.Join(rel, name)

		if isSymlink(entry) {
			if err := handleLeaf(fullPath, relPath, targetRoot, state); err != nil {
				return err
			}
			continue
		}
		if entry.IsDir() {
			if err := walkDir(fullPath, relPath, targetRoot, state); err != nil {
				return err
			}
			continue
		}
		if err := handleLeaf(fullPath, relPath, targetRoot, state); err != nil {
			return err
		}
	}
	return nil
}

func handleLeaf(sourcePath, relPath, targetRoot string, state *planState) error {
	targetPath := filepath.Join(targetRoot, relPath)
	if _, exists := state.seenTargets[targetPath]; exists {
		state.result.Conflicts = append(state.result.Conflicts, Conflict{
			Target: targetPath,
			Reason: "duplicate target planned",
		})
		return nil
	}
	state.seenTargets[targetPath] = struct{}{}
	conflict, conflictReason, isNoOp, err := detectConflict(targetPath, sourcePath)
	if err != nil {
		return err
	}
	if isNoOp {
		return nil
	}
	if conflict {
		state.result.Conflicts = append(state.result.Conflicts, Conflict{
			Target: targetPath,
			Reason: conflictReason,
		})
		return nil
	}
	state.result.Operations = append(state.result.Operations, Operation{
		Source: sourcePath,
		Target: targetPath,
	})
	return nil
}

func detectConflict(targetPath, sourcePath string) (conflict bool, reason string, noOp bool, err error) {
	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", false, nil
		}
		return false, "", false, &PathError{Path: targetPath, Err: err}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		matches, err := symlinkMatches(targetPath, sourcePath)
		if err != nil {
			return false, "", false, &PathError{Path: targetPath, Err: err}
		}
		if matches {
			return false, "", true, nil
		}
		return true, "symlink points elsewhere", false, nil
	}
	return true, "target already exists", false, nil
}

func symlinkMatches(targetPath, sourcePath string) (bool, error) {
	linkTarget, err := os.Readlink(targetPath)
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(linkTarget) {
		linkTarget = filepath.Join(filepath.Dir(targetPath), linkTarget)
	}
	absLinkTarget, err := filepath.Abs(linkTarget)
	if err != nil {
		return false, err
	}
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return false, err
	}
	return filepath.Clean(absLinkTarget) == filepath.Clean(absSource), nil
}

func isSymlink(entry os.DirEntry) bool {
	return entry.Type()&os.ModeSymlink != 0
}
