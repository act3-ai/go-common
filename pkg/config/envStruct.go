package config

import (
	"errors"
	"os"
	"path/filepath"
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
}

// lookup funcs for each type
func (h *helper) lookupString() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	constType := h.pntr.(*string)
	*constType = envVal
	return nil
}

func (h *helper) lookupInt() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	parsedVal, err := strconv.Atoi(envVal)
	if err != nil {
		return ErrParseEnvVar
	}
	constType := h.pntr.(*int)
	*constType = parsedVal
	return nil
}

func (h *helper) lookupBool() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	parsedVal, err := strconv.ParseBool(envVal)
	if err != nil {
		return ErrParseEnvVar
	}
	constType := h.pntr.(*bool)
	*constType = parsedVal
	return nil
}

func (h *helper) lookupQuantity() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	parsedVal, err := resource.ParseQuantity(envVal)
	if err != nil {
		return ErrParseEnvVar
	}
	constType := h.pntr.(*resource.Quantity)
	*constType = parsedVal
	return nil
}

func (h *helper) lookupDuration() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	parsedVal, err := time.ParseDuration(envVal)
	if err != nil {
		return ErrParseEnvVar
	}
	constType := h.pntr.(*time.Duration)
	*constType = parsedVal
	return nil
}

func (h *helper) lookupStringArray() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	parsedVal := strings.Split(envVal, h.sep)
	constType := h.pntr.(*[]string)
	*constType = parsedVal
	return nil
}

func (h *helper) lookupPath() error {
	envVal, ok := os.LookupEnv(h.name)
	if !ok {
		return ErrEnvVarNotFound
	}
	parsedVal := strings.Split(envVal, h.sep)
	constType := h.pntr.(*[]string)
	*constType = parsedVal
	return nil
}

// Struct is for helping other users to manage their variables and
// the variable's possible overrides based on flags or environment variables
type EnvStruct struct {
	// the variables that are added to the struct
	variables []helper

	// function for handling successful lookups and parses
	handleSuccess func()

	// functions for handling errors
	handleLookupErr func(err error) error
	handleParseErr  func(err error) error
}

func NewEnvStruct() *EnvStruct {
	return &EnvStruct{
		variables: []helper{},
		// default functions are just passthroughs
		handleSuccess: func() {},
		handleLookupErr: func(err error) error {
			return err
		},
		handleParseErr: func(err error) error {
			return err
		},
	}
}

// We want methods for adding in any variable type to the struct

func (es *EnvStruct) AddString(pntr *string, name string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{stringType, name, pntr, ""})
}

func (es *EnvStruct) AddInt(pntr *int, name string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{intType, name, pntr, ""})
}

func (es *EnvStruct) AddBool(pntr *bool, name string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{boolType, name, pntr, ""})
}

func (es *EnvStruct) AddQuantity(pntr *resource.Quantity, name string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{quantityType, name, pntr, ""})
}

func (es *EnvStruct) AddDuration(pntr *time.Duration, name string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{durationType, name, pntr, ""})
}

func (es *EnvStruct) AddStringArray(pntr *[]string, name string, sep string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{stringArrayType, name, pntr, sep})
}

func (es *EnvStruct) AddPath(pntr *[]string, name string) {
	if name == "" {
		panic("name must not be empty")
	}
	es.variables = append(es.variables, helper{pathType, name, pntr, string(filepath.ListSeparator)})
}

func (es *EnvStruct) handle(err error) error {
	if err == nil {
		es.handleSuccess()
		return nil
	}
	if err == ErrEnvVarNotFound {
		return es.handleLookupErr(err)
	}
	if err == ErrParseEnvVar {
		return es.handleParseErr(err)
	}
	panic("unknown error")
}

func (es *EnvStruct) SetHandleSuccess(f func()) {
	es.handleSuccess = f
}

// we want a method for adding a function to handle what happens when there is a failed lookup
func (es *EnvStruct) SetHandleLookupErr(f func(err error) error) {
	es.handleLookupErr = f
}

// we want a method for adding a function to handle what happens when there is a failed parse
func (es *EnvStruct) SetHandleParseErr(f func(err error) error) {
	es.handleParseErr = f
}

// we want a method for creating a doc string for the varaibles added to the struct
func (es *EnvStruct) DocString() string {
	return ""
}

// we want a method for doing the work of parsing the variables
func (es *EnvStruct) EnvOverrides() error {
	var wasErr bool
	for _, v := range es.variables {
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
		if anyErr := es.handle(err); anyErr != nil {
			wasErr = true
		}
	}
	if wasErr {
		return errors.New("one or more errors occurred")
	}
	return nil
}
