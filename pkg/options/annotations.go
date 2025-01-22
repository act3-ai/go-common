package options

import (
	"github.com/spf13/pflag"

	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// ParseEnvOverrides receives a flag set after it has been parsed and
// sets the flag values to environment variables if the flag defines an
// "env" annotation.
//
// Any parsing errors are logged at slog.LevelWarn and are discarded.
func ParseEnvOverrides(flagSet *pflag.FlagSet) {
	flagutil.ParseEnvOverrides(flagSet, envAnno)
}

// GroupFlags marks flags as part of a [Group].
func GroupFlags(g *Group, flags ...*pflag.Flag) {
	groupInfo := []string{g.Name, g.Description}
	for _, f := range flags {
		if f.Annotations == nil {
			f.Annotations = map[string][]string{groupAnno: groupInfo}
		} else {
			f.Annotations[groupAnno] = groupInfo
		}
	}
}

// getGroupFlags returns a list of flags in the flag set that are part of the named group.
func getGroupFlags(flagSet *pflag.FlagSet, group *Group) []*pflag.Flag {
	var flags []*pflag.Flag
	flagSet.VisitAll(func(f *pflag.Flag) {
		flagGroup, ok := flagutil.GetFirstAnnotation(f, groupAnno)
		if ok && flagGroup == group.Name {
			flags = append(flags, f)
		}
	})
	return flags
}

// GetGroupFlagSet returns a filtered FlagSet only containing flags that are part of the named group.
func GetGroupFlagSet(flagSet *pflag.FlagSet, group *Group) *pflag.FlagSet {
	flags := getGroupFlags(flagSet, group)
	if len(flags) == 0 {
		return nil
	}
	out := pflag.NewFlagSet(group.Name, pflag.ContinueOnError)
	for _, f := range flags {
		out.AddFlag(f)
	}
	return out
}

// getNoGroupFlags returns a list of flags in the flag set that are not part of a group.
func getNoGroupFlags(flagSet *pflag.FlagSet) []*pflag.Flag {
	var flags []*pflag.Flag
	flagSet.VisitAll(func(f *pflag.Flag) {
		_, ok := flagutil.GetFirstAnnotation(f, groupAnno)
		if !ok {
			flags = append(flags, f)
		}
	})
	return flags
}

// GetNoGroupFlagSet returns a filtered FlagSet only containing flags that are not part of a group.
func GetNoGroupFlagSet(flagSet *pflag.FlagSet) *pflag.FlagSet {
	flags := getNoGroupFlags(flagSet)
	if len(flags) == 0 {
		return nil
	}
	out := pflag.NewFlagSet("noGroup", pflag.ContinueOnError)
	for _, f := range flags {
		out.AddFlag(f)
	}
	return out
}

const (
	typeAnno        = "type"    // annotation for options.Option.Type
	defaultAnno     = "default" // annotation for options.Option.Default
	jsonAnno        = "json"    // annotation for options.Option.Path
	envAnno         = "env"     // annotation for options.Option.Env
	shortAnno       = "short"   // annotation for options.Option.Short
	longAnno        = "long"    // annotation for options.Option.Long
	targetGroupAnno = "target"  // annotation for options.Option.TargetGroupName
	groupAnno       = "group"   // used to group flags
)

// withOptionConfig adds sets annotations on the flag from the option definition.
func withOptionConfig(f *pflag.Flag, opts *Option) {
	if opts.Type != "" {
		flagutil.SetAnnotation(f, typeAnno, string(opts.Type))
	}
	if opts.TargetGroupName != "" {
		flagutil.SetAnnotation(f, targetGroupAnno, opts.TargetGroupName)
	}
	if opts.Default != "" {
		flagutil.SetAnnotation(f, defaultAnno, opts.Default)
	}
	if opts.Path != "" {
		flagutil.SetAnnotation(f, jsonAnno, opts.Path)
	}
	if opts.Env != "" {
		flagutil.SetAnnotation(f, envAnno, opts.Env)
	}
	if opts.Short != "" {
		flagutil.SetAnnotation(f, shortAnno, opts.Short)
	}
	if opts.Long != "" {
		flagutil.SetAnnotation(f, longAnno, opts.Long)
	}
}

// ToGroups converts all flags in the flag set into grouped options.
func ToGroups(flagSet *pflag.FlagSet) (groups []*Group, ungrouped []*Option) {
	groups = []*Group{}
	groupMap := map[string]*Group{}
	flagSet.VisitAll(func(f *pflag.Flag) {
		opt := toOption(f)
		groupName, ok := flagutil.GetFirstAnnotation(f, groupAnno)
		switch {
		// The flag is not part of a group
		case !ok:
			ungrouped = append(ungrouped, opt)
		// This is not the first option found from this group
		case groupMap[groupName] != nil:
			groupMap[groupName].Options = append(groupMap[groupName].Options, opt)
		// This is the first option found from this group
		default:
			desc := ""
			if len(f.Annotations[groupAnno]) > 1 {
				desc = f.Annotations[groupAnno][1]
			}
			groupMap[groupName] = &Group{
				Name:        groupName,
				Description: desc,
				Options:     []*Option{opt},
			}
			groups = append(groups, groupMap[groupName])
		}
	})
	return groups, ungrouped
}

// Converts all flags in the flag set into a flat list of options.
// func toOptions(flagSet *pflag.FlagSet) []*Option {
// 	opts := make([]*Option, 0, flagSet.NFlag())
// 	flagSet.VisitAll(func(f *pflag.Flag) {
// 		opts = append(opts, toOption(f))
// 	})
// 	return opts
// }

func toOption(f *pflag.Flag) *Option {
	o := &Option{
		Type:            Type(flagutil.GetFirstAnnotationOr(f, typeAnno, "")),
		TargetGroupName: flagutil.GetFirstAnnotationOr(f, targetGroupAnno, ""),
		Default:         flagutil.GetFirstAnnotationOr(f, defaultAnno, ""),
		Path:            flagutil.GetFirstAnnotationOr(f, jsonAnno, ""),
		Flag:            f.Name,
		FlagShorthand:   f.Shorthand,
		Short:           flagutil.GetFirstAnnotationOr(f, shortAnno, ""),
		Long:            flagutil.GetFirstAnnotationOr(f, longAnno, ""),
	}
	// Set short description from annotation if it is different than the flag usage string
	short, ok := flagutil.GetFirstAnnotation(f, shortAnno)
	if ok {
		o.flagUsage = f.Usage
		o.Short = short
	}
	return o
}
