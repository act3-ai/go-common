// Package optionshelp produces markdown documentation for options and creates CLI commands utilizing this documentation.
package optionshelp

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/act3-ai/go-common/pkg/options"
	"github.com/act3-ai/go-common/pkg/termdoc"
	"github.com/act3-ai/go-common/pkg/termdoc/mdfmt"
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

	scope := newTemplateScope(groups...)

	err = optionsTemplate.
		Funcs(scope.templateFuncs()).
		Execute(w, groups)
	if err != nil {
		return w.String(), fmt.Errorf("bug in optionshelp template: %w", err)
	}

	return strings.TrimSpace(w.String()), nil
}

var (
	//go:embed options.md.tmpl
	optionsTemplateStr string

	// Parsed template.
	optionsTemplate = template.Must(
		template.New("").
			Funcs(newTemplateScope().templateFuncs()).
			Parse(optionsTemplateStr))
)

// SetTemplate overrides the default template.
func SetTemplate(tmpl string) error {
	parsed, err := template.New("").
		Funcs(newTemplateScope().templateFuncs()).
		Parse(tmpl)
	if err != nil {
		return fmt.Errorf("overriding template: %w", err)
	}
	optionsTemplate = parsed
	return nil
}
