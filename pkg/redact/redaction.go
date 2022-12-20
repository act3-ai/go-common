// Package redact performs data redaction to prevent credential leakage in logs or the console.
package redact

import "net/url"

// datapolicy could be used to redact sensitive information before logging (not implemented yet).  Something like https://gist.github.com/hvoecking/10772475
// Redacting bools does not make any sense. Redacted pointer, slices, arrays, can be nilled.  Redacted string can be "[REDACTED]".  Redacted values of map[string]string we need to know if the redaction should happen in the key or value or both.  Maybe we can change the type to map[string]Secret wherethe struct Secret has a field that has the datapolicy tag.

// The below redaction approach is OK but not ideal.  I think using the tags on the fields would be a better approach.  Using special types is also possible but often not ideal because it makes the types more complex from a parsing perspective.

// Redacted is a string used to replace redacted data
const Redacted = "[REDACTED]"

// URLString removes the password from the URL if present
func URLString(s string) string {
	if s == "" {
		return s
	}
	u, err := url.Parse(s)
	if err != nil {
		// Should we include err.Error() in the return value?  That might leak credentials.
		return "[Invalid URL]"
	}
	return u.Redacted()
}

// String redacts the string when non-empty
func String(s string) string {
	if s == "" {
		return s
	}
	return Redacted
}
