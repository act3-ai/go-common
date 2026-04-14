package timestamp

import (
	"time"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/act3-ai/go-common/pkg/timefmt"
)

// UTCDate is a date-only UTC timestamp in RFC3339 format.
type UTCDate struct {
	Time time.Time
}

// String implements fmt.Stringer.
func (ts UTCDate) String() string {
	return timefmt.RFC3339UTCDate.Format(ts.Time)
}

// IsZero reports whether ts represents the zero time instant,
// January 1, year 1, 00:00:00 UTC.
func (ts UTCDate) IsZero() bool {
	return ts.Time.IsZero()
}

// MarshalJSON implements json.Marshaler.
func (ts UTCDate) MarshalJSON() ([]byte, error) {
	return timefmt.RFC3339UTCDate.TimeMarshalJSON(ts.Time)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ts *UTCDate) UnmarshalJSON(data []byte) error {
	return timefmt.RFC3339UTCDate.TimeUnmarshalJSON(data, &ts.Time)
}

// AppendText implements encoding.TextAppender.
func (ts UTCDate) AppendText(b []byte) ([]byte, error) {
	return timefmt.RFC3339UTCDate.TimeAppendText(b, ts.Time)
}

// MarshalText implements encoding.TextMarshaler.
func (ts UTCDate) MarshalText() ([]byte, error) {
	return timefmt.RFC3339UTCDate.TimeMarshalText(ts.Time)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ts *UTCDate) UnmarshalText(data []byte) error {
	return timefmt.RFC3339UTCDate.TimeUnmarshalText(data, &ts.Time)
}

// JSONSchema produces the JSON Schema representation.
func (UTCDate) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Format:      "date",
		Description: "Date-only UTC timestamp in RFC3339 format.",
		Pattern:     timefmt.PatternRFC3339Date,
		Examples:    []any{timefmt.RFC3339UTCDate},
	}
}

// UTCDateTime is a UTC timestamp in RFC3339 format with up to nanosecond precision.
type UTCDateTime struct {
	Time time.Time
}

// String implements fmt.Stringer.
func (ts UTCDateTime) String() string {
	return timefmt.RFC3339UTCDateTime.Format(ts.Time)
}

// IsZero reports whether ts represents the zero time instant,
// January 1, year 1, 00:00:00 UTC.
func (ts UTCDateTime) IsZero() bool {
	return ts.Time.IsZero()
}

// MarshalJSON implements json.Marshaler.
func (ts UTCDateTime) MarshalJSON() ([]byte, error) {
	return timefmt.RFC3339UTCDateTime.TimeMarshalJSON(ts.Time)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ts *UTCDateTime) UnmarshalJSON(data []byte) error {
	return timefmt.RFC3339UTCDateTime.TimeUnmarshalJSON(data, &ts.Time)
}

// AppendText implements encoding.TextAppender.
func (ts UTCDateTime) AppendText(b []byte) ([]byte, error) {
	return timefmt.RFC3339UTCDateTime.TimeAppendText(b, ts.Time)
}

// MarshalText implements encoding.TextMarshaler.
func (ts UTCDateTime) MarshalText() ([]byte, error) {
	return timefmt.RFC3339UTCDateTime.TimeMarshalText(ts.Time)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ts *UTCDateTime) UnmarshalText(data []byte) error {
	return timefmt.RFC3339UTCDateTime.TimeUnmarshalText(data, &ts.Time)
}

// JSONSchema produces the JSON Schema representation.
func (UTCDateTime) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Format:      "date-time",
		Description: "UTC timestamp in RFC3339 format with up to nanosecond precision.",
		Pattern:     timefmt.PatternRFC3339UTCDateTime,
		Examples:    []any{timefmt.RFC3339UTCDateTime},
	}
}
