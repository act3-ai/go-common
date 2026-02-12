package timefmttest

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/act3-ai/go-common/pkg/timefmt"

	"github.com/act3-ai/go-common/pkg/testutil"
)

type Tests struct {
	Layout                timefmt.Interface
	Input                 InputTestSuite
	AdditionalInputTests  []InputTest
	Output                OutputTestSuite
	AdditionalOutputTests []OutputTest
}

func Run(t *testing.T, suite Tests) {
	t.Helper()
	runInputTestSuite(t, suite.Layout, suite.Input)
	runInputTests(t, suite.Layout, suite.AdditionalInputTests)
	runOutputTestSuite(t, suite.Layout, suite.Output)
	runOutputTests(t, suite.Layout, suite.AdditionalOutputTests)
}

const (
	milliPrecision = 5 * time.Millisecond
	microPrecision = milliPrecision + (6 * time.Microsecond)
	nanoPrecision  = microPrecision + (7 * time.Nanosecond)
)

func TSDateOnly() time.Time    { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
func TSSecond() time.Time      { return time.Date(2026, 1, 1, 2, 30, 4, 0, time.UTC) }
func TSMillisecond() time.Time { return time.Date(2026, 1, 1, 2, 30, 4, int(milliPrecision), time.UTC) }
func TSMicrosecond() time.Time { return time.Date(2026, 1, 1, 2, 30, 4, int(microPrecision), time.UTC) }
func TSNanosecond() time.Time  { return time.Date(2026, 1, 1, 2, 30, 4, int(nanoPrecision), time.UTC) }

type InputTest struct {
	Name    string
	Value   string
	Want    time.Time
	WantErr bool
}

func runInputTests(t *testing.T, layout timefmt.Interface, tests []InputTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Run("Parse", func(t *testing.T) {
				got, err := layout.Parse(tt.Value)
				testutil.AssertErrorIf(t, tt.WantErr, err, "Parse() error")
				assertTimeEquals(t, tt.Want, got, "Parse() output")
			})
			t.Run("ParseInLocation", func(t *testing.T) {
				got, err := layout.ParseInLocation(tt.Value, time.UTC)
				testutil.AssertErrorIf(t, tt.WantErr, err, "ParseInLocation() error")
				assertTimeEquals(t, tt.Want.In(time.UTC), got, "ParseInLocation() output")
			})
			t.Run("TimeUnmarshalJSON", func(t *testing.T) {
				var got time.Time
				err := layout.TimeUnmarshalJSON(fmt.Appendf(nil, "%q", tt.Value), &got)
				testutil.AssertErrorIf(t, tt.WantErr, err, "TimeUnmarshalJSON() error")
				assertTimeEquals(t, tt.Want, got, "TimeUnmarshalJSON() output")
			})
			t.Run("TimeUnmarshalText", func(t *testing.T) {
				var got time.Time
				err := layout.TimeUnmarshalText([]byte(tt.Value), &got)
				testutil.AssertErrorIf(t, tt.WantErr, err, "TimeUnmarshalText() error")
				assertTimeEquals(t, tt.Want, got, "TimeUnmarshalText() output")
			})
		})
	}
}

type InputTestSuite struct {
	WantDateOnly    time.Time
	WantSecond      time.Time
	WantMillisecond time.Time
	WantMicrosecond time.Time
	WantNanosecond  time.Time
}

func runInputTestSuite(t *testing.T, layout timefmt.Interface, suite InputTestSuite) {
	t.Helper()
	runInputTests(t, layout, []InputTest{
		{
			Name:    "Date only",
			Value:   "2026-01-01",
			Want:    suite.WantDateOnly,
			WantErr: suite.WantDateOnly.IsZero(),
		},
		{
			Name:    "Second precision",
			Value:   "2026-01-01T02:30:04Z",
			Want:    suite.WantSecond,
			WantErr: suite.WantSecond.IsZero(),
		},
		{
			Name:    "Millisecond precision",
			Value:   "2026-01-01T02:30:04.005Z",
			Want:    suite.WantMillisecond,
			WantErr: suite.WantMillisecond.IsZero(),
		},
		{
			Name:    "Microsecond precision",
			Value:   "2026-01-01T02:30:04.005006Z",
			Want:    suite.WantMicrosecond,
			WantErr: suite.WantMicrosecond.IsZero(),
		},
		{
			Name:    "Nanosecond precision",
			Value:   "2026-01-01T02:30:04.005006007Z",
			Want:    suite.WantNanosecond,
			WantErr: suite.WantNanosecond.IsZero(),
		},
	})
}

type OutputTest struct {
	Name  string
	Value time.Time
	Want  string
}

func runOutputTests(t *testing.T, layout timefmt.Interface, tests []OutputTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Run("Format", func(t *testing.T) {
				got := layout.Format(tt.Value)
				assert.Equal(t, tt.Want, got, "Format() output")
			})
			t.Run("TimeMarshalJSON", func(t *testing.T) {
				wantJSON := fmt.Sprintf("%q", tt.Want)
				got, err := layout.TimeMarshalJSON(tt.Value)
				require.NoError(t, err, "TimeMarshalJSON() error")
				assert.Equal(t, wantJSON, string(got), "TimeMarshalJSON() output")
			})
			t.Run("TimeMarshalText", func(t *testing.T) {
				got, err := layout.TimeMarshalText(tt.Value)
				require.NoError(t, err, "TimeMarshalText() error")
				assert.Equal(t, tt.Want, string(got), "TimeMarshalText() output")
			})
		})
	}
}

type OutputTestSuite struct {
	WantDateOnly          string
	WantSecond            string
	WantMillisecond       string
	WantMicrosecond       string
	WantNanosecond        string
	Want999Millisecond    string
	Want999Microsecond    string
	Want999Nanosecond     string
	WantNonUTCNextDay     string
	WantNonUTCPreviousDay string
	WantZero              string
}

func runOutputTestSuite(t *testing.T, layout timefmt.Interface, suite OutputTestSuite) {
	t.Helper()
	runOutputTests(t, layout, []OutputTest{
		{
			Name:  "Date only",
			Value: TSDateOnly(),
			Want:  suite.WantDateOnly,
		},
		{
			Name:  "Second precision",
			Value: TSSecond(),
			Want:  suite.WantSecond,
		},
		{
			Name:  "Millisecond precision",
			Value: TSMillisecond(),
			Want:  suite.WantMillisecond,
		},
		{
			Name:  "Microsecond precision",
			Value: TSMicrosecond(),
			Want:  suite.WantMicrosecond,
		},
		{
			Name:  "Nanosecond precision",
			Value: TSNanosecond(),
			Want:  suite.WantNanosecond,
		},
		{
			Name:  "999 milliseconds",
			Value: time.Date(2026, 1, 1, 2, 30, 4, int(999*time.Millisecond), time.UTC),
			Want:  suite.Want999Millisecond,
		},
		{
			Name:  "999 microseconds",
			Value: time.Date(2026, 1, 1, 2, 30, 4, int(milliPrecision+(999*time.Microsecond)), time.UTC),
			Want:  suite.Want999Microsecond,
		},
		{
			Name:  "999 nanoseconds",
			Value: time.Date(2026, 1, 1, 2, 30, 4, int(microPrecision+(999*time.Nanosecond)), time.UTC),
			Want:  suite.Want999Nanosecond,
		},
		{
			Name:  "Non-UTC timezone pushes date to next day",
			Value: time.Date(2026, 1, 1, 23, 30, 4, int(nanoPrecision), time.FixedZone("EST", -5*60*60)),
			Want:  suite.WantNonUTCNextDay,
		},
		{
			Name:  "Non-UTC timezone pushes date to previous day",
			Value: time.Date(2026, 1, 1, 2, 30, 4, int(nanoPrecision), time.FixedZone("PKT", 5*60*60)),
			Want:  suite.WantNonUTCPreviousDay,
		},
		{
			Name:  "Zero value",
			Value: time.Time{},
			Want:  suite.WantZero,
		},
	})
}

func assertTimeEquals(t *testing.T, expected, actual time.Time, msgAndArgs ...any) bool {
	t.Helper()
	if expected.IsZero() {
		assert.Zero(t, actual, msgAndArgs...)
	}
	return assert.Equal(t, expected, actual, msgAndArgs...)
}
