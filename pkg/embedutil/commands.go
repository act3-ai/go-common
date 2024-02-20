// Package embedutil defines utilities for embedded files
//
//nolint:unhandled-error
package embedutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// adapted from: https://gitlab.com/gitlab-org/cli/-/blob/main/cmd/gen-docs/docs.go

func renderMarkdownTree(cmd *cobra.Command, dir string, opts *Options) error {
	name := commandFilePath(cmd, opts)

	dest := filepath.Join(dir, name)

	// create parent directory
	err := os.MkdirAll(filepath.Dir(dest), 0o775)
	if err != nil {
		return err
	}

	// Generate parent command
	out := new(bytes.Buffer)
	err = GenMarkdownCustom(cmd, out)
	if err != nil {
		return err
	}

	err = os.WriteFile(dest, out.Bytes(), 0o644)
	if err != nil {
		return err
	}

	for _, cmdC := range cmd.Commands() {
		if cmdC.Name() == "help" {
			continue // skip help commands
		}

		err = renderMarkdownTree(cmdC, dir, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func commandFilePath(cmd *cobra.Command, opts *Options) string {
	switch {
	case opts.Flat:
		// Flat output writes all files using the full command path
		return strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
	case cmd.HasAvailableSubCommands():
		// Parent of a command group is written to <name>/index.md
		name := filepath.Join(strings.Split(cmd.CommandPath(), " ")[1:]...)
		name = filepath.Join(name, "index.md")
		return name
	default:
		// Member of a command group is written to <parent>/<name>.md
		return filepath.Join(strings.Split(cmd.CommandPath(), " ")[1:]...) + ".md"
	}
}

func printSubcommands(cmd *cobra.Command, buf *bytes.Buffer) {
	if !cmd.HasAvailableSubCommands() {
		return
	}

	var subcommands string

	// Generate children commands
	for _, cmdC := range cmd.Commands() {
		if cmdC.Name() == "help" {
			continue // skip help commands
		}

		if cmdC.HasAvailableSubCommands() {
			subcommands += fmt.Sprintf("- [`%s`](%s/index.md)", cmdC.CommandPath(), cmdC.Name())
		} else {
			subcommands += fmt.Sprintf("- [`%s`](%s.md)", cmdC.CommandPath(), cmdC.Name())
		}
		if cmdC.Short != "" {
			subcommands += " - " + cmdC.Short
		}
		subcommands += "\n"
	}

	if subcommands != "" {
		_, _ = buf.WriteString("\n## Subcommands\n\n")
		_, _ = buf.WriteString(subcommands)
	}
}

// GenMarkdownCustom creates custom Markdown output. github.com/spf13/cobra/blob/main/doc/md_docs.go
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	// cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	// Frontmatter (for MkDocs)
	_, _ = buf.WriteString("---" + "\n")
	_, _ = buf.WriteString("title: " + cmd.CommandPath() + "\n")
	_, _ = buf.WriteString("description: " + cmd.Short + "\n")
	_, _ = buf.WriteString("---" + "\n\n")

	// Generated by a script
	_, _ = buf.WriteString("<!--" + "\n")
	_, _ = buf.WriteString("This documentation is auto generated by a script." + "\n")
	_, _ = buf.WriteString("Please do not edit this file directly." + "\n")
	_, _ = buf.WriteString("-->" + "\n\n")

	// Disable markdowlint single title rule for the next line
	_, _ = buf.WriteString("<!-- markdownlint-disable-next-line single-title -->\n")

	_, _ = buf.WriteString("# " + cmd.CommandPath() + "\n")

	if len(cmd.Short) > 0 {
		_, _ = buf.WriteString("\n" + cmd.Short + "\n")
	}

	if len(cmd.Long) > 0 {
		_, _ = buf.WriteString("\n## Synopsis\n\n")
		_, _ = buf.WriteString(cmd.Long + "\n")
	}

	if cmd.Runnable() {
		_, _ = buf.WriteString(fmt.Sprintf("\n## Usage\n\n```plaintext\n%s\n```\n", cmd.UseLine()))
	}

	if len(cmd.Aliases) > 0 {
		_, _ = buf.WriteString("\n## Aliases\n\n```plaintext\n")
		for _, a := range cmd.Aliases {
			_, _ = buf.WriteString(fmt.Sprintf("%s %s\n", cmd.Parent().CommandPath(), a))
		}
		_, _ = buf.WriteString("```\n")
	}

	if len(cmd.Example) > 0 {
		_, _ = buf.WriteString("\n## Examples\n\n")
		_, _ = buf.WriteString(fmt.Sprintf("```plaintext\n%s\n```\n", cmd.Example))
	}

	printOptions(buf, cmd)

	printSubcommands(cmd, buf)

	_, err := buf.WriteTo(w)
	return err
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		_, _ = buf.WriteString("\n## Options\n\n```plaintext\n")
		flags.PrintDefaults()
		_, _ = buf.WriteString("```\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		_, _ = buf.WriteString("\n## Options inherited from parent commands\n\n```plaintext\n")
		parentFlags.PrintDefaults()
		_, _ = buf.WriteString("```\n")
	}
}
