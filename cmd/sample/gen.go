// Package gen is for go:generate directives to generate files.
package main

// Generate JSON Schema definitions with genschema package
//go:generate go run gen/main.go schemas

// Generate CLI documentation with gendocs command
//go:generate go run . gendocs md docs/cli --only-commands
