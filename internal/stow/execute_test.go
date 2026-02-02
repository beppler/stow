package stow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteCreatesSymlink(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	source := filepath.Join(pkg, "alpha.txt")
	mustWriteFile(t, source)

	if !symlinkSupported(t, stowDir) {
		return
	}

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	if err := Execute(plan, ExecuteOptions{DryRun: false}); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	linked := filepath.Join(targetDir, "alpha.txt")
	if _, err := os.Lstat(linked); err != nil {
		t.Fatalf("expected symlink at %s: %v", linked, err)
	}
	matches, err := symlinkMatches(linked, source)
	if err != nil {
		t.Fatalf("symlinkMatches error: %v", err)
	}
	if !matches {
		t.Fatalf("expected symlink to match source")
	}
}

func TestExecuteDryRunSkipsChanges(t *testing.T) {
	stowDir := t.TempDir()
	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	source := filepath.Join(pkg, "alpha.txt")
	mustWriteFile(t, source)

	targetDir := filepath.Join(t.TempDir(), "target")
	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	if err := Execute(plan, ExecuteOptions{DryRun: true}); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	if _, err := os.Stat(targetDir); err == nil {
		t.Fatalf("expected target directory to not exist in dry-run")
	}
}

func TestExecuteSkipsConflicts(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	source := filepath.Join(pkg, "alpha.txt")
	mustWriteFile(t, source)

	conflictPath := filepath.Join(targetDir, "alpha.txt")
	mustWriteFile(t, conflictPath)

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}
	if len(plan.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(plan.Conflicts))
	}

	if err := Execute(plan, ExecuteOptions{DryRun: false}); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	if _, err := os.Stat(conflictPath); err != nil {
		t.Fatalf("expected conflict path to remain: %v", err)
	}
}

func TestExecuteCreatesParentDirectories(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, filepath.Join(pkg, "nested"))
	source := filepath.Join(pkg, "nested", "alpha.txt")
	mustWriteFile(t, source)

	if !symlinkSupported(t, stowDir) {
		return
	}

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	if err := Execute(plan, ExecuteOptions{DryRun: false}); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	parent := filepath.Join(targetDir, "nested")
	if info, err := os.Stat(parent); err != nil || !info.IsDir() {
		t.Fatalf("expected parent directory to exist: %v", err)
	}
}
