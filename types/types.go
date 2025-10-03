package types

import (
	"fmt"
	"strings"
)

// Type represents a type in the TG-Script type system
type Type interface {
	String() string
	Equals(other Type) bool
	IsAssignableTo(other Type) bool
}

// ============================================================================
// PRIMITIVE TYPES
// ============================================================================

// PrimitiveType represents primitive types
type PrimitiveType struct {
	Kind PrimitiveKind
}

type PrimitiveKind int

const (
	// Basic types
	IntKind PrimitiveKind = iota
	FloatKind
	StringKind
	BooleanKind
	NullKind
	UndefinedKind
	VoidKind
	AnyKind
	
	// Extended numeric types
	Int8Kind
	Int16Kind
	Int32Kind
	Int64Kind
	Float32Kind
	Float64Kind
)

func (p *PrimitiveType) String() string {
	switch p.Kind {
	case IntKind:
		return "int"
	case FloatKind:
		return "float"
	case StringKind:
		return "string"
	case BooleanKind:
		return "boolean"
	case NullKind:
		return "null"
	case UndefinedKind:
		return "undefined"
	case VoidKind:
		return "void"
	case AnyKind:
		return "any"
	case Int8Kind:
		return "int8"
	case Int16Kind:
		return "int16"
	case Int32Kind:
		return "int32"
	case Int64Kind:
		return "int64"
	case Float32Kind:
		return "float32"
	case Float64Kind:
		return "float64"
	default:
		return "unknown"
	}
}

func (p *PrimitiveType) Equals(other Type) bool {
	if otherPrim, ok := other.(*PrimitiveType); ok {
		return p.Kind == otherPrim.Kind
	}
	return false
}

func (p *PrimitiveType) IsAssignableTo(other Type) bool {
	if p.Equals(other) {
		return true
	}
	
	// Any type can be assigned to any other type (TypeScript behavior)
	if p.Kind == AnyKind {
		return true
	}
	
	// Any type can accept any value (TypeScript behavior)
	if otherPrim, ok := other.(*PrimitiveType); ok && otherPrim.Kind == AnyKind {
		return true
	}
	
	// Numeric type compatibility
	if otherPrim, ok := other.(*PrimitiveType); ok {
		return p.isNumericCompatible(otherPrim)
	}
	
	return false
}

func (p *PrimitiveType) isNumericCompatible(other *PrimitiveType) bool {
	// Int can be assigned to float
	if p.Kind == IntKind && other.Kind == FloatKind {
		return true
	}
	
	// Specific numeric types can be assigned to general types
	switch p.Kind {
	case Int8Kind, Int16Kind, Int32Kind, Int64Kind:
		return other.Kind == IntKind || other.Kind == FloatKind
	case Float32Kind, Float64Kind:
		return other.Kind == FloatKind
	}
	
	return false
}

// ============================================================================
// ARRAY TYPE
// ============================================================================

// ArrayType represents array types (T[])
type ArrayType struct {
	ElementType Type
}

func (a *ArrayType) String() string {
	return fmt.Sprintf("%s[]", a.ElementType.String())
}

func (a *ArrayType) Equals(other Type) bool {
	if otherArray, ok := other.(*ArrayType); ok {
		return a.ElementType.Equals(otherArray.ElementType)
	}
	return false
}

func (a *ArrayType) IsAssignableTo(other Type) bool {
	if otherArray, ok := other.(*ArrayType); ok {
		return a.ElementType.IsAssignableTo(otherArray.ElementType)
	}
	return false
}

// ============================================================================
// FUNCTION TYPE
// ============================================================================

// FunctionType represents function types
type FunctionType struct {
	Parameters []Type
	ReturnType Type
	Variadic   bool // true if the function accepts variable number of arguments
}

func (f *FunctionType) String() string {
	var params []string
	for _, param := range f.Parameters {
		params = append(params, param.String())
	}
	if f.Variadic {
		params = append(params, "...")
	}
	return fmt.Sprintf("(%s) => %s", strings.Join(params, ", "), f.ReturnType.String())
}

func (f *FunctionType) Equals(other Type) bool {
	if otherFunc, ok := other.(*FunctionType); ok {
		if len(f.Parameters) != len(otherFunc.Parameters) {
			return false
		}
		
		for i, param := range f.Parameters {
			if !param.Equals(otherFunc.Parameters[i]) {
				return false
			}
		}
		
		return f.ReturnType.Equals(otherFunc.ReturnType)
	}
	return false
}

func (f *FunctionType) IsAssignableTo(other Type) bool {
	if otherFunc, ok := other.(*FunctionType); ok {
		// Function types are contravariant in parameters and covariant in return type
		if len(f.Parameters) != len(otherFunc.Parameters) {
			return false
		}
		
		// Parameters: contravariant (other's params must be assignable to this's params)
		for i, param := range f.Parameters {
			if !otherFunc.Parameters[i].IsAssignableTo(param) {
				return false
			}
		}
		
		// Return type: covariant (this's return must be assignable to other's return)
		return f.ReturnType.IsAssignableTo(otherFunc.ReturnType)
	}
	return false
}

// ============================================================================
// OBJECT TYPES
// ============================================================================

// ObjectType represents an object with properties
type ObjectType struct {
	Properties map[string]Type
}

func (o *ObjectType) String() string {
	if len(o.Properties) == 0 {
		return "object"
	}
	
	var props []string
	for name, typ := range o.Properties {
		props = append(props, fmt.Sprintf("%s: %s", name, typ.String()))
	}
	return fmt.Sprintf("{ %s }", strings.Join(props, ", "))
}

func (o *ObjectType) Equals(other Type) bool {
	if otherObj, ok := other.(*ObjectType); ok {
		if len(o.Properties) != len(otherObj.Properties) {
			return false
		}
		for name, typ := range o.Properties {
			if otherType, exists := otherObj.Properties[name]; !exists || !typ.Equals(otherType) {
				return false
			}
		}
		return true
	}
	return false
}

func (o *ObjectType) IsAssignableTo(other Type) bool {
	if otherObj, ok := other.(*ObjectType); ok {
		// Structural typing: this object is assignable to other if it has all required properties
		for name, expectedType := range otherObj.Properties {
			if actualType, exists := o.Properties[name]; !exists || !actualType.IsAssignableTo(expectedType) {
				return false
			}
		}
		return true
	}
	return false
}

// ============================================================================
// UNION TYPES
// ============================================================================

// UnionType represents a union of multiple types (T | U)
type UnionType struct {
	Types []Type
}

func (u *UnionType) String() string {
	var types []string
	for _, t := range u.Types {
		types = append(types, t.String())
	}
	return strings.Join(types, " | ")
}

func (u *UnionType) Equals(other Type) bool {
	if otherUnion, ok := other.(*UnionType); ok {
		if len(u.Types) != len(otherUnion.Types) {
			return false
		}
		
		// Check if all types match (order doesn't matter)
		for _, t1 := range u.Types {
			found := false
			for _, t2 := range otherUnion.Types {
				if t1.Equals(t2) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

func (u *UnionType) IsAssignableTo(other Type) bool {
	// A union type is assignable to another type if all its member types are assignable
	for _, t := range u.Types {
		if !t.IsAssignableTo(other) {
			return false
		}
	}
	return true
}

// ============================================================================
// PREDEFINED TYPES
// ============================================================================

var (
	IntType       = &PrimitiveType{Kind: IntKind}
	FloatType     = &PrimitiveType{Kind: FloatKind}
	StringType    = &PrimitiveType{Kind: StringKind}
	BooleanType   = &PrimitiveType{Kind: BooleanKind}
	NullType      = &PrimitiveType{Kind: NullKind}
	UndefinedType = &PrimitiveType{Kind: UndefinedKind}
	VoidType      = &PrimitiveType{Kind: VoidKind}
	AnyType       = &PrimitiveType{Kind: AnyKind}

	Int8Type    = &PrimitiveType{Kind: Int8Kind}
	Int16Type   = &PrimitiveType{Kind: Int16Kind}
	Int32Type   = &PrimitiveType{Kind: Int32Kind}
	Int64Type   = &PrimitiveType{Kind: Int64Kind}
	Float32Type = &PrimitiveType{Kind: Float32Kind}
	Float64Type = &PrimitiveType{Kind: Float64Kind}
)

// ============================================================================
// TYPE UTILITIES
// ============================================================================

// NewArrayType creates a new array type
func NewArrayType(elementType Type) *ArrayType {
	return &ArrayType{ElementType: elementType}
}

// NewFunctionType creates a new function type
func NewFunctionType(parameters []Type, returnType Type) *FunctionType {
	return &FunctionType{Parameters: parameters, ReturnType: returnType, Variadic: false}
}

// NewVariadicFunctionType creates a new variadic function type
func NewVariadicFunctionType(parameters []Type, returnType Type) *FunctionType {
	return &FunctionType{Parameters: parameters, ReturnType: returnType, Variadic: true}
}

// NewUnionType creates a new union type
func NewUnionType(types ...Type) *UnionType {
	return &UnionType{Types: types}
}

// IsNumericType checks if a type is numeric
func IsNumericType(t Type) bool {
	if prim, ok := t.(*PrimitiveType); ok {
		switch prim.Kind {
		case IntKind, FloatKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind, Float32Kind, Float64Kind:
			return true
		}
	}
	return false
}

// IsStringType checks if a type is string
func IsStringType(t Type) bool {
	if prim, ok := t.(*PrimitiveType); ok {
		return prim.Kind == StringKind
	}
	return false
}

// IsBooleanType checks if a type is boolean
func IsBooleanType(t Type) bool {
	if prim, ok := t.(*PrimitiveType); ok {
		return prim.Kind == BooleanKind
	}
	return false
}