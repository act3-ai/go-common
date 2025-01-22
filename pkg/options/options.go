// Package options provides a framework for defining all overrides of a configurable option in one location.
package options

import (
	"fmt"
	"strings"
)

// ResolveDescriptions set option descriptions from their target groups, if they specify one.
func ResolveDescriptions(groups ...*Group) error {
	allGroups := map[string]*Group{}
	for _, g := range groups {
		if g.Name != "" {
			allGroups[g.Name] = g
		}
	}
	for _, g := range groups {
		for _, o := range g.Options {
			switch {
			// Already set
			case o.Short != "":
				continue
			// Does not have a reference
			case o.TargetGroupName == "":
				continue
			default:
				target, ok := allGroups[o.TargetGroupName]
				if !ok {
					return fmt.Errorf("Group %q, Option %q: could not resolve TargetGroupName %q", g.Name, o.Header(), o.TargetGroupName)
				}
				o.Short = target.Description
			}
		}
	}
	return nil
}

// Group represents a group of options.
type Group struct {
	Name        string    // Name of the group
	Description string    // Description of the group
	Options     []*Option // Options contained in this group
}

// MarkdownLink produces a markdown link to the group.
func (g *Group) MarkdownLink() string {
	return markdownLink(g.Name)
}

// Type represents the type of an option.
type Type string

// Defined option types.
const (
	String    Type = "string"            // String type.
	Boolean   Type = "boolean"           // Boolean type.
	Integer   Type = "integer"           // Integer type.
	Duration  Type = "duration (string)" // Duration string type.
	Object    Type = "object"            // Object type.
	List      Type = "list"              // List type.
	StringMap Type = "map"               // String map type.
)

// Option represents an option.
type Option struct {
	Type            Type   // Type of the field
	TargetGroupName string // Target group ID
	Default         string // Default value (as a string)
	Path            string // Path to field in JSON config file
	Env             string // Environment variable name
	Flag            string // Flag name
	FlagShorthand   string // Flag shorthand
	FlagUsage       string // Flag usage (if different than the short description)
	Short           string // Short description
	Long            string // Long description
	// Examples    []*Example // Usage examples for this option
}

// formattedFlagUsage produces a flag usage string for the option.
func (o *Option) formattedFlagUsage() string {
	usage := ""
	if o.FlagUsage != "" {
		usage += o.FlagUsage
	} else if o.Short != "" {
		usage += o.Short
	}
	if o.Env != "" {
		usage += " (env: " + o.Env + ")"
	}
	return usage
}

// type ExampleType string

// const (
// 	ExampleJSON ExampleType = "json"
// 	ExampleYAML ExampleType = "yaml"
// 	ExampleFlag ExampleType = "flag"
// 	ExampleEnv  ExampleType = "env"
// )

// type Example struct {
// 	Type        ExampleType
// 	Name        string
// 	Description string
// 	Content     string
// }

// FormattedType formats the type of the option for markdown output.
func (o Option) FormattedType() string {
	switch o.Type {
	case String, Boolean, Integer, Duration:
		return string(o.Type)
	case Object:
		if link := o.TargetLink(); link != "" {
			return link
		}
		return "object"
	case List:
		return "list"
	case StringMap:
		return "object"
		// if link := o.TargetLink(); link != "" {
		// 	return fmt.Sprintf("object(keys: string, values: %s)", link)
		// }
		// return "object(keys: string, values: any)"
	default:
		return o.Default
	}
}

// FormattedDefault formats the default value of the option for markdown output.
func (o Option) FormattedDefault() string {
	if o.Default == "" {
		return ""
	}
	switch o.Type {
	case Boolean, Integer, Object, List, StringMap:
		return o.Default
	case String, Duration:
		return `"` + o.Default + `"`
	default:
		return o.Default
	}
}

// Header formats the name of the option for markdown output.
func (o Option) Header() string {
	switch {
	case o.Path != "":
		return o.Path
	case o.Flag != "":
		return "--" + o.Flag
	case o.Env != "":
		return o.Env
	default:
		return ""
	}
}

// MarkdownLink produces a markdown link to the option.
func (o Option) MarkdownLink() string {
	return markdownLink(o.Header())
}

// TargetLink produces a link to the option's target group.
func (o Option) TargetLink() string {
	if o.TargetGroupName == "" {
		return ""
	}
	// Return empty for unsupported option types.
	if o.Type != Object &&
		o.Type != List &&
		o.Type != StringMap {
		return ""
	}
	return markdownLink(o.TargetGroupName)
}

// ShortDescription produces the short description of the option.
func (o Option) ShortDescription() string {
	switch {
	case o.Short != "":
		return o.Short
	case o.FlagUsage != "":
		return o.FlagUsage
	default:
		return ""
	}
}

// markdownLink produces a markdown link to the given header.
func markdownLink(header string) string {
	return fmt.Sprintf("[%s](#%s)", header, toMarkdownLinkFragment(header))
}

// toMarkdownLinkFragment formats the string as a markdown link fragment.
func toMarkdownLinkFragment(s string) string {
	// Lowercase
	return strings.ToLower(
		// Replace forbidden characters
		mdlinkReplacer.Replace(
			// Trim forbidden leading/trailing characters
			strings.Trim(s, mdlinkCutset)))
}

// mdlinkCutset is used to trim characters from the beginning and end of strings.
var mdlinkCutset = "-"

// mdlinkReplacer replaces characters to produce the equivalent markdown link handle
var mdlinkReplacer = strings.NewReplacer(
	" ", "-",
	".", "",
	"/", "",
	"*", "",
	"`", "",
	"'", "",
	`"`, "",
	"_", "",
)
