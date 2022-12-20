// Package runner provides common bootstrapping functionality for CLI tools
package runner

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"git.act3-ace.com/ace/go-common/pkg/config"
)

// Run will run the root level cobra command but first setup logging with Zap
func Run(cmd *cobra.Command, verbosityEnvName string) error {
	// Create the zap logger configuration
	conf := zap.NewProductionConfig()
	conf.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	// create the concrete zap logger
	zapLog, err := conf.Build()
	if err != nil {
		panic(err)
	}
	defer zapLog.Sync() //nolint

	// convert to logr
	logger := zapr.NewLogger(zapLog)

	// flags
	var verbosity int8

	cobra.OnInitialize(func() {
		// negative sign is necessary since zap has more important levels as higher values
		conf.Level.SetLevel(-zapcore.Level(verbosity))
	})

	cmd.PersistentFlags().Int8VarP(&verbosity, "verbosity", "v", int8(config.EnvIntOr(verbosityEnvName, 0)),
		"Logging verbosity level (also setable with environment variable "+verbosityEnvName+")")
	x := cmd.PersistentFlags().Lookup("verbosity")
	x.NoOptDefVal = "1"

	ctx := logr.NewContext(context.Background(), logger)
	return cmd.ExecuteContext(ctx)
}
