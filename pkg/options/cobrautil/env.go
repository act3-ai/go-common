package cobrautil

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// ParseEnvOverrides receives a flag set after it has been parsed
// and parses environment variables to set the value of any unset
// flags in the FlagSet, if they have an environment variable defined.
// Flag environment variables can be set with [SetEnvName].
// The flag creation functions in pkg/options/flags.go set an
// environment variable for the flag if Option.Env is set.
//
// Parsing errors are handled with cmd.FlagErrorFunc().
// The first non-nil error returned from cmd.FlagErrorFunc()
// is returned.
func ParseEnvOverrides(cmd *cobra.Command) error {
	if !cmd.Flags().Parsed() {
		return errors.New("cannot parse environment variables before command-line arguments")
	}

	// Store first non-empty error.
	var flagErr error

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// skip parsing after first error
		if flagErr != nil {
			return
		}

		err := flagutil.ParseEnvOverrides(f)
		if err != nil {
			// Use command's FlagErrorFunc to handle the env var error the same as flag errs.
			flagErr = cmd.FlagErrorFunc()(cmd, err)
		}
	})

	return flagErr
}
