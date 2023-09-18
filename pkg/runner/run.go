// Package runner provides common bootstrapping functionality for CLI tools
package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// RunWithContext will run the root level cobra command but first setup logging with Zap
func RunWithContext(ctx context.Context, cmd *cobra.Command, verbosityEnvName string) error {

	level := new(slog.LevelVar)
	options := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, options))

	// Flags
	var verbosityFlag []string

	// Set verbosity in the "OnInitialize" function,
	//  verbosity flag must be parsed before it can be used
	cobra.OnInitialize(func() {
		// Convert verbosity flag input to a slog.Level
		level.Set(getLogLevel(verbosityFlag))
	})

	cmd.PersistentFlags().StringSliceVarP(&verbosityFlag, "verbosity", "v",
		[]string{config.EnvOr(verbosityEnvName, "error")},
		`Logging verbosity level (also setable with environment variable `+verbosityEnvName+`)
Aliases: error=0, warn=4, info=8, debug=12`)
	x := cmd.PersistentFlags().Lookup("verbosity")
	x.NoOptDefVal = "warn"

	ctx = logger.NewContext(ctx, log)
	return cmd.ExecuteContext(ctx)
}

// Run will run the root level cobra command but first setup logging with Zap
func Run(cmd *cobra.Command, verbosityEnvName string) error {
	return RunWithContext(context.Background(), cmd, verbosityEnvName)
}

var (
	verbosityAliases = map[string]int{
		"error": 0,
		"warn":  4,
		"info":  8,
		"debug": 12,
	}
)

/*
slog.Level values from https://pkg.go.dev/log/slog#Level

	const (
		LevelDebug Level = -4
		LevelInfo  Level = 0
		LevelWarn  Level = 4
		LevelError Level = 8
	)

LevelError is set as the default logging level.

For slog, "lower" levels mean a chattier logger, so a user is intending to decrease the value of the slog logger's Level when they increase the value of the verbosity flag. Since slog's levels are on multiples of 4, the value of the verbosity flag is multiplied by 4 to easily increase the verbosity to the next level defined. Without the multiplication, a user rerunning a command with a verbosity of 1 to see more logs would see no difference in output, and there is no reason for them to learn the conventions of the Go log/slog package to confidently use our tools.
*/
func getLogLevel(verbosityFlag []string) slog.Level {
	// Start with the default level
	level := slog.LevelError

	// Iterate over flag values, subtracting from the slog level to increase verbosity
	for _, val := range verbosityFlag {
		if l, ok := verbosityAliases[val]; ok {
			// Set level to verbosity alias level
			level -= slog.Level(l)
		} else if l, err := strconv.Atoi(val); err == nil {
			// Set integer verbosity
			level -= slog.Level(l)
		} else {
			fmt.Printf("Error: invalid argument %q for \"-v, --verbosity\" flag\n", val) //nolint:revive
			os.Exit(1)
		}
	}

	return level
}
