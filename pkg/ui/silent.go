package ui

import (
	"context"

	"github.com/go-logr/logr"
)

// silentUI is a UI that outputs nothing.
type silentUI struct {
	done chan struct{}
}

// NewSilentUI returns a UI that outputs nothing (quiet mode).
func NewSilentUI() UI {
	return &silentUI{
		done: make(chan struct{}),
	}
}

// Run implements UI.
func (u *silentUI) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-u.done:
		return nil
	}
}

// Shutdown implements UI.
func (u *silentUI) Shutdown() {
	close(u.done)
}

// Root implements UI.
func (u *silentUI) Root(ctx context.Context) *Task {
	return newRootTask(logr.FromContextOrDiscard(ctx).WithName("UI").V(1), nil)
}
