package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// instead of using reflection, we can create our own enum of types
const (
	stringType = iota
	intType
	boolType
	quantityType
	durationType
	stringArrayType
	pathType
)

// helper is to store the needed information about each variable added to the envstruct
type helper struct {
	varType int
	name    string
	pntr    any
	sep     string

	// function for handling successful lookups and parses
	handleSuccess func()

	// functions for handling errors
	handleLookupErr func() error
	handleParseErr  func(failedStr string) error
}

// docString returns the documentation for the variable
func (h *helper) docString() string {
	switch h.varType {
	case stringType:
		return fmt.Sprintf("string var: %s, allows any valid string", h.name)
	case intType:
		return fmt.Sprintf("int var: %s, allows any valid integer", h.name)
	case boolType:
		return fmt.Sprintf("bool var: %s, allows true or false", h.name)
	case quantityType:
		return fmt.Sprintf("quantity var: %s, allows any valid resource.Quantity", h.name)
	case durationType:
		return fmt.Sprintf("duration var: %s, allows any valid time.Duration", h.name)
	case stringArrayType:
		return fmt.Sprintf("string array var: %s, allows any valid string array with seperator: %s", h.name, h.sep)
	case pathType:
		return fmt.Sprintf("path var: %s, allows any valid path with seperator: %s", h.name, h.sep)
	default:
		return ""
	}
}

// lookup funcs for each type
func (h *helper) lookupString() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	constType := h.pntr.(*string)
	*constType = envVal
	h.handleSuccess()
	return nil
}

func (h *helper) lookupInt() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	parsedVal, err := strconv.Atoi(envVal)
	if err != nil {
		return h.handleParseErr(envVal)
	}
	constType := h.pntr.(*int)
	*constType = parsedVal
	h.handleSuccess()
	return nil
}

func (h *helper) lookupBool() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	parsedVal, err := strconv.ParseBool(envVal)
	if err != nil {
		return h.handleParseErr(envVal)
	}
	constType := h.pntr.(*bool)
	*constType = parsedVal
	h.handleSuccess()
	return nil
}

func (h *helper) lookupQuantity() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	parsedVal, err := resource.ParseQuantity(envVal)
	if err != nil {
		return h.handleParseErr(envVal)
	}
	constType := h.pntr.(*resource.Quantity)
	*constType = parsedVal // We don't allow nil pointers so this is safe
	h.handleSuccess()
	return nil
}

func (h *helper) lookupDuration() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	parsedVal, err := time.ParseDuration(envVal)
	if err != nil {
		return h.handleParseErr(envVal)
	}
	constType := h.pntr.(*time.Duration)
	*constType = parsedVal
	h.handleSuccess()
	return nil
}

func (h *helper) lookupStringArray() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	parsedVal := strings.Split(envVal, h.sep)
	constType := h.pntr.(*[]string)
	*constType = parsedVal
	h.handleSuccess()
	return nil
}

func (h *helper) lookupPath() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return h.handleLookupErr()
	}
	parsedVal := strings.Split(envVal, h.sep)
	constType := h.pntr.(*[]string)
	*constType = parsedVal
	h.handleSuccess()
	return nil
}

// EnvStruct is an environment variable override helper.
// Each variable added to the EnvStruct is looked up and parsed from the environment during runtime.
// Each variable also gets documentation generated based on type and name.
// Variables added to the EnvStruct are non-nil pointers to the type of variable.
// Set the handle functions to customize the 3 different end states of each variable.
type EnvStruct struct {
	// the variables that are added to the struct
	variables []helper

	// function for handling successful lookups and parses
	handleSuccess func(name string, value reflect.Value)

	// functions for handling errors
	handleLookupErr func(name string, err error) error
	handleParseErr  func(name string, value string, err error) error
}

// NewEnvStruct returns a new EnvStruct.
// The default handle functions are error passthroughs (return the error).
// The default handle functions can be overridden by setting the handle functions.
func NewEnvStruct() *EnvStruct {
	return &EnvStruct{
		variables: []helper{},
		// default functions are just passthroughs
		handleSuccess: func(name string, value reflect.Value) {},
		handleLookupErr: func(name string, err error) error {
			return err
		},
		handleParseErr: func(name string, value string, err error) error {
			return err
		},
	}
}

// validateArgs panics if the name is empty or the pointer is nil.
func validateArgs(pntr any, name string) {
	if name == "" {
		panic(errors.New("name must not be empty"))
	}
	if reflect.ValueOf(pntr).IsNil() {
		panic(errors.New("pntr must not be nil for env: " + name))
	}
}

// AddString adds a string variable to the EnvStruct.
// The pointer must be a non-nil pointer to a string.
func (es *EnvStruct) AddString(pntr *string, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: stringType,
		name:    name,
		pntr:    pntr,
	})
}

// AddInt adds an int variable to the EnvStruct.
// The pointer must be a non-nil pointer to an int.
func (es *EnvStruct) AddInt(pntr *int, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: intType,
		name:    name,
		pntr:    pntr,
	})
}

// AddBool adds a bool variable to the EnvStruct.
// The pointer must be a non-nil pointer to a bool.
func (es *EnvStruct) AddBool(pntr *bool, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: boolType,
		name:    name,
		pntr:    pntr,
	})
}

// AddQuantity adds a resource.Quantity variable to the EnvStruct.
// The pointer must be a non-nil pointer to a resource.Quantity.
func (es *EnvStruct) AddQuantity(pntr *resource.Quantity, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: quantityType,
		name:    name,
		pntr:    pntr,
	})
}

// AddDuration adds a time.Duration variable to the EnvStruct.
// The pointer must be a non-nil pointer to a time.Duration.
func (es *EnvStruct) AddDuration(pntr *time.Duration, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: durationType,
		name:    name,
		pntr:    pntr,
	})
}

// AddStringArray adds a []string variable to the EnvStruct.
// The pointer must be a non-nil pointer to a []string.
func (es *EnvStruct) AddStringArray(pntr *[]string, name string, sep string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: stringArrayType,
		name:    name,
		pntr:    pntr,
		sep:     sep,
	})
}

// AddPath adds a []string variable to the EnvStruct.
// The pointer must be a non-nil pointer to a []string.
func (es *EnvStruct) AddPath(pntr *[]string, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: pathType,
		name:    name,
		pntr:    pntr,
		sep:     string(filepath.ListSeparator),
	})
}

// SetHandleSuccess sets the function to handle what happens when there is a successful lookup and parse.
// Default is a no-op.
func (es *EnvStruct) SetHandleSuccess(f func(name string, value reflect.Value)) {
	es.handleSuccess = f
}

// SetHandleLookupErr sets the function to handle what happens when there is a failed lookup.
// Default is to return an ErrEnvVarNotFound error.
func (es *EnvStruct) SetHandleLookupErr(f func(name string, err error) error) {
	es.handleLookupErr = f
}

// SetHandleParseErr sets the function to handle what happens when there is a failed parse.
// Default is to return an ErrParseEnvVar error.
func (es *EnvStruct) SetHandleParseErr(f func(name string, value string, err error) error) {
	es.handleParseErr = f
}

// DocString returns a string of the documentation for the EnvStruct.
// The documentation is a concatenation of the documentation for each variable.
// Each variable's documentation is it's type, name and valid values.
func (es *EnvStruct) DocString() string {
	// build string from variables
	var b strings.Builder
	for _, v := range es.variables {
		b.WriteString(v.docString() + "\n")
	}
	return b.String()
}

// EnvOverrides overrides the variables added to the EnvStruct with the values from the environment.
// An error is returned if any of the lookups or parses fail.
// Error handling is done by the functions set by SetHandleLookupErr and SetHandleParseErr.
func (es *EnvStruct) EnvOverrides() error {
	var wasErr bool
	for _, v := range es.variables {
		// give helper a copy of the handlers
		es.copyHandlers(&v)
		var err error
		switch v.varType {
		case stringType:
			err = v.lookupString()
		case intType:
			err = v.lookupInt()
		case boolType:
			err = v.lookupBool()
		case quantityType:
			err = v.lookupQuantity()
		case durationType:
			err = v.lookupDuration()
		case stringArrayType:
			err = v.lookupStringArray()
		case pathType:
			err = v.lookupPath()
		default:
			panic("unknown type")
		}
		if err != nil {
			wasErr = true
		}
	}
	if wasErr {
		return errors.New("one or more errors occurred")
	}
	return nil
}

// copyHandlers copies the handlers from the EnvStruct to the helper.
// This is done so that the helper can use the handlers without having to know about the EnvStruct.
func (es *EnvStruct) copyHandlers(h *helper) {
	h.handleSuccess = func() {
		es.handleSuccess(h.name, reflect.ValueOf(h.pntr).Elem())
	}
	h.handleLookupErr = func() error {
		return es.handleLookupErr(h.name, ErrEnvVarNotFound)
	}
	h.handleParseErr = func(failedStr string) error {
		return es.handleParseErr(h.name, failedStr, ErrParseEnvVar)
	}
}
