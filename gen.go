// Package gen is for go:generate directives to generate files.
package gen

//go:generate go run cmd/sample/gen/main.go cmd/sample/schemas
//go:generate go run github.com/cpuguy83/go-md2man@latest -in README.md -out cmd/sample/manpages/sample-readme.1
