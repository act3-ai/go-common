package timefmt_test

import (
	"testing"
	"time"

	"github.com/act3-ai/go-common/pkg/timefmt"
	"github.com/act3-ai/go-common/pkg/timefmt/timefmttest"
)

func TestUTCTimeFormat(t *testing.T) {
	t.Run("RFC3339UTCDate", func(t *testing.T) {
		timefmttest.Run(t, timefmttest.Tests{
			Layout: timefmt.RFC3339UTCDate,
			Input: timefmttest.InputTestSuite{
				WantDateOnly:    timefmttest.TSDateOnly(),
				WantSecond:      time.Time{},
				WantMillisecond: time.Time{},
				WantMicrosecond: time.Time{},
				WantNanosecond:  time.Time{},
			},
			AdditionalInputTests: []timefmttest.InputTest{
				{
					Name:    "Day without leading zero",
					Value:   "2026-01-1",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Month without leading zero",
					Value:   "2026-1-01",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Time included",
					Value:   "2026-01-01T02:30:04Z",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Trailing T",
					Value:   "2026-01-01T",
					Want:    time.Time{},
					WantErr: true,
				},
			},
			Output: timefmttest.OutputTestSuite{
				WantDateOnly:          "2026-01-01",
				WantSecond:            "2026-01-01",
				WantMillisecond:       "2026-01-01",
				WantMicrosecond:       "2026-01-01",
				WantNanosecond:        "2026-01-01",
				Want999Millisecond:    "2026-01-01",
				Want999Microsecond:    "2026-01-01",
				Want999Nanosecond:     "2026-01-01",
				WantNonUTCNextDay:     "2026-01-02",
				WantNonUTCPreviousDay: "2025-12-31",
				WantZero:              "0001-01-01",
			},
		})
	})
	t.Run("RFC3339UTCDateTime", func(t *testing.T) {
		timefmttest.Run(t, timefmttest.Tests{
			Layout: timefmt.RFC3339UTCDateTime,
			Input: timefmttest.InputTestSuite{
				WantDateOnly:    time.Time{},
				WantSecond:      timefmttest.TSSecond(),
				WantMillisecond: timefmttest.TSMillisecond(),
				WantMicrosecond: timefmttest.TSMicrosecond(),
				WantNanosecond:  timefmttest.TSNanosecond(),
			},
			AdditionalInputTests: []timefmttest.InputTest{
				{
					Name:    "No trailing Z",
					Value:   "2026-01-01T02:30:04",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Unsupported timezone",
					Value:   "2026-01-01T02:30:04Z10:00",
					Want:    time.Time{},
					WantErr: true,
				},
			},
			Output: timefmttest.OutputTestSuite{
				WantDateOnly:          "2026-01-01T00:00:00Z",
				WantSecond:            "2026-01-01T02:30:04Z",
				WantMillisecond:       "2026-01-01T02:30:04.005Z",
				WantMicrosecond:       "2026-01-01T02:30:04.005006Z",
				WantNanosecond:        "2026-01-01T02:30:04.005006007Z",
				Want999Millisecond:    "2026-01-01T02:30:04.999Z",
				Want999Microsecond:    "2026-01-01T02:30:04.005999Z",
				Want999Nanosecond:     "2026-01-01T02:30:04.005006999Z",
				WantNonUTCNextDay:     "2026-01-02T04:30:04.005006007Z",
				WantNonUTCPreviousDay: "2025-12-31T21:30:04.005006007Z",
				WantZero:              "0001-01-01T00:00:00Z",
			},
		})
	})
	t.Run("RFC3339UTCDateTimeS", func(t *testing.T) {
		timefmttest.Run(t, timefmttest.Tests{
			Layout: timefmt.RFC3339UTCDateTimeS,
			Input: timefmttest.InputTestSuite{
				WantDateOnly:    time.Time{},
				WantSecond:      timefmttest.TSSecond(),
				WantMillisecond: timefmttest.TSMillisecond(),
				WantMicrosecond: timefmttest.TSMicrosecond(),
				WantNanosecond:  timefmttest.TSNanosecond(),
			},
			AdditionalInputTests: []timefmttest.InputTest{
				{
					Name:    "No trailing Z",
					Value:   "2026-01-01T02:30:04",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Unsupported timezone",
					Value:   "2026-01-01T02:30:04Z10:00",
					Want:    time.Time{},
					WantErr: true,
				},
			},
			Output: timefmttest.OutputTestSuite{
				WantDateOnly:          "2026-01-01T00:00:00Z",
				WantSecond:            "2026-01-01T02:30:04Z",
				WantMillisecond:       "2026-01-01T02:30:04Z",
				WantMicrosecond:       "2026-01-01T02:30:04Z",
				WantNanosecond:        "2026-01-01T02:30:04Z",
				Want999Millisecond:    "2026-01-01T02:30:04Z",
				Want999Microsecond:    "2026-01-01T02:30:04Z",
				Want999Nanosecond:     "2026-01-01T02:30:04Z",
				WantNonUTCNextDay:     "2026-01-02T04:30:04Z",
				WantNonUTCPreviousDay: "2025-12-31T21:30:04Z",
				WantZero:              "0001-01-01T00:00:00Z",
			},
		})
	})
	t.Run("RFC3339UTCDateTimeMilli", func(t *testing.T) {
		timefmttest.Run(t, timefmttest.Tests{
			Layout: timefmt.RFC3339UTCDateTimeMilli,
			Input: timefmttest.InputTestSuite{
				WantDateOnly:    time.Time{},
				WantSecond:      time.Time{},
				WantMillisecond: timefmttest.TSMillisecond(),
				WantMicrosecond: time.Time{},
				WantNanosecond:  time.Time{},
			},
			AdditionalInputTests: []timefmttest.InputTest{
				{
					Name:    "No trailing Z",
					Value:   "2026-01-01T02:30:04.005",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Unsupported timezone",
					Value:   "2026-01-01T02:30:04.005Z10:00",
					Want:    time.Time{},
					WantErr: true,
				},
			},
			Output: timefmttest.OutputTestSuite{
				WantDateOnly:          "2026-01-01T00:00:00.000Z",
				WantSecond:            "2026-01-01T02:30:04.000Z",
				WantMillisecond:       "2026-01-01T02:30:04.005Z",
				WantMicrosecond:       "2026-01-01T02:30:04.005Z",
				WantNanosecond:        "2026-01-01T02:30:04.005Z",
				Want999Millisecond:    "2026-01-01T02:30:04.999Z",
				Want999Microsecond:    "2026-01-01T02:30:04.005Z",
				Want999Nanosecond:     "2026-01-01T02:30:04.005Z",
				WantNonUTCNextDay:     "2026-01-02T04:30:04.005Z",
				WantNonUTCPreviousDay: "2025-12-31T21:30:04.005Z",
				WantZero:              "0001-01-01T00:00:00.000Z",
			},
		})
	})
	t.Run("RFC3339UTCDateTimeMicro", func(t *testing.T) {
		timefmttest.Run(t, timefmttest.Tests{
			Layout: timefmt.RFC3339UTCDateTimeMicro,
			Input: timefmttest.InputTestSuite{
				WantDateOnly:    time.Time{},
				WantSecond:      time.Time{},
				WantMillisecond: time.Time{},
				WantMicrosecond: timefmttest.TSMicrosecond(),
				WantNanosecond:  time.Time{},
			},
			AdditionalInputTests: []timefmttest.InputTest{
				{
					Name:    "No trailing Z",
					Value:   "2026-01-01T02:30:04.005006",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Unsupported timezone",
					Value:   "2026-01-01T02:30:04.005006Z10:00",
					Want:    time.Time{},
					WantErr: true,
				},
			},
			Output: timefmttest.OutputTestSuite{
				WantDateOnly:          "2026-01-01T00:00:00.000000Z",
				WantSecond:            "2026-01-01T02:30:04.000000Z",
				WantMillisecond:       "2026-01-01T02:30:04.005000Z",
				WantMicrosecond:       "2026-01-01T02:30:04.005006Z",
				WantNanosecond:        "2026-01-01T02:30:04.005006Z",
				Want999Millisecond:    "2026-01-01T02:30:04.999000Z",
				Want999Microsecond:    "2026-01-01T02:30:04.005999Z",
				Want999Nanosecond:     "2026-01-01T02:30:04.005006Z",
				WantNonUTCNextDay:     "2026-01-02T04:30:04.005006Z",
				WantNonUTCPreviousDay: "2025-12-31T21:30:04.005006Z",
				WantZero:              "0001-01-01T00:00:00.000000Z",
			},
		})
	})
	t.Run("RFC3339UTCDateTimeNano", func(t *testing.T) {
		timefmttest.Run(t, timefmttest.Tests{
			Layout: timefmt.RFC3339UTCDateTimeNano,
			Input: timefmttest.InputTestSuite{
				WantDateOnly:    time.Time{},
				WantSecond:      time.Time{},
				WantMillisecond: time.Time{},
				WantMicrosecond: time.Time{},
				WantNanosecond:  timefmttest.TSNanosecond(),
			},
			AdditionalInputTests: []timefmttest.InputTest{
				{
					Name:    "No trailing Z",
					Value:   "2026-01-01T02:30:04.005006007",
					Want:    time.Time{},
					WantErr: true,
				},
				{
					Name:    "Unsupported timezone",
					Value:   "2026-01-01T02:30:04.005006007Z10:00",
					Want:    time.Time{},
					WantErr: true,
				},
			},
			Output: timefmttest.OutputTestSuite{
				WantDateOnly:          "2026-01-01T00:00:00.000000000Z",
				WantSecond:            "2026-01-01T02:30:04.000000000Z",
				WantMillisecond:       "2026-01-01T02:30:04.005000000Z",
				WantMicrosecond:       "2026-01-01T02:30:04.005006000Z",
				WantNanosecond:        "2026-01-01T02:30:04.005006007Z",
				Want999Millisecond:    "2026-01-01T02:30:04.999000000Z",
				Want999Microsecond:    "2026-01-01T02:30:04.005999000Z",
				Want999Nanosecond:     "2026-01-01T02:30:04.005006999Z",
				WantNonUTCNextDay:     "2026-01-02T04:30:04.005006007Z",
				WantNonUTCPreviousDay: "2025-12-31T21:30:04.005006007Z",
				WantZero:              "0001-01-01T00:00:00.000000000Z",
			},
		})
	})
}
