package formats

import (
	"testing"

	"github.com/act3-ai/go-common/pkg/options"
	"github.com/act3-ai/go-common/pkg/options/flagutil"
	"github.com/act3-ai/go-common/pkg/termdoc/codefmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func Test_faintCommentsCodeFormatter(t *testing.T) {
	codeFormatter := faintCommentsCodeFormatter()
	got := codeFormatter.Comment("# test comment", codefmt.Location{})
	assert.Equal(t, "# test comment", ansi.Strip(got))
}

func Test_Colorful_Usage(t *testing.T) {
	formatOptions := Colorful()

	t.Run("Header", func(t *testing.T) {
		formatOptions.Format.Header("Text")
	})
	t.Run("Command", func(t *testing.T) {
		formatOptions.Format.Command("Text")
	})
	t.Run("Args", func(t *testing.T) {
		formatOptions.Format.Args("Text")
	})
	t.Run("Example", func(t *testing.T) {
		formatOptions.Format.Example("Text")
	})
}

func Test_Colorful_Flags(t *testing.T) {
	formatOptions := Colorful()

	var plainFlagValue bool
	var optionsFlagValue bool
	flagSet := pflag.NewFlagSet("test flag set", pflag.ContinueOnError)
	plainFlag := flagutil.BoolVar(flagSet, &plainFlagValue, "test", false, "Test plain flag")
	optionFlag := options.BoolVar(flagSet, &optionsFlagValue, false, &options.Option{
		Type:     options.Boolean,
		Default:  "Default",
		Name:     "optionsFlag",
		JSON:     "optionsFlag",
		Env:      "OPTIONS_FLAG",
		Flag:     "options-flag",
		FlagType: "FlagType",
		Short:    "Test options flag",
	})
	t.Run("Columns", func(t *testing.T) {
		formatOptions.FlagOptions.Columns.Value()
	})
	t.Run("FormatFlagName", func(t *testing.T) {
		formatOptions.FlagOptions.FormatFlagName(plainFlag, "test")
	})
	t.Run("FormatType", func(t *testing.T) {
		_ = flagSet.BoolSlice("test-bools", nil, "Test bools flag")
		_ = flagSet.StringSlice("test-strings", nil, "Test strings flag")
		_ = flagSet.IntSlice("test-ints", nil, "Test ints flag")
		_ = flagSet.UintSlice("test-uints", nil, "Test uints flag")
		tests := []struct {
			name     string
			flag     *pflag.Flag
			typeName string
			want     string
		}{
			{"bool", plainFlag, "bool", "bool"},
			{"strings", flagSet.Lookup("test-strings"), "strings", "string..."},
			{"ints", flagSet.Lookup("test-ints"), "ints", "int..."},
			{"uints", flagSet.Lookup("test-uints"), "uints", "uint..."},
			{"bools", flagSet.Lookup("test-bools"), "bools", "bool..."},
			{"override from options", optionFlag, "bool", "FlagType"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := formatOptions.FlagOptions.FormatType(tt.flag, tt.typeName)
				assert.Equal(t, tt.want, ansi.Strip(got))
			})
		}
	})
	t.Run("FormatValue", func(t *testing.T) {
		t.Run("plain", func(t *testing.T) {
			got := formatOptions.FlagOptions.FormatValue(plainFlag, "false")
			assert.Equal(t, "false", ansi.Strip(got))
		})
		t.Run("options", func(t *testing.T) {
			got := formatOptions.FlagOptions.FormatValue(optionFlag, "false")
			assert.Equal(t, "false", ansi.Strip(got))
		})
	})
	t.Run("FormatUsage", func(t *testing.T) {
		t.Run("plain", func(t *testing.T) {
			got := formatOptions.FlagOptions.FormatUsage(plainFlag, "Test plain flag")
			assert.Equal(t, "Test plain flag", ansi.Strip(got))
		})
		t.Run("options", func(t *testing.T) {
			got := formatOptions.FlagOptions.FormatUsage(optionFlag, "Test options flag")
			assert.Equal(t, "Test options flag (env: OPTIONS_FLAG)", ansi.Strip(got))
		})
	})
}
