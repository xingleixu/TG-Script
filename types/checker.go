package types

import (
	"fmt"
	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// TypeError represents a type checking error
type TypeError struct {
	Position lexer.Position
	Message  string
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("Type error at line %d, column %d: %s", 
		e.Position.Line, e.Position.Column, e.Message)
}

// TypeChecker performs static type checking
type TypeChecker struct {
	resolver  *Resolver
	inferrer  *TypeInferrer
	errors    []*TypeError
	strictMode bool
}

// NewTypeChecker creates a new type checker
func NewTypeChecker() *TypeChecker {
	resolver := NewResolver()
	inferrer := NewTypeInferrer(resolver)
	
	return &TypeChecker{
		resolver:   resolver,
		inferrer:   inferrer,
		strictMode: true, // Enable strict mode by default for better error detection
	}
}

// Check performs type checking on a program
func (tc *TypeChecker) Check(program *ast.Program) []*TypeError {
	tc.errors = nil
	
	// First pass: resolve symbols and build symbol table
	tc.resolver.ResolveProgram(program)
	
	// Second pass: type check all statements
	for _, stmt := range program.Body {
		tc.checkStatement(stmt)
	}
	
	return tc.errors
}

// checkStatement type checks a statement
func (tc *TypeChecker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		tc.checkVariableDeclaration(s)
	case *ast.FunctionDeclaration:
		tc.checkFunctionDeclaration(s)
	case *ast.ExpressionStatement:
		tc.checkExpression(s.Expression)
	case *ast.BlockStatement:
		tc.checkBlockStatement(s)
	case *ast.IfStatement:
		tc.checkIfStatement(s)
	case *ast.WhileStatement:
		tc.checkWhileStatement(s)
	case *ast.ForStatement:
		tc.checkForStatement(s)
	case *ast.ReturnStatement:
		tc.checkReturnStatement(s)
	}
}

// checkVariableDeclaration type checks a variable declaration
func (tc *TypeChecker) checkVariableDeclaration(decl *ast.VariableDeclaration) {
	for _, declarator := range decl.Declarations {
		if declarator.Init != nil {
			// Type check the initializer expression
			initType := tc.checkExpression(declarator.Init)
			
			// TODO: Handle type annotations when they're added to VariableDeclarator
			// For now, we just infer the type from the initializer
			_ = initType
		}
	}
}

// checkFunctionDeclaration type checks a function declaration
func (tc *TypeChecker) checkFunctionDeclaration(decl *ast.FunctionDeclaration) {
	// Enter function scope
	tc.resolver.EnterScope()
	defer tc.resolver.ExitScope()
	
	// Add parameters to scope
	for _, param := range decl.Parameters {
		var paramType Type = UndefinedType
		if param.TypeAnnotation != nil {
			paramType = tc.resolveTypeAnnotation(param.TypeAnnotation)
		}
		tc.resolver.Define(param.Name.Name, paramType, ParameterSymbol, param.Name.Pos())
	}
	
	// Check function body
	if decl.Body != nil {
		tc.checkBlockStatement(decl.Body)
	}
	
	// TODO: Check return type compatibility
}

// checkExpression type checks an expression
func (tc *TypeChecker) checkExpression(expr ast.Expression) Type {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return tc.checkBinaryExpression(e)
	case *ast.UnaryExpression:
		return tc.checkUnaryExpression(e)
	case *ast.CallExpression:
		return tc.checkCallExpression(e)
	case *ast.MemberExpression:
		return tc.checkMemberExpression(e)
	case *ast.AssignmentExpression:
		return tc.checkAssignmentExpression(e)
	case *ast.ArrayLiteral:
		return tc.checkArrayLiteral(e)
	case *ast.Identifier:
		return tc.checkIdentifier(e)
	default:
		return tc.inferrer.InferType(expr)
	}
}

// checkBinaryExpression type checks a binary expression
func (tc *TypeChecker) checkBinaryExpression(expr *ast.BinaryExpression) Type {
	leftType := tc.checkExpression(expr.Left)
	rightType := tc.checkExpression(expr.Right)
	
	operator := expr.Operator.String()
	
	// Type compatibility checks
	switch operator {
	case "+":
		// Allow string concatenation or numeric addition
		if IsStringType(leftType) || IsStringType(rightType) {
			return StringType
		}
		if IsNumericType(leftType) && IsNumericType(rightType) {
			if leftType.Equals(FloatType) || rightType.Equals(FloatType) {
				return FloatType
			}
			return IntType
		}
		tc.addError(expr.Pos(), 
			fmt.Sprintf("Cannot apply operator '%s' to types '%s' and '%s'",
				operator, leftType.String(), rightType.String()))
		return UndefinedType
		
	case "-", "*", "/", "%":
		if !IsNumericType(leftType) || !IsNumericType(rightType) {
			tc.addError(expr.Pos(),
				fmt.Sprintf("Cannot apply operator '%s' to non-numeric types '%s' and '%s'",
					operator, leftType.String(), rightType.String()))
			return UndefinedType
		}
		if leftType.Equals(FloatType) || rightType.Equals(FloatType) {
			return FloatType
		}
		return IntType
		
	case "==", "!=":
		// Allow comparison of any types
		return BooleanType
		
	case "<", ">", "<=", ">=":
		if !IsNumericType(leftType) || !IsNumericType(rightType) {
			tc.addError(expr.Pos(),
				fmt.Sprintf("Cannot compare non-numeric types '%s' and '%s'",
					leftType.String(), rightType.String()))
		}
		return BooleanType
		
	case "&&", "||":
		return BooleanType
		
	default:
		return tc.inferrer.InferType(expr)
	}
}

// checkUnaryExpression type checks a unary expression
func (tc *TypeChecker) checkUnaryExpression(expr *ast.UnaryExpression) Type {
	operandType := tc.checkExpression(expr.Operand)
	operator := expr.Operator.String()
	
	switch operator {
	case "+", "-":
		if !IsNumericType(operandType) {
			tc.addError(expr.Pos(),
				fmt.Sprintf("Cannot apply unary operator '%s' to non-numeric type '%s'",
					operator, operandType.String()))
			return UndefinedType
		}
		return operandType
		
	case "!":
		return BooleanType
		
	case "++", "--":
		if !IsNumericType(operandType) {
			tc.addError(expr.Pos(),
				fmt.Sprintf("Cannot apply operator '%s' to non-numeric type '%s'",
					operator, operandType.String()))
			return UndefinedType
		}
		return operandType
		
	default:
		return tc.inferrer.InferType(expr)
	}
}

// checkCallExpression type checks a call expression
func (tc *TypeChecker) checkCallExpression(expr *ast.CallExpression) Type {
	calleeType := tc.checkExpression(expr.Callee)
	
	if funcType, ok := calleeType.(*FunctionType); ok {
		// Check argument count for non-variadic functions
		if !funcType.Variadic {
			if len(expr.Arguments) != len(funcType.Parameters) {
				tc.addError(expr.Pos(),
					fmt.Sprintf("Expected %d arguments, got %d",
						len(funcType.Parameters), len(expr.Arguments)))
			}
		} else {
			// For variadic functions, check minimum argument count
			if len(expr.Arguments) < len(funcType.Parameters) {
				tc.addError(expr.Pos(),
					fmt.Sprintf("Expected at least %d arguments, got %d",
						len(funcType.Parameters), len(expr.Arguments)))
			}
		}
		
		// Check argument types
		for i, arg := range expr.Arguments {
			argType := tc.checkExpression(arg)
			
			if i < len(funcType.Parameters) {
				// Check regular parameters
				expectedType := funcType.Parameters[i]
				if !tc.isAssignable(argType, expectedType) {
					tc.addError(expr.Pos(),
						fmt.Sprintf("Argument %d: cannot assign type '%s' to parameter of type '%s'",
							i+1, argType.String(), expectedType.String()))
				}
			} else if funcType.Variadic {
				// For variadic arguments, we accept any type for now
				// In a more sophisticated implementation, we would check against the variadic parameter type
				continue
			}
		}
		
		return funcType.ReturnType
	}
	
	tc.addError(expr.Pos(),
		fmt.Sprintf("Cannot call non-function type '%s'", calleeType.String()))
	return UndefinedType
}

// checkIdentifier type checks an identifier and reports undefined variables/functions
func (tc *TypeChecker) checkIdentifier(expr *ast.Identifier) Type {
	if symbol, exists := tc.resolver.Lookup(expr.Name); exists {
		return symbol.Type
	}
	
	// In strict mode, report undefined identifiers as errors
	if tc.strictMode {
		tc.addError(expr.Pos(), fmt.Sprintf("Undefined identifier '%s'", expr.Name))
	}
	
	return UndefinedType
}

// checkMemberExpression type checks a member expression
func (tc *TypeChecker) checkMemberExpression(expr *ast.MemberExpression) Type {
	objectType := tc.checkExpression(expr.Object)
	
	if arrayType, ok := objectType.(*ArrayType); ok {
		if expr.Computed {
			// Check index type
			indexType := tc.checkExpression(expr.Property)
			if !IsNumericType(indexType) {
				tc.addError(expr.Pos(),
					fmt.Sprintf("Array index must be numeric, got '%s'", indexType.String()))
			}
			return arrayType.ElementType
		}
	}
	
	// Handle object property access
	if objType, ok := objectType.(*ObjectType); ok {
		if !expr.Computed {
			// Property access like obj.prop
			if propIdent, ok := expr.Property.(*ast.Identifier); ok {
				if propType, exists := objType.Properties[propIdent.Name]; exists {
					return propType
				}
				if tc.strictMode {
					tc.addError(expr.Pos(),
						fmt.Sprintf("Property '%s' does not exist on object", propIdent.Name))
				}
			}
		} else {
			// Computed property access like obj[prop]
			propType := tc.checkExpression(expr.Property)
			if !IsStringType(propType) {
				tc.addError(expr.Pos(),
					fmt.Sprintf("Object property key must be string, got '%s'", propType.String()))
			}
			// For computed access, we can't determine the exact property type at compile time
			return UndefinedType
		}
	}
	
	return UndefinedType
}

// checkAssignmentExpression type checks an assignment expression
func (tc *TypeChecker) checkAssignmentExpression(expr *ast.AssignmentExpression) Type {
	leftType := tc.checkExpression(expr.Left)
	rightType := tc.checkExpression(expr.Right)
	
	if !tc.isAssignable(rightType, leftType) {
		tc.addError(expr.Pos(),
			fmt.Sprintf("Cannot assign type '%s' to type '%s'",
				rightType.String(), leftType.String()))
	}
	
	return rightType
}

// checkArrayLiteral type checks an array literal
func (tc *TypeChecker) checkArrayLiteral(expr *ast.ArrayLiteral) Type {
	if len(expr.Elements) == 0 {
		return NewArrayType(UndefinedType)
	}
	
	// Check all elements and find common type
	var elementType Type
	for i, element := range expr.Elements {
		if element != nil {
			elemType := tc.checkExpression(element)
			if i == 0 {
				elementType = elemType
			} else if !elementType.Equals(elemType) {
				// TODO: In a more sophisticated implementation,
				// we would create union types or find common supertypes
				tc.addError(expr.Pos(),
					"Array elements must have the same type")
				break
			}
		}
	}
	
	if elementType == nil {
		elementType = UndefinedType
	}
	
	return NewArrayType(elementType)
}

// checkBlockStatement type checks a block statement
func (tc *TypeChecker) checkBlockStatement(stmt *ast.BlockStatement) {
	tc.resolver.EnterScope()
	defer tc.resolver.ExitScope()
	
	for _, s := range stmt.Body {
		tc.checkStatement(s)
	}
}

// checkIfStatement type checks an if statement
func (tc *TypeChecker) checkIfStatement(stmt *ast.IfStatement) {
	// Check condition
	condType := tc.checkExpression(stmt.Test)
	if tc.strictMode && !IsBooleanType(condType) {
		tc.addError(stmt.Pos(),
			fmt.Sprintf("If condition must be boolean, got '%s'", condType.String()))
	}
	
	// Check consequent
	tc.checkStatement(stmt.Consequent)
	
	// Check alternate if present
	if stmt.Alternate != nil {
		tc.checkStatement(stmt.Alternate)
	}
}

// checkWhileStatement type checks a while statement
func (tc *TypeChecker) checkWhileStatement(stmt *ast.WhileStatement) {
	// Check condition
	condType := tc.checkExpression(stmt.Test)
	if tc.strictMode && !IsBooleanType(condType) {
		tc.addError(stmt.Pos(),
			fmt.Sprintf("While condition must be boolean, got '%s'", condType.String()))
	}
	
	// Check body
	tc.checkStatement(stmt.Body)
}

// checkForStatement type checks a for statement
func (tc *TypeChecker) checkForStatement(stmt *ast.ForStatement) {
	tc.resolver.EnterScope()
	defer tc.resolver.ExitScope()
	
	// Check init
	if stmt.Init != nil {
		tc.checkStatement(stmt.Init)
	}
	
	// Check test
	if stmt.Test != nil {
		condType := tc.checkExpression(stmt.Test)
		if tc.strictMode && !IsBooleanType(condType) {
			tc.addError(stmt.Pos(),
				fmt.Sprintf("For condition must be boolean, got '%s'", condType.String()))
		}
	}
	
	// Check update
	if stmt.Update != nil {
		tc.checkExpression(stmt.Update)
	}
	
	// Check body
	tc.checkStatement(stmt.Body)
}

// checkReturnStatement type checks a return statement
func (tc *TypeChecker) checkReturnStatement(stmt *ast.ReturnStatement) {
	if stmt.Argument != nil {
		tc.checkExpression(stmt.Argument)
	}
	// TODO: Check return type compatibility with function signature
}

// resolveTypeAnnotation resolves a type annotation to a Type
func (tc *TypeChecker) resolveTypeAnnotation(annotation ast.TypeNode) Type {
	switch t := annotation.(type) {
	case *ast.BasicType:
		switch t.Kind {
		case lexer.NUMBER_T:
			return FloatType
		case lexer.STRING_T:
			return StringType
		case lexer.BOOLEAN_T:
			return BooleanType
		case lexer.VOID:
			return VoidType
		case lexer.NULL:
			return NullType
		case lexer.UNDEFINED:
			return UndefinedType
		default:
			return UndefinedType
		}
	case *ast.ArrayType:
		elementType := tc.resolveTypeAnnotation(t.ElementType)
		return NewArrayType(elementType)
	case *ast.UnionType:
		var types []Type
		for _, typeNode := range t.Types {
			types = append(types, tc.resolveTypeAnnotation(typeNode))
		}
		return NewUnionType(types...)
	default:
		return UndefinedType
	}
}

// isAssignable checks if source type can be assigned to target type
func (tc *TypeChecker) isAssignable(source, target Type) bool {
	// Same type
	if source.Equals(target) {
		return true
	}
	
	// Undefined can be assigned to anything (for now)
	if source.Equals(UndefinedType) {
		return true
	}
	
	// Numeric type compatibility
	if IsNumericType(source) && IsNumericType(target) {
		return true
	}
	
	// Union type handling
	if unionType, ok := target.(*UnionType); ok {
		for _, t := range unionType.Types {
			if tc.isAssignable(source, t) {
				return true
			}
		}
	}
	
	return false
}

// addError adds a type error
func (tc *TypeChecker) addError(pos lexer.Position, message string) {
	tc.errors = append(tc.errors, &TypeError{
		Position: pos,
		Message:  message,
	})
}

// GetErrors returns all type checking errors
func (tc *TypeChecker) GetErrors() []*TypeError {
	return tc.errors
}

// SetStrictMode enables or disables strict type checking
func (tc *TypeChecker) SetStrictMode(strict bool) {
	tc.strictMode = strict
}