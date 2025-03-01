package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainSetup(t *testing.T) {
	ctx, rootCmd, otelCfg, err := mainSetup()
	assert.NoError(t, err)

	// Check if context is not nil
	assert.NotNil(t, ctx, "Context should not be nil")

	// Check if root command is not nil
	assert.NotNil(t, rootCmd, "Root command should not be nil")

	// Check if otelCfg is not nil
	assert.NotNil(t, otelCfg, "Otel config should not be nil")

	// Check if the root command has the expected subcommands
	commandNames := map[string]struct{}{}
	for _, cmd := range rootCmd.Commands() {
		commandNames[cmd.Name()] = struct{}{}
	}
	assert.Contains(t, commandNames, "version")

	rootCmd.SetContext(ctx)
	// test root command PersistentPreRun by running the main func
	// rootCmd.PersistentPreRun(rootCmd, []string{})
}

func Test_mainE(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"help command", []string{"--help"}, false},
		{"version command", []string{"version"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mainE(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("mainE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
