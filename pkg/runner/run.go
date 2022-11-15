package runner

import (
	"context"
	"os"
	"time"

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

	cmd.PersistentFlags().Int8VarP(&verbosity, "verbosity", "v", int8(config.EnvIntOr(verbosityEnvName, -1)),
		`Verbosity level (setable with env `+verbosityEnvName+")")
	x := cmd.PersistentFlags().Lookup("verbosity")
	x.NoOptDefVal = "0"

	ctx := logr.NewContext(context.Background(), logger)
	return cmd.ExecuteContext(ctx)
}

// switchableSyncWriter adds one level of indirection so we can change the WriteLogger for logtostderr functionality
// There might be a more idomatic way to accomplish this task in GO.
type switchableSyncWriter struct {
	zapcore.WriteSyncer
}

func (s switchableSyncWriter) Sync() error {
	return nil
}

func run(cmd *cobra.Command, verbosityEnvName string) error {
	encConf := zap.NewProductionEncoderConfig()
	encConf.EncodeCaller = zapcore.FullCallerEncoder

	enc := zapcore.NewJSONEncoder(encConf)
	levelEnabler := zap.NewAtomicLevel()
	syncWriter := &switchableSyncWriter{os.Stderr}
	core := zapcore.NewCore(enc, syncWriter, levelEnabler)

	// enabling down sampling
	core = zapcore.NewSamplerWithOptions(core, time.Second, 5, 10)

	// create the zap.Logger
	zapLog := zap.New(core)
	defer zapLog.Sync() //nolint

	// convert to logr.Logger
	logger := zapr.NewLogger(zapLog)

	// flags
	var verbosity int8
	var logtostderr bool

	cobra.OnInitialize(func() {
		// negative sign is necessary since zap has more important levels as higher values
		levelEnabler.SetLevel(zapcore.Level(-verbosity))

		// possibly modify the writer for the zap log
		if !logtostderr {
			// change to stdout for logging
			syncWriter.WriteSyncer = os.Stdout
		}
	})

	cmd.PersistentFlags().Int8VarP(&verbosity, "verbosity", "v", int8(config.EnvIntOr(verbosityEnvName, 0)),
		`Verbosity level (setable with env `+verbosityEnvName+")")
	x := cmd.PersistentFlags().Lookup("verbosity")
	x.NoOptDefVal = "1"

	logstostderrEnvName := "ACE_TELEMETRY_LOGTOSTDERR"
	cmd.PersistentFlags().BoolVar(&logtostderr, "logtostderr", config.EnvBoolOr(logstostderrEnvName, true),
		`Logs to stderr (instead of stdout). (setable with env `+logstostderrEnvName+")")

	ctx := logr.NewContext(context.Background(), logger)
	return cmd.ExecuteContext(ctx)
}
