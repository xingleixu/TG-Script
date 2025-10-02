package vm

import (
	"fmt"
)

// RuntimeError represents a runtime error in the virtual machine
type RuntimeError struct {
	Message string
	PC      int // program counter where error occurred
	Stack   []string // call stack trace
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("Runtime Error: %s", e.Message)
}

// NewRuntimeError creates a new runtime error
func NewRuntimeError(format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		Message: fmt.Sprintf(format, args...),
		PC:      -1,
		Stack:   make([]string, 0),
	}
}

// CompileError represents a compilation error
type CompileError struct {
	Message string
	Line    int
	Column  int
	File    string
}

func (e *CompileError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("Compile Error at %s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("Compile Error at %d:%d: %s", e.Line, e.Column, e.Message)
}

// NewCompileError creates a new compile error
func NewCompileError(message string, line, column int, file string) *CompileError {
	return &CompileError{
		Message: message,
		Line:    line,
		Column:  column,
		File:    file,
	}
}

// VMError represents a virtual machine error
type VMError struct {
	Type    string
	Message string
	Cause   error
}

func (e *VMError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("VM Error [%s]: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("VM Error [%s]: %s", e.Type, e.Message)
}

func (e *VMError) Unwrap() error {
	return e.Cause
}

// NewVMError creates a new VM error
func NewVMError(cause error, format string, args ...interface{}) *VMError {
	return &VMError{
		Type:    "VMError",
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

// NewVMErrorWithType creates a new VM error with specific type
func NewVMErrorWithType(errorType string, cause error, format string, args ...interface{}) *VMError {
	return &VMError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

// IsVMError checks if an error is a VM error
func IsVMError(err error) bool {
	_, ok := err.(*VMError)
	return ok
}

// IsRuntimeError checks if an error is a runtime error
func IsRuntimeError(err error) bool {
	_, ok := err.(*RuntimeError)
	return ok
}

// IsCompileError checks if an error is a compile error
func IsCompileError(err error) bool {
	_, ok := err.(*CompileError)
	return ok
}

// Common VM error types
var (
	ErrStackOverflow     = "StackOverflow"
	ErrStackUnderflow    = "StackUnderflow"
	ErrTypeError         = "TypeError"
	ErrDivisionByZero    = "DivisionByZero"
	ErrUndefinedVariable = "UndefinedVariable"
	ErrInvalidOperation  = "InvalidOperation"
	ErrIndexOutOfBounds  = "IndexOutOfBounds"
	ErrInvalidArguments  = "InvalidArguments"
)