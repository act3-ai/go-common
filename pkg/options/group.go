package options

import (
	"github.com/spf13/pflag"

	"github.com/act3-ai/go-common/pkg/options/flagutil"
)

// groupInfoAnnotation creates the annotation for flags in this group
func groupInfoAnnotation(g *Group) []string {
	return []string{g.Key, g.Title, g.Description, g.JSON}
}

// parseGroupDataFromFlag sets fields from the group annotation on the Group
func parseGroupDataFromFlag(f *pflag.Flag, g *Group) {
	groupValues := f.Annotations[groupAnno]
	if len(groupValues) > 0 {
		g.Key = groupValues[0]
	}
	if len(groupValues) > 1 {
		g.Title = groupValues[1]
	}
	if len(groupValues) > 2 {
		g.Description = groupValues[2]
	}
	if len(groupValues) > 3 {
		g.JSON = groupValues[3]
	}
}

// GroupFlags marks flags as part of a [Group].
func GroupFlags(g *Group, flags ...*pflag.Flag) {
	groupInfo := groupInfoAnnotation(g)
	for _, f := range flags {
		switch {
		// Skip nil flags
		case f == nil:
			continue
		// Initialize annotations and set group annotation values
		case f.Annotations == nil:
			f.Annotations = map[string][]string{groupAnno: groupInfo}
		// Set group annotation values
		default:
			f.Annotations[groupAnno] = groupInfo
		}
	}
}

// GroupOptionFlags marks flags as part of a [Group].
func GroupOptionFlags(g *Group, flagSet *pflag.FlagSet) {
	flagNames := make(map[string]struct{}, len(g.Options))
	for _, opt := range g.Options {
		if opt.Flag != "" {
			flagNames[opt.Flag] = struct{}{}
		}
	}
	// Visit all flags in the FlagSet, collecting a list of matched flags
	var flags []*pflag.Flag
	flagSet.VisitAll(func(f *pflag.Flag) {
		_, ok := flagNames[f.Name]
		if ok {
			// If flag is in the group, add to the collected list
			flags = append(flags, f)
		}
	})
	// Mark the matched flags as part of this group
	GroupFlags(g, flags...)
}

// CombineGroups combines flags in all of the given groups into a single group.
func CombineGroups(combined *Group, flagSet *pflag.FlagSet, groups ...*Group) {
	if len(groups) == 0 {
		return
	}
	// Create map for lookup of old group names
	oldGroup := make(map[string]bool, len(groups))
	for _, g := range groups {
		oldGroup[g.Key] = true
	}
	combinedInfo := groupInfoAnnotation(combined)
	flagSet.VisitAll(func(f *pflag.Flag) {
		groupName, ok := flagutil.GetFirstAnnotation(f, groupAnno)
		if ok && oldGroup[groupName] {
			// update flags that are in one of the targeted groups
			f.Annotations[groupAnno] = combinedInfo
		}
	})
}

// VisitAllGroupFlags visits all flags in the flag set that are part of the group.
//
// Example:
//
//	// Hide all flags in the "boring" group
//	options.VisitAllGroupFlags(flagSet,
//		func(f *pflag.Flag) { f.Hidden = true },
//		&Group{Name:"boring"})
func VisitAllGroupFlags(flagSet *pflag.FlagSet, fn func(*pflag.Flag), groups ...*Group) {
	for _, g := range groups {
		flagSet.VisitAll(func(f *pflag.Flag) {
			groupName, ok := flagutil.GetFirstAnnotation(f, groupAnno)
			if ok && groupName == g.Key {
				fn(f)
			}
		})
	}
}

// ToGroups converts all flags in the flag set into grouped options.
func ToGroups(flagSet *pflag.FlagSet) (groups []*Group, ungrouped []*Option) {
	groups = []*Group{}
	groupMap := map[string]*Group{}
	flagSet.SortFlags = false
	flagSet.VisitAll(func(f *pflag.Flag) {
		opt := FromFlag(f)
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
			// Initialize the group
			g := &Group{
				Key:     groupName,
				Options: []*Option{opt},
			}
			// Set group metadata fields from annotations
			parseGroupDataFromFlag(f, g)
			// Add to map and list
			groupMap[groupName] = g
			groups = append(groups, g)
		}
	})
	return groups, ungrouped
}

// GroupedFlags represents a group of flags.
type GroupedFlags struct {
	*Group
	FlagSet *pflag.FlagSet
}

// ToGroupFlagSets produces a list of groups, corresponding list of flag sets, and a remainder set of ungrouped flags.
func ToGroupFlagSets(flagSet *pflag.FlagSet) (groupList []*GroupedFlags, ungrouped *GroupedFlags) {
	// Group map for lookups
	groupMap := map[string]*GroupedFlags{}

	// Initialize ungrouped var
	ungrouped = &GroupedFlags{
		Group:   &Group{},
		FlagSet: pflag.NewFlagSet("flags", pflag.ContinueOnError),
	}
	ungrouped.FlagSet.SortFlags = flagSet.SortFlags // Preserve parent sort setting

	// flagSet.SortFlags = false
	flagSet.VisitAll(func(f *pflag.Flag) {
		groupName, ok := flagutil.GetFirstAnnotation(f, groupAnno)
		switch {
		// The flag is not part of a group
		case !ok:
			ungrouped.FlagSet.AddFlag(f)
			ungrouped.Options = append(ungrouped.Options, FromFlag(f))
		// This is the first option found from this group
		case groupMap[groupName] == nil:
			g := &GroupedFlags{
				Group:   &Group{Key: groupName},
				FlagSet: pflag.NewFlagSet(groupName, pflag.ContinueOnError),
			}
			g.FlagSet.SortFlags = flagSet.SortFlags // Preserve parent sort setting

			// Parse the group metadata
			parseGroupDataFromFlag(f, g.Group)

			// Add the new group and flag set
			groupMap[groupName] = g
			groupList = append(groupList, g)

			// Continue to next case where flag is added
			fallthrough
		// This is not the first option found from this group
		default:
			g := groupMap[groupName]                   // Get grouped flags
			g.FlagSet.AddFlag(f)                       // Add flag to flag set
			g.Options = append(g.Options, FromFlag(f)) // Add option to options list
		}
	})
	return groupList, ungrouped
}
