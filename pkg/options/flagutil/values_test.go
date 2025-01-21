package flagutil

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setMapKeys(t *testing.T) {
	type mapValue struct {
		priority int
	}
	type args struct {
		flagValues []string
		in         map[string]mapValue
	}
	tests := []struct {
		name string
		args args
		want map[string]mapValue
	}{
		{"nilFlagValues",
			args{
				nil,
				map[string]mapValue{
					"a": {priority: 5},
					"b": {priority: 6},
					"c": {priority: 7},
				},
			},
			map[string]mapValue{}},
		{"nilConfig",
			args{
				[]string{"a", "b", "c"},
				nil,
			},
			map[string]mapValue{
				"a": {priority: 0},
				"b": {priority: 0},
				"c": {priority: 0},
			}},
		{"emptyConfig",
			args{
				[]string{"a", "b", "c"},
				map[string]mapValue{},
			},
			map[string]mapValue{
				"a": {priority: 0},
				"b": {priority: 0},
				"c": {priority: 0},
			}},
		{"allExist",
			args{
				[]string{"a", "b", "c"},
				map[string]mapValue{
					"a": {priority: 5},
					"b": {priority: 6},
					"c": {priority: 7},
				},
			},
			map[string]mapValue{
				"a": {priority: 5},
				"b": {priority: 6},
				"c": {priority: 7},
			}},
		{"someExist",
			args{
				[]string{"a", "b", "c"},
				map[string]mapValue{
					"a": {priority: 5},
					"c": {priority: 7},
				},
			},
			map[string]mapValue{
				"a": {priority: 5},
				"b": {priority: 0},
				"c": {priority: 7},
			}},
		{"noMatches",
			args{
				[]string{"a", "b", "c"},
				map[string]mapValue{
					"d": {priority: 5},
					"e": {priority: 7},
				},
			},
			map[string]mapValue{
				"a": {priority: 0},
				"b": {priority: 0},
				"c": {priority: 0},
			}},
		{"filtered",
			args{
				[]string{"a"},
				map[string]mapValue{
					"a": {priority: 5},
					"b": {priority: 6},
					"c": {priority: 7},
				},
			},
			map[string]mapValue{
				"a": {priority: 5},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setMapKeys(tt.args.flagValues, tt.args.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_setMapValues(t *testing.T) {
	type mapValue struct {
		priority int
	}
	type args struct {
		flagValues map[string]string
		in         map[string]mapValue
		setter     SetFunc[string, mapValue]
		// setter     func(key string, flagValue string, entry mapValue, ok bool) (mapValue, bool)
	}
	type testCase struct {
		name string
		args args
		want map[string]mapValue
	}

	// Always returns a mapValue with updated priority field
	alwaysSetInt := SetFunc[string, mapValue](
		func(flagValue string, entry *mapValue, _ bool) bool {
			p, err := strconv.Atoi(flagValue)
			if err != nil {
				return false
			}
			if entry == nil {
				entry = &mapValue{}
			}
			entry.priority = p
			return false
		})

	// Only returns an updated mapValue if the key was found in the existing map
	setIntIfExists := SetFunc[string, mapValue](
		func(flagValue string, entry *mapValue, ok bool) bool {
			if !ok {
				return true
			}
			return alwaysSetInt(flagValue, entry, true)
		})

	tests := []testCase{
		{"nilFlagValues/alwaysSet",
			args{
				nil,
				map[string]mapValue{
					"a": {priority: 5},
					"b": {priority: 6},
					"c": {priority: 7},
				},
				alwaysSetInt,
			},
			map[string]mapValue{
				"a": {priority: 5},
				"b": {priority: 6},
				"c": {priority: 7},
			}},
		{"nilConfig/alwaysSet",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				nil,
				alwaysSetInt,
			},
			map[string]mapValue{
				"a": {priority: 1},
				"b": {priority: 2},
				"c": {priority: 3},
			}},
		{"emptyConfig/alwaysSet",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				map[string]mapValue{},
				alwaysSetInt,
			},
			map[string]mapValue{
				"a": {priority: 1},
				"b": {priority: 2},
				"c": {priority: 3},
			}},
		{"allExist/alwaysSet",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				map[string]mapValue{
					"a": {priority: 5},
					"b": {priority: 6},
					"c": {priority: 7},
				},
				alwaysSetInt,
			},
			map[string]mapValue{
				"a": {priority: 1},
				"b": {priority: 2},
				"c": {priority: 3},
			}},
		{"someExist/alwaysSet",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				map[string]mapValue{
					"a": {priority: 5},
					"c": {priority: 7},
				},
				alwaysSetInt,
			},
			map[string]mapValue{
				"a": {priority: 1},
				"b": {priority: 2},
				"c": {priority: 3},
			}},
		{"nilConfig/setIfExists",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				nil,
				setIntIfExists,
			},
			nil},
		{"emptyConfig/setIfExists",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				map[string]mapValue{},
				setIntIfExists,
			},
			map[string]mapValue{}},
		{"allExist/setIfExists",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				map[string]mapValue{
					"a": {priority: 5},
					"b": {priority: 6},
					"c": {priority: 7},
				},
				setIntIfExists,
			},
			map[string]mapValue{
				"a": {priority: 1},
				"b": {priority: 2},
				"c": {priority: 3},
			}},
		{"someExist/setIfExists",
			args{
				map[string]string{
					"a": "1",
					"b": "2",
					"c": "3",
				},
				map[string]mapValue{
					"a": {priority: 5},
					"c": {priority: 7},
				},
				setIntIfExists,
			},
			map[string]mapValue{
				"a": {priority: 1},
				"c": {priority: 3},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setMapValues(tt.args.flagValues, tt.args.in, tt.args.setter)
			assert.Equal(t, tt.want, got)
		})
	}
}
