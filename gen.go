// Package gen is for go:generate directives to generate files.
package gen

// Generate JSON Schema definitions with genschema package
//go:generate go run cmd/sample/gen/main.go cmd/sample/schemas

// Generate CLI documentation with gendocs command
//go:generate go run ./cmd/sample gendocs md cmd/sample/docs/cli --only-commands
