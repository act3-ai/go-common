package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

	info, err := astutil.LoadPackageInfo(ctx, os.Args[1:]...)
	if err != nil {
		return err
	}

	comments := info.AllComments()

	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	return e.Encode(comments)
}
