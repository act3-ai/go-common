// A generated module for GoCommon functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/go-common/internal/dagger"
)

type GoCommon struct {
	// source code directory
	Source *dagger.Directory
}

func New(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) *GoCommon {
	return &GoCommon{
		Source: src,
	}
}

// Run all tests
func (m *GoCommon) Test(
	ctx context.Context,
) (string, error) {
	return dag.Go().
		WithSource(m.Source).
		Exec([]string{"test", "./..."}).Stdout(ctx)
}

// Build the sample executable
func (m *GoCommon) Build(
	ctx context.Context,
) *dagger.File {
	return dag.Go().
		WithSource(m.Source).
		WithCgoDisabled().
		Build()
}
