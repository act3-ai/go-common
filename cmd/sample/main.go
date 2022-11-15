package main

import (
	"os"

	"github.com/spf13/cobra"

	commands "git.act3-ace.com/ace/go-common/pkg/cmd"
	"git.act3-ace.com/ace/go-common/pkg/runner"
	vv "git.act3-ace.com/ace/go-common/pkg/version"
)

// getVersionInfo retreives the proper version information for this executable
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

func main() {
	info := getVersionInfo()

	// NOTE Often the main command is created elsewhere and imported
	root := &cobra.Command{
		Use: "sample",
	}

	root.AddCommand(
		commands.NewVersionCmd(info),
		commands.NewGendocsCmd(),
	)

	if err := runner.Run(root, "ACE_TELEMETRY_VERBOSITY"); err != nil {
		// fmt.Fprintln(os.Stderr, "Error occurred", err)
		os.Exit(1)
	}
}
