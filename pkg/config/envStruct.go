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

// the envStruct is a middleman between variables retreived from the configuration and the internal config struct

// Configuration is a string, no getting around parsing values from env or flags

// The documentation for what is a valid value for the variable is different
// from the documentation of the internal configuration struct

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
		return fmt.Sprintf("quantity var: %s, allows any valid quantity", h.name)
	case durationType:
		return fmt.Sprintf("duration var: %s, allows any valid duration", h.name)
	case stringArrayType:
		return fmt.Sprintf("string array var: %s, allows any valid string array", h.name)
	case pathType:
		return fmt.Sprintf("path var: %s, allows any valid path", h.name)
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

// Struct is for helping other users to manage their variables and
// the variable's possible overrides based on flags or environment variables
type EnvStruct struct {
	// the variables that are added to the struct
	variables []helper

	// function for handling successful lookups and parses
	handleSuccess func(name string, value reflect.Value)

	// functions for handling errors
	handleLookupErr func(name string, err error) error
	handleParseErr  func(name string, value string, err error) error
}

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

func validateArgs(pntr any, name string) {
	if reflect.ValueOf(pntr).IsNil() {
		panic(errors.New("pntr must not be nil"))
	}
	if name == "" {
		panic(errors.New("name must not be empty"))
	}
}

func (es *EnvStruct) AddString(pntr *string, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: stringType,
		name:    name,
		pntr:    pntr,
	})
}

func (es *EnvStruct) AddInt(pntr *int, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: intType,
		name:    name,
		pntr:    pntr,
	})
}

func (es *EnvStruct) AddBool(pntr *bool, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: boolType,
		name:    name,
		pntr:    pntr,
	})
}

func (es *EnvStruct) AddQuantity(pntr *resource.Quantity, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: quantityType,
		name:    name,
		pntr:    pntr,
	})
}

func (es *EnvStruct) AddDuration(pntr *time.Duration, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: durationType,
		name:    name,
		pntr:    pntr,
	})
}

func (es *EnvStruct) AddStringArray(pntr *[]string, name string, sep string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: stringArrayType,
		name:    name,
		pntr:    pntr,
		sep:     sep,
	})
}

func (es *EnvStruct) AddPath(pntr *[]string, name string) {
	validateArgs(pntr, name)
	es.variables = append(es.variables, helper{
		varType: pathType,
		name:    name,
		pntr:    pntr,
		sep:     string(filepath.ListSeparator),
	})
}

func (es *EnvStruct) SetHandleSuccess(f func(name string, value reflect.Value)) {
	es.handleSuccess = f
}

// we want a method for adding a function to handle what happens when there is a failed lookup
func (es *EnvStruct) SetHandleLookupErr(f func(name string, err error) error) {
	es.handleLookupErr = f
}

// we want a method for adding a function to handle what happens when there is a failed parse
func (es *EnvStruct) SetHandleParseErr(f func(name string, value string, err error) error) {
	es.handleParseErr = f
}

// we want a method for creating a doc string for the varaibles added to the struct
func (es *EnvStruct) DocString() string {
	// build string from variables
	var b strings.Builder
	for _, v := range es.variables {
		b.WriteString(v.docString())
	}
	return b.String()
}

// we want a method for doing the work of parsing the variables
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
