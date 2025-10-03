package types

import (
	"fmt"

	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// TypeError represents a type checking error
// ErrorCode represents different types of type errors
type ErrorCode string

const (
	UndefinedIdentifierError     ErrorCode = "E001"
	TypeMismatchError            ErrorCode = "E002"
	InvalidOperatorError         ErrorCode = "E003"
	InvalidCallError             ErrorCode = "E004"
	InvalidAssignmentError       ErrorCode = "E005"
	InvalidMemberAccessError     ErrorCode = "E006"
	InvalidArrayElementError     ErrorCode = "E007"
	InvalidReturnTypeError       ErrorCode = "E008"
	InvalidConditionError        ErrorCode = "E009"
	ArgumentCountMismatchError   ErrorCode = "E010"
	ConstReassignmentError       ErrorCode = "E011"
	ArrowFunctionAssignmentError ErrorCode = "E012"
	LetRedeclarationError        ErrorCode = "E013"
)

type TypeError struct {
	Position   lexer.Position
	Message    string
	Code       ErrorCode
	Suggestion string
	Context    string
}

func (e *TypeError) Error() string {
	result := fmt.Sprintf("[%s] Type error at line %d, column %d: %s",
		e.Code, e.Position.Line, e.Position.Column, e.Message)

	if e.Context != "" {
		result += fmt.Sprintf("\n  Context: %s", e.Context)
	}

	if e.Suggestion != "" {
		result += fmt.Sprintf("\n  Suggestion: %s", e.Suggestion)
	}

	return result
}

// TypeChecker performs static type checking
type TypeChecker struct {
	resolver   *Resolver
	inferrer   *TypeInferrer
	errors     []*TypeError
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

	// Check for resolver errors and convert them to TypeError
	if resolverErrors := tc.resolver.GetErrors(); len(resolverErrors) > 0 {
		for _, err := range resolverErrors {
			if typeErr, ok := err.(*TypeError); ok {
				tc.errors = append(tc.errors, typeErr)
			}
		}
	}

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
		var declaredType Type = UndefinedType
		var finalType Type = UndefinedType

		// Check if there's a type annotation
		if declarator.TypeAnnotation != nil {
			declaredType = tc.resolveTypeAnnotation(declarator.TypeAnnotation)
		}

		// Check initializer if present
		if declarator.Init != nil {
			initType := tc.checkExpression(declarator.Init)

			// Check if arrow function is assigned to non-const variable
			if _, isArrowFunction := declarator.Init.(*ast.ArrowFunctionExpression); isArrowFunction {
				if decl.Kind != lexer.CONST {
					if id, ok := declarator.Id.(*ast.Identifier); ok {
						tc.addDetailedError(
							declarator.Init.Pos(),
							fmt.Sprintf("Arrow functions can only be assigned to 'const' variables, not '%s'", decl.Kind.String()),
							ArrowFunctionAssignmentError,
							"Change the variable declaration to 'const' to ensure arrow function immutability",
							fmt.Sprintf("Arrow function assigned to '%s' variable '%s' - this violates immutability requirements",
								decl.Kind.String(), id.Name),
						)
					}
				}
			}

			// If we have both type annotation and initializer, check compatibility
			if declarator.TypeAnnotation != nil {
				if !tc.isAssignable(initType, declaredType) {
					tc.addDetailedError(
						declarator.Init.Pos(),
						fmt.Sprintf("Cannot assign value of type '%s' to variable of type '%s'",
							initType.String(), declaredType.String()),
						TypeMismatchError,
						fmt.Sprintf("Change the initializer to match type '%s' or remove the type annotation to allow type inference",
							declaredType.String()),
						fmt.Sprintf("Variable '%s' is declared with type '%s' but initialized with incompatible type '%s'",
							declarator.Id.String(), declaredType.String(), initType.String()),
					)
				}
				finalType = declaredType
			} else {
				// No type annotation, infer type from initializer
				finalType = initType

			}
		} else if declarator.TypeAnnotation == nil {
			// No type annotation and no initializer - this should be an error in strict mode
			if tc.strictMode {
				tc.addDetailedError(
					declarator.Id.Pos(),
					"Variable declaration must have either a type annotation or an initializer",
					TypeMismatchError,
					"Add a type annotation (e.g., ': string') or provide an initializer (e.g., '= \"value\"')",
					fmt.Sprintf("Variable '%s' has no type information", declarator.Id.String()),
				)
			}
			finalType = declaredType
		} else {
			finalType = declaredType
		}

		// Update variable type in symbol table (it was already defined during resolution)
		if id, ok := declarator.Id.(*ast.Identifier); ok {
			if err := tc.resolver.UpdateType(id.Name, finalType); err != nil {
				// If update fails, try to define it (fallback)
				tc.resolver.DefineWithDeclarationKind(id.Name, finalType, VariableSymbol, decl.Kind, id.Pos())
			}
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
	case *ast.ArrowFunctionExpression:
		return tc.checkArrowFunctionExpression(e)
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
		suggestion := fmt.Sprintf("Use numeric types (int or float) with operator '%s'", operator)
		context := fmt.Sprintf("Left operand: %s, Right operand: %s", leftType.String(), rightType.String())
		tc.addDetailedError(expr.Pos(),
			fmt.Sprintf("Cannot apply operator '%s' to types '%s' and '%s'",
				operator, leftType.String(), rightType.String()),
			InvalidOperatorError,
			suggestion,
			context)
		return UndefinedType

	case "-", "*", "/", "%":
		if !IsNumericType(leftType) || !IsNumericType(rightType) {
			suggestion := fmt.Sprintf("Convert operands to numeric types (int or float) before using '%s'", operator)
			context := fmt.Sprintf("Left operand: %s, Right operand: %s", leftType.String(), rightType.String())
			tc.addDetailedError(expr.Pos(),
				fmt.Sprintf("Cannot apply operator '%s' to non-numeric types '%s' and '%s'",
					operator, leftType.String(), rightType.String()),
				InvalidOperatorError,
				suggestion,
				context)
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
			suggestion := "Use numeric types (int or float) for comparison operations"
			context := fmt.Sprintf("Left operand: %s, Right operand: %s", leftType.String(), rightType.String())
			tc.addDetailedError(expr.Pos(),
				fmt.Sprintf("Cannot compare non-numeric types '%s' and '%s'",
					leftType.String(), rightType.String()),
				InvalidOperatorError,
				suggestion,
				context)
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
			suggestion := fmt.Sprintf("Use numeric types (int or float) with unary operator '%s'", operator)
			context := fmt.Sprintf("Operand type: %s", operandType.String())
			tc.addDetailedError(expr.Pos(),
				fmt.Sprintf("Cannot apply unary operator '%s' to non-numeric type '%s'",
					operator, operandType.String()),
				InvalidOperatorError,
				suggestion,
				context)
			return UndefinedType
		}
		return operandType

	case "!":
		return BooleanType

	case "++", "--":
		if !IsNumericType(operandType) {
			suggestion := fmt.Sprintf("Use numeric types (int or float) with operator '%s'", operator)
			context := fmt.Sprintf("Operand type: %s", operandType.String())
			tc.addDetailedError(expr.Pos(),
				fmt.Sprintf("Cannot apply operator '%s' to non-numeric type '%s'",
					operator, operandType.String()),
				InvalidOperatorError,
				suggestion,
				context)
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
				suggestion := fmt.Sprintf("Provide exactly %d arguments to match function signature", len(funcType.Parameters))
				context := fmt.Sprintf("Function signature requires %d parameters", len(funcType.Parameters))
				tc.addDetailedError(expr.Pos(),
					fmt.Sprintf("Expected %d arguments, got %d",
						len(funcType.Parameters), len(expr.Arguments)),
					ArgumentCountMismatchError,
					suggestion,
					context)
			}
		} else {
			// For variadic functions, check minimum argument count
			if len(expr.Arguments) < len(funcType.Parameters) {
				suggestion := fmt.Sprintf("Provide at least %d arguments for this variadic function", len(funcType.Parameters))
				context := fmt.Sprintf("Variadic function requires minimum %d parameters", len(funcType.Parameters))
				tc.addDetailedError(expr.Pos(),
					fmt.Sprintf("Expected at least %d arguments, got %d",
						len(funcType.Parameters), len(expr.Arguments)),
					ArgumentCountMismatchError,
					suggestion,
					context)
			}
		}

		// Check argument types
		for i, arg := range expr.Arguments {
			argType := tc.checkExpression(arg)

			if i < len(funcType.Parameters) {
				// Check regular parameters
				expectedType := funcType.Parameters[i]
				if !tc.isAssignable(argType, expectedType) {
					suggestion := fmt.Sprintf("Convert argument %d to type '%s' or check function signature", i+1, expectedType.String())
					context := fmt.Sprintf("Function expects parameter %d of type '%s', but got '%s'", i+1, expectedType.String(), argType.String())
					tc.addDetailedError(expr.Pos(),
						fmt.Sprintf("Argument %d: cannot assign type '%s' to parameter of type '%s'",
							i+1, argType.String(), expectedType.String()),
						ArgumentCountMismatchError,
						suggestion,
						context)
				}
			} else if funcType.Variadic {
				// For variadic arguments, we accept any type for now
				// In a more sophisticated implementation, we would check against the variadic parameter type
				continue
			}
		}

		return funcType.ReturnType
	}

	suggestion := "Ensure the expression evaluates to a function before calling it"
	context := fmt.Sprintf("Attempting to call expression of type '%s'", calleeType.String())
	tc.addDetailedError(expr.Pos(),
		fmt.Sprintf("Cannot call non-function type '%s'", calleeType.String()),
		InvalidCallError,
		suggestion,
		context)
	return UndefinedType
}

// checkIdentifier type checks an identifier and reports undefined variables/functions
func (tc *TypeChecker) checkIdentifier(expr *ast.Identifier) Type {
	if symbol, exists := tc.resolver.Lookup(expr.Name); exists {
		return symbol.Type
	}
	// In strict mode, report undefined identifiers as errors
	if tc.strictMode {
		suggestion := fmt.Sprintf("Declare '%s' before using it, or check for typos", expr.Name)
		context := fmt.Sprintf("Identifier '%s' is not defined in the current scope", expr.Name)
		tc.addDetailedError(expr.Pos(),
			fmt.Sprintf("Undefined identifier '%s'", expr.Name),
			UndefinedIdentifierError,
			suggestion,
			context)
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
				suggestion := "Use numeric types (int or float) for array indexing"
				context := fmt.Sprintf("Index type: %s", indexType.String())
				tc.addDetailedError(expr.Pos(),
					fmt.Sprintf("Array index must be numeric, got '%s'", indexType.String()),
					InvalidArrayElementError,
					suggestion,
					context)
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
					suggestion := fmt.Sprintf("Check if property '%s' exists or verify the object type", propIdent.Name)
					context := fmt.Sprintf("Accessing property '%s' on object of type '%s'", propIdent.Name, objectType.String())
					tc.addDetailedError(expr.Pos(),
						fmt.Sprintf("Property '%s' does not exist on object", propIdent.Name),
						InvalidMemberAccessError,
						suggestion,
						context)
				}
			}
		} else {
			// Computed property access like obj[prop]
			propType := tc.checkExpression(expr.Property)
			if !IsStringType(propType) {
				suggestion := "Use string type for object property keys"
				context := fmt.Sprintf("Property key type: %s", propType.String())
				tc.addDetailedError(expr.Pos(),
					fmt.Sprintf("Object property key must be string, got '%s'", propType.String()),
					InvalidMemberAccessError,
					suggestion,
					context)
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

	// Check if we're trying to reassign a const variable
	if id, ok := expr.Left.(*ast.Identifier); ok {
		if symbol, exists := tc.resolver.Lookup(id.Name); exists {
			if symbol.DeclarationKind == lexer.CONST {
				suggestion := "Use 'let' or 'var' instead of 'const' if you need to reassign the variable"
				context := fmt.Sprintf("Variable '%s' was declared with 'const' and cannot be reassigned", id.Name)
				tc.addDetailedError(expr.Pos(),
					fmt.Sprintf("Cannot assign to const variable '%s'", id.Name),
					ConstReassignmentError,
					suggestion,
					context)
				return rightType
			}
		}
	}

	if !tc.isAssignable(rightType, leftType) {
		suggestion := fmt.Sprintf("Convert the value to type '%s' or change the variable type", leftType.String())
		context := fmt.Sprintf("Assigning value of type '%s' to variable of type '%s'", rightType.String(), leftType.String())
		tc.addDetailedError(expr.Pos(),
			fmt.Sprintf("Cannot assign type '%s' to type '%s'",
				rightType.String(), leftType.String()),
			InvalidAssignmentError,
			suggestion,
			context)
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
				suggestion := fmt.Sprintf("Ensure all array elements have the same type '%s'", elementType.String())
				context := fmt.Sprintf("Element %d has type '%s', but expected '%s'", i, elemType.String(), elementType.String())
				tc.addDetailedError(expr.Pos(),
					"Array elements must have the same type",
					TypeMismatchError,
					suggestion,
					context)
				break
			}
		}
	}

	if elementType == nil {
		elementType = UndefinedType
	}

	return NewArrayType(elementType)
}

// checkArrowFunctionExpression type checks an arrow function expression
func (tc *TypeChecker) checkArrowFunctionExpression(expr *ast.ArrowFunctionExpression) Type {

	// Enter function scope
	tc.resolver.EnterScope()
	defer tc.resolver.ExitScope()

	// Process parameters and build parameter types
	var paramTypes []Type
	var paramsNeedInference []int // Track which parameters need type inference

	for i, param := range expr.Parameters {
		var paramType Type = UndefinedType
		if param.TypeAnnotation != nil {
			paramType = tc.resolveTypeAnnotation(param.TypeAnnotation)
		} else {
			paramsNeedInference = append(paramsNeedInference, i)
		}
		paramTypes = append(paramTypes, paramType)
		tc.resolver.Define(param.Name.Name, paramType, ParameterSymbol, param.Name.Pos())
	}

	// Determine return type
	var returnType Type = UndefinedType
	if expr.ReturnType != nil {
		returnType = tc.resolveTypeAnnotation(expr.ReturnType)
	}

	// Perform type inference for parameters that need it first
	if len(paramsNeedInference) > 0 {
		for _, paramIndex := range paramsNeedInference {
			param := expr.Parameters[paramIndex]
			// Try to infer type from usage in the function body
			inferredType := tc.inferParameterType(param.Name.Name, expr.Body)
			if inferredType != UndefinedType {
				paramTypes[paramIndex] = inferredType
				// Update the parameter type in the symbol table
				tc.resolver.UpdateType(param.Name.Name, inferredType)
			} else {
				// Default to int for numeric operations, or keep as undefined
				paramTypes[paramIndex] = IntType
				tc.resolver.UpdateType(param.Name.Name, IntType)
			}
		}
	}

	// Check function body after parameter type inference
	if expr.Body != nil {
		switch body := expr.Body.(type) {
		case *ast.BlockStatement:
			tc.checkBlockStatement(body)

			// For arrow functions with expression bodies (wrapped in BlockStatement with ReturnStatement),
			// we need to infer the return type from the return statement
			if returnType == UndefinedType && len(body.Body) == 1 {
				if returnStmt, ok := body.Body[0].(*ast.ReturnStatement); ok && returnStmt.Argument != nil {
					returnType = tc.checkExpression(returnStmt.Argument)
				}
			}
		case ast.Expression:
			// For expression bodies, check the expression and use its type as return type
			if returnType == UndefinedType {
				returnType = tc.checkExpression(body)
			} else {
				// If return type is explicitly specified, check compatibility
				exprType := tc.checkExpression(body)
				if !returnType.Equals(exprType) {
					suggestion := fmt.Sprintf("Change return type to '%s' or modify expression", exprType.String())
					context := fmt.Sprintf("Expression returns '%s', but function expects '%s'", exprType.String(), returnType.String())
					tc.addDetailedError(expr.Pos(),
						"Arrow function expression type doesn't match declared return type",
						TypeMismatchError,
						suggestion,
						context)
				}
			}
		default:
			// Handle other node types if needed
			if returnType == UndefinedType {
				returnType = UndefinedType
			}
		}
	}

	// Create and return function type
	funcType := &FunctionType{
		Parameters: paramTypes,
		ReturnType: returnType,
		Variadic:   false,
	}
	return funcType
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
		suggestion := "Use boolean expressions in if conditions (e.g., x > 0, x === true)"
		context := fmt.Sprintf("Condition type: %s", condType.String())
		tc.addDetailedError(stmt.Pos(),
			fmt.Sprintf("If condition must be boolean, got '%s'", condType.String()),
			InvalidConditionError,
			suggestion,
			context)
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
		suggestion := "Use boolean expressions in while conditions (e.g., x > 0, x !== null)"
		context := fmt.Sprintf("Condition type: %s", condType.String())
		tc.addDetailedError(stmt.Pos(),
			fmt.Sprintf("While condition must be boolean, got '%s'", condType.String()),
			InvalidConditionError,
			suggestion,
			context)
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
			suggestion := "Use boolean expressions in for conditions (e.g., i < 10, x !== null)"
			context := fmt.Sprintf("Condition type: %s", condType.String())
			tc.addDetailedError(stmt.Pos(),
				fmt.Sprintf("For condition must be boolean, got '%s'", condType.String()),
				InvalidConditionError,
				suggestion,
				context)
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
		case lexer.INT_T:
			return IntType
		case lexer.FLOAT_T:
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
		// Extended numeric types
		case lexer.INT8_T:
			return Int8Type
		case lexer.INT16_T:
			return Int16Type
		case lexer.INT32_T:
			return Int32Type
		case lexer.INT64_T:
			return Int64Type
		case lexer.FLOAT32_T:
			return Float32Type
		case lexer.FLOAT64_T:
			return Float64Type
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

// addError adds a type error with basic information
func (tc *TypeChecker) addError(pos lexer.Position, message string) {
	tc.errors = append(tc.errors, &TypeError{
		Position: pos,
		Message:  message,
		Code:     TypeMismatchError, // Default error code
	})
}

// addDetailedError adds a type error with detailed information
func (tc *TypeChecker) addDetailedError(pos lexer.Position, message string, code ErrorCode, suggestion string, context string) {
	tc.errors = append(tc.errors, &TypeError{
		Position:   pos,
		Message:    message,
		Code:       code,
		Suggestion: suggestion,
		Context:    context,
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

// inferParameterType attempts to infer the type of a parameter from its usage in the function body
func (tc *TypeChecker) inferParameterType(paramName string, body ast.Node) Type {
	// This is a simple implementation that looks for numeric operations
	// In a full implementation, this would be much more sophisticated

	switch node := body.(type) {
	case *ast.BlockStatement:
		for _, stmt := range node.Body {
			if returnStmt, ok := stmt.(*ast.ReturnStatement); ok && returnStmt.Argument != nil {
				return tc.inferParameterTypeFromExpression(paramName, returnStmt.Argument)
			}
		}
	case ast.Expression:
		return tc.inferParameterTypeFromExpression(paramName, node)
	}

	return UndefinedType
}

// inferParameterTypeFromExpression infers parameter type from expression usage
func (tc *TypeChecker) inferParameterTypeFromExpression(paramName string, expr ast.Expression) Type {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		// Check if the parameter is used in a binary expression
		if tc.expressionUsesParameter(e.Left, paramName) || tc.expressionUsesParameter(e.Right, paramName) {
			// For arithmetic operations, assume int
			switch e.Operator.String() {
			case "+", "-", "*", "/", "%":
				return IntType
			case "==", "!=", "<", ">", "<=", ">=":
				return IntType // Comparison operations often use numbers
			}
		}
	case *ast.Identifier:
		if e.Name == paramName {
			// Parameter used directly, can't infer much
			return UndefinedType
		}
	}

	return UndefinedType
}

// expressionUsesParameter checks if an expression uses a specific parameter
func (tc *TypeChecker) expressionUsesParameter(expr ast.Expression, paramName string) bool {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name == paramName
	case *ast.BinaryExpression:
		return tc.expressionUsesParameter(e.Left, paramName) || tc.expressionUsesParameter(e.Right, paramName)
	case *ast.UnaryExpression:
		return tc.expressionUsesParameter(e.Operand, paramName)
	case *ast.CallExpression:
		if tc.expressionUsesParameter(e.Callee, paramName) {
			return true
		}
		for _, arg := range e.Arguments {
			if tc.expressionUsesParameter(arg, paramName) {
				return true
			}
		}
	}

	return false
}
