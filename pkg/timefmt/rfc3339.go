package timefmt

// RFC3339 formats.
const (
	RFC3339Date          = TimeFormat("2006-01-02")                          // RFC3339 date
	RFC3339DateTime      = TimeFormat("2006-01-02T15:04:05.999999999Z07:00") // RFC3339 date-time with up to nanosecond precision
	RFC3339DateTimeS     = TimeFormat("2006-01-02T15:04:05Z07:00")           // RFC3339 date-time with second precision
	RFC3339DateTimeMilli = TimeFormat("2006-01-02T15:04:05.000Z07:00")       // RFC3339 date-time with millisecond precision
	RFC3339DateTimeMicro = TimeFormat("2006-01-02T15:04:05.000000Z07:00")    // RFC3339 date-time with microsecond precision
	RFC3339DateTimeNano  = TimeFormat("2006-01-02T15:04:05.000000000Z07:00") // RFC3339 date-time with nanosecond precision
)

// RFC3339 UTC formats.
const (
	RFC3339UTCDate          = UTCTimeFormat("2006-01-02")                     // RFC3339 UTC date
	RFC3339UTCDateTime      = UTCTimeFormat("2006-01-02T15:04:05.999999999Z") // RFC3339 UTC date-time with up to nanosecond precision
	RFC3339UTCDateTimeS     = UTCTimeFormat("2006-01-02T15:04:05Z")           // RFC3339 UTC date-time with second precision
	RFC3339UTCDateTimeMilli = UTCTimeFormat("2006-01-02T15:04:05.000Z")       // RFC3339 UTC date-time with millisecond precision
	RFC3339UTCDateTimeMicro = UTCTimeFormat("2006-01-02T15:04:05.000000Z")    // RFC3339 UTC date-time with microsecond precision
	RFC3339UTCDateTimeNano  = UTCTimeFormat("2006-01-02T15:04:05.000000000Z") // RFC3339 UTC date-time with nanosecond precision
)

// RFC3339 UTC regex patterns.
const (
	PatternRFC3339Date             = `[0-9]{4}-[0-9]{2}-[0-9]{2}`
	PatternRFC3339UTCDateTime      = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?Z`
	PatternRFC3339UTCDateTimeS     = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`
	PatternRFC3339UTCDateTimeMilli = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z`
	PatternRFC3339UTCDateTimeMicro = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{6}Z`
	PatternRFC3339UTCDateTimeNano  = `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{9}Z`
)
