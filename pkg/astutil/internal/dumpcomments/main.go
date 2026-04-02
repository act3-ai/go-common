package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"

	"github.com/act3-ai/go-common/pkg/astutil"
)

func main() {
	if err := mainE(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		os.Exit(1)
	}
}

func mainE(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: dumpcomments PATTERN...")
	}

	info, err := astutil.LoadPackageInfo(ctx, os.Args[1:], func(cfg *packages.Config) {
		cfg.Tests = true
	})
	if err != nil {
		return err
	}

	// comments := info.AllComments()
	comments := astutil.ExtractComments(info.Pkgs)

	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	return e.Encode(comments)
}
