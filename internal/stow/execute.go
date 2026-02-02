package stow

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExecuteOptions controls execution behavior.
type ExecuteOptions struct {
	DryRun bool
}

// OpError provides context for execution failures.
type OpError struct {
	Target string
	Err    error
}

func (e *OpError) Error() string {
	return fmt.Sprintf("%s: %v", e.Target, e.Err)
}

func (e *OpError) Unwrap() error {
	return e.Err
}

// Execute applies planned operations. When DryRun is true, it makes no filesystem changes.
func Execute(plan PlanResult, opts ExecuteOptions) error {
	if opts.DryRun {
		return nil
	}
	for _, op := range plan.Operations {
		parent := filepath.Dir(op.Target)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return &OpError{Target: op.Target, Err: err}
		}
		if err := os.Symlink(op.Source, op.Target); err != nil {
			return &OpError{Target: op.Target, Err: err}
		}
	}
	return nil
}
