// Package optionshelp produces markdown documentation for options and creates CLI commands utilizing this documentation.
package optionshelp

import (
	"fmt"
	"strings"
	"text/template"

	_ "embed"

	"github.com/spf13/cobra"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/mdfmt"
)

// Command creates a command to display help for the given options.
func Command(name, short string, groups []*options.Group, format *mdfmt.Formatter) *cobra.Command {
	optionsDoc, err := MarkdownDoc(groups)
	if err != nil {
		panic(err)
	}
	return termdoc.AdditionalHelpTopic(name, short, optionsDoc, format)
}

// MarkdownDoc produces markdown documentation for the given options.
func MarkdownDoc(groups []*options.Group) (string, error) {
	err := options.ResolveDescriptions(groups...)
	if err != nil {
		return "", err
	}

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
