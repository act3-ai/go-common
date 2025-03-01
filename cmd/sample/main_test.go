package main

import (
	"testing"
)

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
