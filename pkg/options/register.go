package options

import (
	"context"

	"github.com/spf13/pflag"
)

// FlagGroups represents groups of options for the configuration type C.
type FlagGroups[C any] []*FlagGroup[C]

// Groups converts each group to an [Group].
func (coll FlagGroups[C]) Groups() []*Group {
	out := make([]*Group, 0, len(coll))
	for _, group := range coll {
		out = append(out, group.Group())
	}
	return out
}

// RegisterFlags registers the all option groups in the given flag set, returning an override function.
func (coll FlagGroups[C]) RegisterFlags(f *pflag.FlagSet) OverrideFunc[C] {
	// Register each group, collecting all override functions into a list
	overrides := make([]OverrideFunc[C], 0, len(coll))
	for _, group := range coll {
		overrides = append(overrides, group.registerFlags(f))
	}
	return JoinOverrides(overrides)
}

// FlagGroup represents a group of options for the configuration type C.
type FlagGroup[C any] struct {
	Key         string
	Title       string
	Description string
	JSON        string
	Flags       []*FlagOption[C] // Options contained in this group
}

// Group converts a group to an [Group].
func (g *FlagGroup[C]) Group() *Group {
	group := &Group{
		Key:         g.Key,
		Title:       g.Title,
		Description: g.Description,
		JSON:        g.JSON,
	}
	for _, opt := range g.Flags {
		group.Options = append(group.Options, opt.Option)
	}
	return group
}

// registerFlags registers the group's options in the given flag set, returning an override function.
func (g *FlagGroup[C]) registerFlags(f *pflag.FlagSet) OverrideFunc[C] {
	// Register each flag, collecting all override functions into a list
	overrides := make([]OverrideFunc[C], 0, len(g.Flags))
	for _, opt := range g.Flags {
		overrides = append(overrides, opt.RegisterFlag(f, opt.Option))
	}
	if g.Key != "" {
		// Group the flags
		GroupOptionFlags(g.Group(), f)
	}
	return JoinOverrides(overrides)
}

// FlagOption is an option for the configuration type C.
type FlagOption[C any] struct {
	// Option documents the Option and its usage
	Option *Option
	// RegisterFlag registers the option in the given flag set, returning an override function
	RegisterFlag func(f *pflag.FlagSet, option *Option) OverrideFunc[C]
}

// MapFlagGroup maps a FlagGroup's override function from a child configuration to a parent configuration.
func MapFlagGroup[Child, Parent any](in *FlagGroup[Child], accessor func(cfg *Parent) *Child) *FlagGroup[Parent] {
	return &FlagGroup[Parent]{
		Key:         in.Key,
		Title:       in.Title,
		Description: in.Description,
		JSON:        in.JSON,
		Flags:       MapFlagOptions(in.Flags, accessor),
	}
}

// MapFlagOptions maps a list of FlagOption's override functions from a child configuration to a parent configuration.
func MapFlagOptions[Child, Parent any](in []*FlagOption[Child], accessor func(cfg *Parent) *Child) []*FlagOption[Parent] {
	out := make([]*FlagOption[Parent], 0, len(in))
	for _, opt := range in {
		out = append(out, MapFlagOption(opt, accessor))
	}
	return out
}

// MapFlagOption maps a FlagOption's override function from a child configuration to a parent configuration.
func MapFlagOption[Child, Parent any](in *FlagOption[Child], accessor func(cfg *Parent) *Child) *FlagOption[Parent] {
	return &FlagOption[Parent]{
		Option: in.Option,
		RegisterFlag: func(f *pflag.FlagSet, option *Option) OverrideFunc[Parent] {
			override := in.RegisterFlag(f, option)
			return func(ctx context.Context, parent *Parent) error {
				child := accessor(parent)
				return override(ctx, child)
			}
		},
	}
}

// Prefix stores the prefixes for flags and environment variables.
type Prefix struct {
	JSON string // JSON field name prefix
	Env  string // Environment variable name prefix
	Flag string // Flag name prefix
}

// OverrideFunc overrides configuration values.
type OverrideFunc[Config any] func(ctx context.Context, c *Config) error

// JoinOverrides joins multiple override functions into a single override function.
func JoinOverrides[C any](overrides []OverrideFunc[C]) OverrideFunc[C] {
	return func(ctx context.Context, c *C) error {
		var err error
		for _, override := range overrides {
			err = override(ctx, c)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
