package vm

import (
	"fmt"
	"strconv"
	"strings"
)

// ValueType represents the type of a value in the virtual machine
type ValueType int

const (
	TypeNil ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeArray
	TypeObject
	TypeFunction
	TypeNativeFunction
	TypeUpvalue
)

// Value represents a value in the virtual machine
type Value struct {
	Type ValueType
	Data interface{}
}

// Common value constants
var (
	NilValue   = Value{Type: TypeNil, Data: nil}
	TrueValue  = Value{Type: TypeBool, Data: true}
	FalseValue = Value{Type: TypeBool, Data: false}
)

// NewIntValue creates a new integer value
func NewIntValue(i int64) Value {
	return Value{Type: TypeInt, Data: i}
}

// NewFloatValue creates a new float value
func NewFloatValue(f float64) Value {
	return Value{Type: TypeFloat, Data: f}
}

// NewStringValue creates a new string value
func NewStringValue(s string) Value {
	return Value{Type: TypeString, Data: s}
}

// NewBoolValue creates a new boolean value
func NewBoolValue(b bool) Value {
	if b {
		return TrueValue
	}
	return FalseValue
}

// NewArrayValue creates a new array value
func NewArrayValue(arr *Array) Value {
	return Value{Type: TypeArray, Data: arr}
}

// NewObjectValue creates a new object value
func NewObjectValue(obj *Object) Value {
	return Value{Type: TypeObject, Data: obj}
}

// NewFunctionValue creates a new function value
func NewFunctionValue(fn *Function) Value {
	return Value{Type: TypeFunction, Data: fn}
}

// NewNativeFunctionValue creates a new native function value
func NewNativeFunctionValue(fn *NativeFunction) Value {
	return Value{Type: TypeNativeFunction, Data: fn}
}

// IsNil returns true if the value is nil
func (v Value) IsNil() bool {
	return v.Type == TypeNil
}

// IsBool returns true if the value is a boolean
func (v Value) IsBool() bool {
	return v.Type == TypeBool
}

// IsInt returns true if the value is an integer
func (v Value) IsInt() bool {
	return v.Type == TypeInt
}

// IsFloat returns true if the value is a float
func (v Value) IsFloat() bool {
	return v.Type == TypeFloat
}

// IsNumber returns true if the value is a number (int or float)
func (v Value) IsNumber() bool {
	return v.Type == TypeInt || v.Type == TypeFloat
}

// IsString returns true if the value is a string
func (v Value) IsString() bool {
	return v.Type == TypeString
}

// IsArray returns true if the value is an array
func (v Value) IsArray() bool {
	return v.Type == TypeArray
}

// IsObject returns true if the value is an object
func (v Value) IsObject() bool {
	return v.Type == TypeObject
}

// IsFunction returns true if the value is a function
func (v Value) IsFunction() bool {
	return v.Type == TypeFunction || v.Type == TypeNativeFunction
}

// IsCallable returns true if the value can be called
func (v Value) IsCallable() bool {
	return v.IsFunction()
}

// ToBool converts the value to a boolean
func (v Value) ToBool() bool {
	switch v.Type {
	case TypeNil:
		return false
	case TypeBool:
		return v.Data.(bool)
	case TypeInt:
		return v.Data.(int64) != 0
	case TypeFloat:
		return v.Data.(float64) != 0.0
	case TypeString:
		return v.Data.(string) != ""
	default:
		return true // objects, arrays, functions are truthy
	}
}

// ToInt converts the value to an integer
func (v Value) ToInt() (int64, bool) {
	switch v.Type {
	case TypeInt:
		return v.Data.(int64), true
	case TypeFloat:
		f := v.Data.(float64)
		return int64(f), f == float64(int64(f))
	case TypeString:
		if i, err := strconv.ParseInt(v.Data.(string), 10, 64); err == nil {
			return i, true
		}
		return 0, false
	case TypeBool:
		if v.Data.(bool) {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// ToFloat converts the value to a float
func (v Value) ToFloat() (float64, bool) {
	switch v.Type {
	case TypeFloat:
		return v.Data.(float64), true
	case TypeInt:
		return float64(v.Data.(int64)), true
	case TypeString:
		if f, err := strconv.ParseFloat(v.Data.(string), 64); err == nil {
			return f, true
		}
		return 0.0, false
	case TypeBool:
		if v.Data.(bool) {
			return 1.0, true
		}
		return 0.0, true
	default:
		return 0.0, false
	}
}

// ToString converts the value to a string
func (v Value) ToString() string {
	switch v.Type {
	case TypeNil:
		return "nil"
	case TypeBool:
		if v.Data.(bool) {
			return "true"
		}
		return "false"
	case TypeInt:
		return strconv.FormatInt(v.Data.(int64), 10)
	case TypeFloat:
		return strconv.FormatFloat(v.Data.(float64), 'g', -1, 64)
	case TypeString:
		return v.Data.(string)
	case TypeArray:
		arr := v.Data.(*Array)
		var parts []string
		for i := 0; i < arr.Length(); i++ {
			if val, ok := arr.Get(i); ok {
				parts = append(parts, val.ToString())
			} else {
				parts = append(parts, "nil")
			}
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case TypeObject:
		obj := v.Data.(*Object)
		var parts []string
		for key, val := range obj.Properties {
			parts = append(parts, fmt.Sprintf("%s: %s", key, val.ToString()))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case TypeFunction:
		fn := v.Data.(*Function)
		return fmt.Sprintf("function<%s>", fn.Name)
	case TypeNativeFunction:
		fn := v.Data.(*NativeFunction)
		return fmt.Sprintf("native_function<%s>", fn.Name)
	default:
		return fmt.Sprintf("unknown_type<%d>", v.Type)
	}
}

// TypeName returns the name of the value's type
func (v Value) TypeName() string {
	switch v.Type {
	case TypeNil:
		return "nil"
	case TypeBool:
		return "boolean"
	case TypeInt:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	case TypeFunction:
		return "function"
	case TypeNativeFunction:
		return "native_function"
	case TypeUpvalue:
		return "upvalue"
	default:
		return "unknown"
	}
}

// Equals checks if two values are equal
func (v Value) Equals(other Value) bool {
	if v.Type != other.Type {
		return false
	}

	switch v.Type {
	case TypeNil:
		return true
	case TypeBool:
		return v.Data.(bool) == other.Data.(bool)
	case TypeInt:
		return v.Data.(int64) == other.Data.(int64)
	case TypeFloat:
		return v.Data.(float64) == other.Data.(float64)
	case TypeString:
		return v.Data.(string) == other.Data.(string)
	case TypeArray:
		return v.Data.(*Array) == other.Data.(*Array) // reference equality
	case TypeObject:
		return v.Data.(*Object) == other.Data.(*Object) // reference equality
	case TypeFunction:
		return v.Data.(*Function) == other.Data.(*Function) // reference equality
	case TypeNativeFunction:
		return v.Data.(*NativeFunction) == other.Data.(*NativeFunction) // reference equality
	default:
		return false
	}
}

// Compare compares two values (-1: less, 0: equal, 1: greater)
func (v Value) Compare(other Value) (int, bool) {
	// Try numeric comparison first
	if v.IsNumber() && other.IsNumber() {
		vf, _ := v.ToFloat()
		of, _ := other.ToFloat()
		if vf < of {
			return -1, true
		} else if vf > of {
			return 1, true
		} else {
			return 0, true
		}
	}

	// String comparison
	if v.IsString() && other.IsString() {
		vs := v.Data.(string)
		os := other.Data.(string)
		if vs < os {
			return -1, true
		} else if vs > os {
			return 1, true
		} else {
			return 0, true
		}
	}

	// Boolean comparison
	if v.IsBool() && other.IsBool() {
		vb := v.Data.(bool)
		ob := other.Data.(bool)
		if !vb && ob {
			return -1, true
		} else if vb && !ob {
			return 1, true
		} else {
			return 0, true
		}
	}

	return 0, false // incomparable types
}

// Array represents an array value
type Array struct {
	Elements []Value
}

// NewArray creates a new array with the given capacity
func NewArray(capacity int) *Array {
	return &Array{
		Elements: make([]Value, 0, capacity),
	}
}

// Length returns the length of the array
func (a *Array) Length() int {
	return len(a.Elements)
}

// Get returns the element at the given index
func (a *Array) Get(index int) (Value, bool) {
	if index < 0 || index >= len(a.Elements) {
		return NilValue, false
	}
	return a.Elements[index], true
}

// Set sets the element at the given index
func (a *Array) Set(index int, value Value) bool {
	if index < 0 {
		return false
	}
	
	// Extend array if necessary
	for index >= len(a.Elements) {
		a.Elements = append(a.Elements, NilValue)
	}
	
	a.Elements[index] = value
	return true
}

// Push appends a value to the end of the array
func (a *Array) Push(value Value) {
	a.Elements = append(a.Elements, value)
}

// Pop removes and returns the last element
func (a *Array) Pop() (Value, bool) {
	if len(a.Elements) == 0 {
		return NilValue, false
	}
	
	last := a.Elements[len(a.Elements)-1]
	a.Elements = a.Elements[:len(a.Elements)-1]
	return last, true
}

// Object represents an object value
type Object struct {
	Properties map[string]Value
	Prototype  *Object
}

// NewObject creates a new object
func NewObject() *Object {
	return &Object{
		Properties: make(map[string]Value),
		Prototype:  nil,
	}
}

// Get returns the property value for the given key
func (o *Object) Get(key string) (Value, bool) {
	if val, ok := o.Properties[key]; ok {
		return val, true
	}
	
	// Check prototype chain
	if o.Prototype != nil {
		return o.Prototype.Get(key)
	}
	
	return NilValue, false
}

// Set sets the property value for the given key
func (o *Object) Set(key string, value Value) {
	o.Properties[key] = value
}

// Has checks if the object has the given property
func (o *Object) Has(key string) bool {
	if _, ok := o.Properties[key]; ok {
		return true
	}
	
	// Check prototype chain
	if o.Prototype != nil {
		return o.Prototype.Has(key)
	}
	
	return false
}

// Delete removes the property with the given key
func (o *Object) Delete(key string) bool {
	if _, ok := o.Properties[key]; ok {
		delete(o.Properties, key)
		return true
	}
	return false
}

// Keys returns all property keys
func (o *Object) Keys() []string {
	keys := make([]string, 0, len(o.Properties))
	for key := range o.Properties {
		keys = append(keys, key)
	}
	return keys
}