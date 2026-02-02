package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunDryRunOutput(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	source := filepath.Join(pkg, "alpha.txt")
	mustWriteFile(t, source)

	var stdout, stderr bytes.Buffer
	code := run([]string{"-n", "-d", stowDir, "-t", targetDir, "pkg"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	stowDirAbs, _ := filepath.Abs(stowDir)
	targetAbs, _ := filepath.Abs(targetDir)
	expected := "LINK " + filepath.Join(targetAbs, "alpha.txt") + " -> " + filepath.Join(stowDirAbs, "pkg", "alpha.txt") + "\n"
	if stdout.String() != expected {
		t.Fatalf("stdout mismatch:\n got: %q\nwant: %q", stdout.String(), expected)
	}
}

func TestRunConflictExitCode(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	pkg := filepath.Join(stowDir, "pkg")
	mustMkdir(t, pkg)
	source := filepath.Join(pkg, "alpha.txt")
	mustWriteFile(t, source)

	conflict := filepath.Join(targetDir, "alpha.txt")
	mustWriteFile(t, conflict)

	var stdout, stderr bytes.Buffer
	code := run([]string{"-n", "-d", stowDir, "-t", targetDir, "pkg"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}

	expected := "CONFLICT " + conflict + ": target already exists\n"
	if stderr.String() != expected {
		t.Fatalf("stderr mismatch:\n got: %q\nwant: %q", stderr.String(), expected)
	}
}

func TestRunValidationExitCode(t *testing.T) {
	stowDir := t.TempDir()
	targetDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	code := run([]string{"-d", stowDir, "-t", targetDir}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}

	expected := "ERROR " + targetDir + ": at least one package is required\n"
	if stderr.String() != expected {
		t.Fatalf("stderr mismatch:\n got: %q\nwant: %q", stderr.String(), expected)
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
