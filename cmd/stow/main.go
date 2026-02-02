package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/beppler/gstow/internal/stow"
)

const (
	exitSuccess    = 0
	exitConflicts  = 1
	exitValidation = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("stow", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dryRunShort := fs.Bool("n", false, "dry-run; do not make changes")
	dryRunLong := fs.Bool("no", false, "dry-run; do not make changes")
	dir := fs.String("d", ".", "stow directory")
	dirLong := fs.String("dir", "", "stow directory")
	target := fs.String("t", "", "target directory")
	targetLong := fs.String("target", "", "target directory")

	if err := fs.Parse(args); err != nil {
		stowDir := *dir
		if *dirLong != "" {
			stowDir = *dirLong
		}
		stowTarget := resolveTarget(stowDir, *target, *targetLong)
		if stowTarget == "" {
			stowTarget = stowDir
		}
		writeError(stderr, stowTarget, err)
		return exitValidation
	}

	dryRun := *dryRunShort || *dryRunLong
	stowDir := *dir
	if *dirLong != "" {
		stowDir = *dirLong
	}

	stowTarget := resolveTarget(stowDir, *target, *targetLong)
	if stowTarget == "" {
		defaultTarget, err := stow.DefaultTarget(stowDir)
		if err != nil {
			writeError(stderr, stowDir, err)
			return exitValidation
		}
		stowTarget = defaultTarget
	}

	packages := fs.Args()
	if len(packages) == 0 {
		writeError(stderr, stowTarget, errors.New("at least one package is required"))
		return exitValidation
	}

	plan, err := stow.BuildPlan(stow.Options{
		Dir:      stowDir,
		Target:   stowTarget,
		Packages: packages,
	})
	if err != nil {
		path := ""
		var perr *stow.PathError
		if errors.As(err, &perr) {
			path = perr.Path
		}
		writeError(stderr, path, err)
		return exitValidation
	}

	for _, conflict := range plan.Conflicts {
		writeConflict(stderr, conflict.Target, conflict.Reason)
	}

	for _, op := range plan.Operations {
		fmt.Fprintf(stdout, "LINK %s -> %s\n", op.Target, op.Source)
	}

	if err := stow.Execute(plan, stow.ExecuteOptions{DryRun: dryRun}); err != nil {
		path := ""
		var oerr *stow.OpError
		if errors.As(err, &oerr) {
			path = oerr.Target
		}
		writeError(stderr, path, err)
		return exitValidation
	}

	if len(plan.Conflicts) > 0 {
		return exitConflicts
	}
	return exitSuccess
}

func writeConflict(w io.Writer, target, reason string) {
	fmt.Fprintf(w, "CONFLICT %s: %s\n", target, reason)
}

func writeError(w io.Writer, target string, err error) {
	path := strings.TrimSpace(target)
	if path == "" {
		fmt.Fprintf(w, "ERROR : %v\n", err)
		return
	}
	fmt.Fprintf(w, "ERROR %s: %v\n", path, err)
}

func resolveTarget(stowDir, targetShort, targetLong string) string {
	if strings.TrimSpace(targetLong) != "" {
		return targetLong
	}
	if strings.TrimSpace(targetShort) != "" {
		return targetShort
	}
	if strings.TrimSpace(stowDir) == "" {
		return ""
	}
	defaultTarget, err := stow.DefaultTarget(stowDir)
	if err != nil {
		return ""
	}
	return defaultTarget
}
