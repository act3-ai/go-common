package testutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertErrorIf(t *testing.T) {
	type args struct {
		wantErr bool
		err     error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"true/pass", args{true, errors.New("test")}, true},
		{"true/fail", args{true, nil}, false},
		{"false/pass", args{false, nil}, true},
		{"false/fail", args{false, errors.New("test")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssertErrorIf(&testing.T{}, tt.args.wantErr, tt.args.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAssertNilIf(t *testing.T) {
	type args struct {
		wantNil bool
		actual  any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"true/nil", args{true, nil}, true},
		{"true/nonNil", args{true, ""}, false},
		{"false/nil", args{false, nil}, false},
		{"false/nonNil", args{false, ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssertNilIf(&testing.T{}, tt.args.wantNil, tt.args.actual)
			assert.Equal(t, tt.want, got)
		})
	}
}
