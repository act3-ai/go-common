package version

import (
	"runtime/debug"
)

// Info is the struct to hold the version metadata of this tool
type Info struct {
	// Version is the semantic version
	Version string

	// Commit is the Git commit digest
	Commit string

	// Dirty is true if the build was dirty (not matching the commit)
	Dirty bool

	// Built is the datetime of the last commit
	Built string
}

// GetWithOverride returns the version info
func Get() Info {
	return parse()
}

// parse pulls the version info from the build info
// Some fields will be empty depending on how this was built
func parse() Info {
	v := Info{
		Version: "(unknown)",
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return v
	}

	/*
		The main module's version is only populated when installed with `go install`.
		This is a known issue
		https://github.com/golang/go/issues/50603
		https://github.com/golang/go/issues/29228

		There are some pathological issues with using GIT as the version information for the current module.
		For a given commit with multiple tags, which tag should be used as the version.
	*/
	v.Version = info.Main.Version

	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			v.Commit = kv.Value
		case "vcs.time":
			v.Built = kv.Value
		case "vcs.modified":
			v.Dirty = kv.Value == "true"
		}
	}

	return v
}
