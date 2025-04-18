// Package optionshelp produces markdown documentation for options and creates CLI commands utilizing this documentation.
package optionshelp

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	_ "embed"

	"github.com/act3-ai/go-common/pkg/options"
	"github.com/act3-ai/go-common/pkg/termdoc"
	"github.com/act3-ai/go-common/pkg/termdoc/mdfmt"
	"github.com/spf13/cobra"
)

// Command creates a command to display help for the given options.
func Command(name, short string, groups []*options.Group, format *mdfmt.Formatter) *cobra.Command {
	optionsDoc, err := MarkdownDoc(groups)
	if err != nil {
		panic(err)
	}
	return termdoc.AdditionalHelpTopic(name, short, optionsDoc, format)
}

// LazyCommand creates a command to display help for the given options.
func LazyCommand(name, short string, groupFunc func() []*options.Group, format *mdfmt.Formatter) *cobra.Command {
	contentFunc := func(cmd *cobra.Command, args []string) (string, error) {
		return MarkdownDoc(groupFunc())
	}
	return termdoc.LazyAdditionalHelpTopic(name, short, contentFunc, format)
}

// MarkdownDoc produces markdown documentation for the given options.
func MarkdownDoc(groups []*options.Group) (docs string, err error) {
	descErr := options.ResolveDescriptions(groups...)
	defer func() { err = errors.Join(err, descErr) }()

	w := &strings.Builder{}

	err = optionsTemplate.Execute(w, groups)
	if err != nil {
		return w.String(), fmt.Errorf("bug in optionshelp template: %w", err)
	}

	return strings.TrimSpace(w.String()), nil
}

var (
	// Template functions.
	optionsTemplateFunc = template.FuncMap{
		"default":     dfault,
		"groupTable":  groupTable,
		"optionTable": optionTable,
	}

	//go:embed options.md.tmpl
	optionsTemplateStr string

	// Parsed template.
	optionsTemplate = template.Must(
		template.New("").
			Funcs(optionsTemplateFunc).
			Parse(optionsTemplateStr))
)

// SetTemplate overrides the default template.
func SetTemplate(tmpl string) error {
	parsed, err := template.New("").
		Funcs(optionsTemplateFunc).
		Parse(tmpl)
	if err != nil {
		return fmt.Errorf("overriding template: %w", err)
	}
	optionsTemplate = parsed
	return nil
}
