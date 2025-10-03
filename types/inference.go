package types

import (
	"github.com/xingleixu/TG-Script/ast"
)

// TypeInferrer handles type inference for expressions
type TypeInferrer struct {
	resolver *Resolver
	errors   []error
}

// NewTypeInferrer creates a new type inferrer
func NewTypeInferrer(resolver *Resolver) *TypeInferrer {
	return &TypeInferrer{
		resolver: resolver,
	}
}

// InferType infers the type of an expression
func (ti *TypeInferrer) InferType(expr ast.Expression) Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return IntType
	case *ast.FloatLiteral:
		return FloatType
	case *ast.StringLiteral:
		return StringType
	case *ast.BooleanLiteral:
		return BooleanType
	case *ast.NullLiteral:
		return NullType
	case *ast.UndefinedLiteral:
		return UndefinedType
	case *ast.ArrayLiteral:
		return ti.inferArrayLiteralType(e)
	case *ast.Identifier:
		return ti.inferIdentifierType(e)
	case *ast.BinaryExpression:
		return ti.inferBinaryExpressionType(e)
	case *ast.UnaryExpression:
		return ti.inferUnaryExpressionType(e)
	case *ast.CallExpression:
		return ti.inferCallExpressionType(e)
	case *ast.MemberExpression:
		return ti.inferMemberExpressionType(e)
	case *ast.AssignmentExpression:
		return ti.inferAssignmentExpressionType(e)
	default:
		return UndefinedType
	}
}

// inferArrayLiteralType infers the type of an array literal
func (ti *TypeInferrer) inferArrayLiteralType(expr *ast.ArrayLiteral) Type {
	if len(expr.Elements) == 0 {
		// Empty array - we can't infer the element type
		return NewArrayType(UndefinedType)
	}
	
	// Infer type from first non-null element
	var elementType Type
	for _, element := range expr.Elements {
		if element != nil {
			elementType = ti.InferType(element)
			break
		}
	}
	
	if elementType == nil {
		elementType = UndefinedType
	}
	
	// TODO: In a more sophisticated implementation, we would:
	// 1. Check all elements for type compatibility
	// 2. Find the common supertype if elements have different types
	// 3. Create union types for incompatible types
	
	return NewArrayType(elementType)
}

// inferIdentifierType infers the type of an identifier
func (ti *TypeInferrer) inferIdentifierType(expr *ast.Identifier) Type {
	if symbol, exists := ti.resolver.Lookup(expr.Name); exists {
		return symbol.Type
	}
	return UndefinedType
}

// inferBinaryExpressionType infers the type of a binary expression
func (ti *TypeInferrer) inferBinaryExpressionType(expr *ast.BinaryExpression) Type {
	leftType := ti.InferType(expr.Left)
	rightType := ti.InferType(expr.Right)
	
	switch expr.Operator.String() {
	case "+":
		return ti.inferArithmeticType(leftType, rightType)
	case "-", "*", "/", "%":
		return ti.inferArithmeticType(leftType, rightType)
	case "==", "!=", "<", ">", "<=", ">=":
		return BooleanType
	case "&&", "||":
		return BooleanType
	case "&", "|", "^", "<<", ">>":
		return ti.inferBitwiseType(leftType, rightType)
	default:
		return UndefinedType
	}
}

// inferArithmeticType infers the result type of arithmetic operations
func (ti *TypeInferrer) inferArithmeticType(leftType, rightType Type) Type {
	// If either operand is AnyType, return AnyType (TypeScript behavior)
	if leftType.Equals(AnyType) || rightType.Equals(AnyType) {
		return AnyType
	}
	
	// String concatenation
	if IsStringType(leftType) || IsStringType(rightType) {
		return StringType
	}
	
	// Numeric operations
	if IsNumericType(leftType) && IsNumericType(rightType) {
		// If either operand is float, result is float
		if leftType.Equals(FloatType) || rightType.Equals(FloatType) {
			return FloatType
		}
		// Both are integers
		return IntType
	}
	
	return UndefinedType
}

// inferBitwiseType infers the result type of bitwise operations
func (ti *TypeInferrer) inferBitwiseType(leftType, rightType Type) Type {
	// If either operand is AnyType, return AnyType (TypeScript behavior)
	if leftType.Equals(AnyType) || rightType.Equals(AnyType) {
		return AnyType
	}
	
	if IsNumericType(leftType) && IsNumericType(rightType) {
		return IntType // Bitwise operations always return integers
	}
	return UndefinedType
}

// inferUnaryExpressionType infers the type of a unary expression
func (ti *TypeInferrer) inferUnaryExpressionType(expr *ast.UnaryExpression) Type {
	operandType := ti.InferType(expr.Operand)
	
	switch expr.Operator.String() {
	case "+", "-":
		if IsNumericType(operandType) {
			return operandType
		}
		return UndefinedType
	case "!":
		return BooleanType
	case "~":
		if IsNumericType(operandType) {
			return IntType
		}
		return UndefinedType
	case "++", "--":
		if IsNumericType(operandType) {
			return operandType
		}
		return UndefinedType
	default:
		return UndefinedType
	}
}

// inferCallExpressionType infers the type of a call expression
func (ti *TypeInferrer) inferCallExpressionType(expr *ast.CallExpression) Type {
	calleeType := ti.InferType(expr.Callee)
	
	if funcType, ok := calleeType.(*FunctionType); ok {
		return funcType.ReturnType
	}
	
	return UndefinedType
}

// inferMemberExpressionType infers the type of a member expression
func (ti *TypeInferrer) inferMemberExpressionType(expr *ast.MemberExpression) Type {
	objectType := ti.InferType(expr.Object)
	
	// Array element access
	if arrayType, ok := objectType.(*ArrayType); ok {
		if expr.Computed {
			// arr[index] - return element type
			return arrayType.ElementType
		}
	}
	
	// TODO: Object property access would be handled here
	// For now, we return undefined for unknown member access
	return UndefinedType
}

// inferAssignmentExpressionType infers the type of an assignment expression
func (ti *TypeInferrer) inferAssignmentExpressionType(expr *ast.AssignmentExpression) Type {
	// Assignment expressions return the type of the right-hand side
	return ti.InferType(expr.Right)
}

// GetErrors returns all inference errors
func (ti *TypeInferrer) GetErrors() []error {
	return ti.errors
}

// addError adds an error to the inferrer
func (ti *TypeInferrer) addError(err error) {
	ti.errors = append(ti.errors, err)
}