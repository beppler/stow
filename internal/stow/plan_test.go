package stow

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildPlanMappingAndOrder(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := filepath.Join(t.TempDir(), "target")

	pkgA := filepath.Join(stowDir, "pkg-a")
	pkgB := filepath.Join(stowDir, "pkg-b")
	mustMkdir(t, pkgA)
	mustMkdir(t, pkgB)

	mustWriteFile(t, filepath.Join(pkgA, "alpha.txt"))
	mustMkdir(t, filepath.Join(pkgA, "dir"))
	mustWriteFile(t, filepath.Join(pkgA, "dir", "bravo.txt"))
	mustWriteFile(t, filepath.Join(pkgA, "zeta.txt"))
	mustWriteFile(t, filepath.Join(pkgB, "echo.txt"))

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg-b", "pkg-a"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	stowDirAbs, _ := filepath.Abs(stowDir)
	targetAbs, _ := filepath.Abs(targetDir)

	expected := []Operation{
		{Source: filepath.Join(stowDirAbs, "pkg-a", "alpha.txt"), Target: filepath.Join(targetAbs, "alpha.txt")},
		{Source: filepath.Join(stowDirAbs, "pkg-a", "dir", "bravo.txt"), Target: filepath.Join(targetAbs, "dir", "bravo.txt")},
		{Source: filepath.Join(stowDirAbs, "pkg-a", "zeta.txt"), Target: filepath.Join(targetAbs, "zeta.txt")},
		{Source: filepath.Join(stowDirAbs, "pkg-b", "echo.txt"), Target: filepath.Join(targetAbs, "echo.txt")},
	}

	if len(plan.Operations) != len(expected) {
		t.Fatalf("expected %d operations, got %d", len(expected), len(plan.Operations))
	}
	for i, op := range plan.Operations {
		if op != expected[i] {
			t.Fatalf("operation %d mismatch: got %+v, want %+v", i, op, expected[i])
		}
	}
}

func TestBuildPlanConflictDetection(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	mustWriteFile(t, filepath.Join(pkg, "alpha.txt"))

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
	if plan.Conflicts[0].Target != conflictPath {
		t.Fatalf("unexpected conflict target: %s", plan.Conflicts[0].Target)
	}
	if len(plan.Operations) != 0 {
		t.Fatalf("expected no operations due to conflict, got %d", len(plan.Operations))
	}
}

func TestBuildPlanDuplicateTargets(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkgA := filepath.Join(stowDir, "pkg-a")
	pkgB := filepath.Join(stowDir, "pkg-b")
	mustMkdir(t, pkgA)
	mustMkdir(t, pkgB)

	mustWriteFile(t, filepath.Join(pkgA, "alpha.txt"))
	mustWriteFile(t, filepath.Join(pkgB, "alpha.txt"))

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg-b", "pkg-a"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	if len(plan.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(plan.Operations))
	}
	if len(plan.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(plan.Conflicts))
	}

	stowDirAbs, _ := filepath.Abs(stowDir)
	targetAbs, _ := filepath.Abs(targetDir)
	expectedOp := Operation{
		Source: filepath.Join(stowDirAbs, "pkg-a", "alpha.txt"),
		Target: filepath.Join(targetAbs, "alpha.txt"),
	}
	if plan.Operations[0] != expectedOp {
		t.Fatalf("operation mismatch: got %+v, want %+v", plan.Operations[0], expectedOp)
	}
	if plan.Conflicts[0].Target != filepath.Join(targetAbs, "alpha.txt") {
		t.Fatalf("unexpected conflict target: %s", plan.Conflicts[0].Target)
	}
	if plan.Conflicts[0].Reason != "duplicate target planned" {
		t.Fatalf("unexpected conflict reason: %s", plan.Conflicts[0].Reason)
	}
}

func TestBuildPlanNoOpForExistingSymlink(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	source := filepath.Join(pkg, "alpha.txt")
	mustWriteFile(t, source)

	if !symlinkSupported(t, stowDir) {
		return
	}

	target := filepath.Join(targetDir, "alpha.txt")
	if err := os.Symlink(source, target); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}
	if len(plan.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %d", len(plan.Conflicts))
	}
	if len(plan.Operations) != 0 {
		t.Fatalf("expected no operations, got %d", len(plan.Operations))
	}
}

func TestBuildPlanMissingPackage(t *testing.T) {
	stowDir := t.TempDir()
	_, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   filepath.Join(t.TempDir(), "target"),
		Packages: []string{"missing"},
	})
	if err == nil {
		t.Fatalf("expected error for missing package")
	}
}

func TestBuildPlanEmptyPackage(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)

	plan, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}
	if len(plan.Operations) != 0 {
		t.Fatalf("expected no operations, got %d", len(plan.Operations))
	}
	if len(plan.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %d", len(plan.Conflicts))
	}
}

func TestDefaultTarget(t *testing.T) {
	stowDir := t.TempDir()
	absDir, _ := filepath.Abs(stowDir)
	defaultTarget, err := DefaultTarget(stowDir)
	if err != nil {
		t.Fatalf("DefaultTarget error: %v", err)
	}
	expected := filepath.Dir(absDir)
	if defaultTarget != expected {
		t.Fatalf("default target mismatch: got %s, want %s", defaultTarget, expected)
	}
}

func TestBuildPlanDoesNotCreateTarget(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := filepath.Join(t.TempDir(), "target")

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	mustWriteFile(t, filepath.Join(pkg, "alpha.txt"))

	_, err := BuildPlan(Options{
		Dir:      stowDir,
		Target:   targetDir,
		Packages: []string{"pkg"},
	})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}

	if _, err := os.Stat(targetDir); err == nil {
		t.Fatalf("expected target directory to not exist after planning")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func symlinkSupported(t *testing.T, dir string) bool {
	t.Helper()
	if runtime.GOOS == "windows" {
		// Allow tests to proceed if symlink creation is permitted.
	}
	target := filepath.Join(dir, "symlink-target")
	link := filepath.Join(dir, "symlink-link")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatalf("write %s: %v", target, err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
		return false
	}
	return true
}
