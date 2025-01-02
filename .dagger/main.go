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
	"strings"

	"github.com/sourcegraph/conc/pool"
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

// Lint yaml files
func (m *GoCommon) Yamllint(ctx context.Context,
	// Source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("docker.io/cytopia/yamllint:1").
		WithWorkdir("/src").
		WithDirectory("/src", src).
		WithExec([]string{"yamllint", "."}).
		Stdout(ctx)
}

// Lint markdown files
func (m *GoCommon) Markdownlint(ctx context.Context,
	// source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("docker.io/davidanson/markdownlint-cli2:v0.14.0").
		WithWorkdir("/src").
		WithDirectory("/src", src).
		WithExec([]string{"markdownlint-cli2", "."}).
		Stdout(ctx)
}

// Lint all files
func (m *GoCommon) Lint(ctx context.Context) (string, error) {
	p := pool.NewWithResults[string]().WithContext(ctx)

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "yamllint")
		defer span.End()
		return m.Yamllint(ctx, m.Source)
	})

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "markdownlint")
		defer span.End()
		return m.Markdownlint(ctx, m.Source)
	})

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "golangci-lint")
		defer span.End()
		return dag.GolangciLint().
			Run(m.Source, dagger.GolangciLintRunOpts{
				Timeout: "5m",
			}).
			Stdout(ctx)
	})

	s, err := p.Wait()
	// TODO maybe we should order the lint result strings
	return strings.Join(s, "\n=====\n"), err
}

// Run all tests
func (m *GoCommon) Test(
	ctx context.Context,
) (string, error) {
	return dag.Go().
		WithSource(m.Source).
		Exec([]string{"test", "./..."}).
		Stdout(ctx)
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
