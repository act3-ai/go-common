package embedutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
)

// adapted from: https://gitlab.com/gitlab-org/cli/-/blob/main/cmd/gen-docs/docs.go

func renderMarkdownTree(cmd *cobra.Command, dir string, opts *Options) error {
	name := commandFilePath(cmd, opts)

	dest := filepath.Join(dir, name)

	// create parent directory
	err := os.MkdirAll(filepath.Dir(dest), 0o775)
	if err != nil {
		return fmt.Errorf("command docs: %w", err)
	}

	// Generate parent command
	out := new(bytes.Buffer)
	err = GenMarkdownCustom(cmd, out)
	if err != nil {
		return err
	}

	err = os.WriteFile(dest, out.Bytes(), 0o644)
	if err != nil {
		return fmt.Errorf("command docs: %w", err)
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
		buf.WriteString("\n## Subcommands\n\n")
		buf.WriteString(subcommands)
	}
}

// GenMarkdownCustom creates custom Markdown output. github.com/spf13/cobra/blob/main/doc/md_docs.go
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	// cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	// Frontmatter (for MkDocs)
	buf.WriteString("---" + "\n")
	buf.WriteString("title: " + cmd.CommandPath() + "\n")
	buf.WriteString("description: " + cmd.Short + "\n")
	buf.WriteString("---" + "\n\n")

	// Generated by a script
	buf.WriteString("<!--" + "\n")
	buf.WriteString("This documentation is auto generated by a script." + "\n")
	buf.WriteString("Please do not edit this file directly." + "\n")
	buf.WriteString("-->" + "\n\n")

	// Disable markdowlint single title rule for the next line
	buf.WriteString("<!-- markdownlint-disable-next-line single-title -->\n")

	buf.WriteString("# " + cmd.CommandPath() + "\n")

	if len(cmd.Short) > 0 {
		buf.WriteString("\n" + cmd.Short + "\n")
	}

	if len(cmd.Long) > 0 {
		buf.WriteString("\n## Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("\n## Usage\n\n```plaintext\n%s\n```\n", cmd.UseLine()))
	}

	if len(cmd.Aliases) > 0 {
		buf.WriteString("\n## Aliases\n\n```plaintext\n")
		for _, a := range cmd.Aliases {
			buf.WriteString(fmt.Sprintf("%s %s\n", cmd.Parent().CommandPath(), a))
		}
		buf.WriteString("```\n")
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("\n## Examples\n\n")
		buf.WriteString(fmt.Sprintf("```sh\n%s\n```\n", cmd.Example))
	}

	printOptions(buf, cmd)

	printSubcommands(cmd, buf)

	_, err := buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("command docs: %w", err)
	}

	return nil
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) {
	flags := cmd.LocalFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("\n## Options\n\n```plaintext\n")
		flags.PrintDefaults()
		buf.WriteString("```\n")
		// TODO: use new flagusages func
		// buf.WriteString("\n## Options\n\n")
		// buf.WriteString("| Flag | Default | Usage |\n")
		// buf.WriteString("| ---- | ------- | ----- |\n")
		// buf.WriteString(flagutil.FlagUsages(flags, flagutil.UsageFormatOptions{LineFunc: flagLineFunc}))
		// buf.WriteString("\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("\n## Options inherited from parent commands\n\n```plaintext\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n")
		// TODO: use new flagusages func
		// buf.WriteString("\n## Options inherited from parent commands\n\n")
		// buf.WriteString("| Flag | Default | Usage |\n")
		// buf.WriteString("| ---- | ------- | ----- |\n")
		// buf.WriteString(flagutil.FlagUsages(parentFlags, flagutil.UsageFormatOptions{LineFunc: flagLineFunc}))
		// buf.WriteString("\n")
	}
}

func flagLineFunc(flag *pflag.Flag) (line string, skip bool) { //nolint:unused
	if flag.Hidden {
		return "", true
	}

	name := "`--" + flag.Name + "`"
	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		name += ", `-" + flag.Shorthand + "`"
	}

	name += " (" + flag.Value.Type() + ")"

	opt := options.FromFlag(flag)

	usage := opt.FlagUsage
	if usage == "" {
		usage = opt.Short
	}
	if opt.Env != "" {
		if usage != "" {
			usage += "<br />"
		}
		usage += "Env: `" + opt.Env + "`"
	}

	return fmt.Sprintf("| %s | %s | %s |",
		name,
		flag.DefValue,
		strings.ReplaceAll(usage, "\n", "<br />"),
	), false
}
