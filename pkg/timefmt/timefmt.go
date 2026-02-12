package timefmt

import (
	"bytes"
	"encoding/json"
	"time"
)

// Interface represents a timestamp format.
type Interface interface {
	Parse(value string) (time.Time, error)
	ParseInLocation(value string, loc *time.Location) (time.Time, error)
	Format(ts time.Time) string

	TimeMarshalJSON(ts time.Time) ([]byte, error)
	TimeUnmarshalJSON(data []byte, ts *time.Time) error
	TimeAppendText(b []byte, ts time.Time) ([]byte, error)
	TimeMarshalText(ts time.Time) ([]byte, error)
	TimeUnmarshalText(data []byte, ts *time.Time) error
}

// TimeFormat defines a timestamp format.
type TimeFormat string

// Parse parses a timestamp.
func (layout TimeFormat) Parse(value string) (time.Time, error) {
	return time.Parse(string(layout), value)
}

// ParseInLocation parses a timestamp using the given location.
func (layout TimeFormat) ParseInLocation(value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(string(layout), value, loc)
}

// Format is used for implementing fmt.Stringer.
func (layout TimeFormat) Format(ts time.Time) string {
	return ts.Format(string(layout))
}

// TimeMarshalJSON is used for implementing json.Marshaler.
func (layout TimeFormat) TimeMarshalJSON(ts time.Time) ([]byte, error) {
	return json.Marshal(layout.Format(ts))
}

// TimeUnmarshalJSON is used for implementing json.Unmarshaler.
func (layout TimeFormat) TimeUnmarshalJSON(data []byte, ts *time.Time) error {
	return timeUnmarshalJSON(string(layout), data, ts)
}

// TimeAppendText is used for implementing encoding.TextAppender.
func (layout TimeFormat) TimeAppendText(b []byte, ts time.Time) ([]byte, error) {
	return ts.AppendFormat(b, string(layout)), nil
}

// TimeMarshalText is used for implementing encoding.TextMarshaler.
func (layout TimeFormat) TimeMarshalText(ts time.Time) ([]byte, error) {
	return layout.TimeAppendText(make([]byte, 0, len(layout)), ts)
}

// TimeUnmarshalText is used for implementing encoding.TextUnmarshaler.
func (layout TimeFormat) TimeUnmarshalText(data []byte, ts *time.Time) error {
	return timeUnmarshalText(string(layout), string(data), ts)
}

// UTCTimeFormat defines a timestamp format that will always be serialized as UTC.
type UTCTimeFormat string

// Parse parses a timestamp.
func (layout UTCTimeFormat) Parse(value string) (time.Time, error) {
	return time.Parse(string(layout), value)
}

// ParseInLocation parses a timestamp using the given location.
func (layout UTCTimeFormat) ParseInLocation(value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(string(layout), value, loc)
}

// Format is used for implementing fmt.Stringer.
func (layout UTCTimeFormat) Format(ts time.Time) string {
	return ts.UTC().Format(string(layout))
}

// TimeMarshalJSON is used for implementing json.Marshaler.
func (layout UTCTimeFormat) TimeMarshalJSON(ts time.Time) ([]byte, error) {
	return json.Marshal(layout.Format(ts))
}

// TimeUnmarshalJSON is used for implementing json.Unmarshaler.
func (layout UTCTimeFormat) TimeUnmarshalJSON(data []byte, ts *time.Time) error {
	return timeUnmarshalJSON(string(layout), data, ts)
}

// TimeAppendText is used for implementing encoding.TextAppender.
func (layout UTCTimeFormat) TimeAppendText(b []byte, ts time.Time) ([]byte, error) {
	return ts.UTC().AppendFormat(b, string(layout)), nil
}

// TimeMarshalText is used for implementing encoding.TextMarshaler.
func (layout UTCTimeFormat) TimeMarshalText(ts time.Time) ([]byte, error) {
	return layout.TimeAppendText(make([]byte, 0, len(layout)), ts)
}

// TimeUnmarshalText is used for implementing encoding.TextUnmarshaler.
func (layout UTCTimeFormat) TimeUnmarshalText(data []byte, ts *time.Time) error {
	return timeUnmarshalText(string(layout), string(data), ts)
}

var nullBytes = []byte(`null`)

// timeUnmarshalJSON is a standardized parser for JSON timestamps.
func timeUnmarshalJSON(layout string, data []byte, ts *time.Time) error {
	if bytes.Equal(data, nullBytes) {
		return nil
	}
	var strValue string
	err := json.Unmarshal(data, &strValue)
	if err != nil {
		return err
	}
	return timeUnmarshalText(layout, strValue, ts)
}

// timeUnmarshalText is a standardized parser for text timestamps.
func timeUnmarshalText(layout string, text string, ts *time.Time) error {
	parsed, err := time.Parse(layout, text)
	if err != nil {
		return err
	}
	*ts = parsed
	return nil
}
