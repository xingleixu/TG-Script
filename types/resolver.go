package types

import (
	"fmt"

	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name            string
	Type            Type
	Kind            SymbolKind
	DeclarationKind lexer.Token // LET, CONST, VAR for variables
	Position        lexer.Position
	Scope           *Scope
}

type SymbolKind int

const (
	VariableSymbol SymbolKind = iota
	FunctionSymbol
	ParameterSymbol
	TypeSymbol
)

func (s SymbolKind) String() string {
	switch s {
	case VariableSymbol:
		return "variable"
	case FunctionSymbol:
		return "function"
	case ParameterSymbol:
		return "parameter"
	case TypeSymbol:
		return "type"
	default:
		return "unknown"
	}
}

// Scope represents a lexical scope
type Scope struct {
	Parent   *Scope
	Symbols  map[string]*Symbol
	Children []*Scope
}

// NewScope creates a new scope
func NewScope(parent *Scope) *Scope {
	scope := &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
	}
	
	if parent != nil {
		parent.Children = append(parent.Children, scope)
	}
	
	return scope
}

// Define defines a symbol in the current scope
func (s *Scope) Define(name string, symbol *Symbol) error {
	if _, exists := s.Symbols[name]; exists {
		return fmt.Errorf("symbol '%s' already defined in this scope", name)
	}
	
	symbol.Scope = s
	s.Symbols[name] = symbol
	return nil
}

// Lookup looks up a symbol in the current scope and parent scopes
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	if symbol, exists := s.Symbols[name]; exists {
		return symbol, true
	}
	
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	
	return nil, false
}

// LookupLocal looks up a symbol only in the current scope
func (s *Scope) LookupLocal(name string) (*Symbol, bool) {
	symbol, exists := s.Symbols[name]
	return symbol, exists
}

// Resolver handles symbol resolution and scope management
type Resolver struct {
	currentScope *Scope
	globalScope  *Scope
	errors       []error
}

// NewResolver creates a new resolver
func NewResolver() *Resolver {
	globalScope := NewScope(nil)
	
	// Define built-in types and functions
	resolver := &Resolver{
		currentScope: globalScope,
		globalScope:  globalScope,
	}
	
	resolver.defineBuiltins()
	return resolver
}

// defineBuiltins defines built-in symbols
func (r *Resolver) defineBuiltins() {
	// Built-in functions
	builtins := map[string]Type{
		"print":  NewVariadicFunctionType([]Type{}, VoidType), // print accepts any number of arguments of any type
		"len":    NewFunctionType([]Type{NewArrayType(StringType)}, IntType),
		"typeof": NewFunctionType([]Type{StringType}, StringType),
	}
	
	for name, typ := range builtins {
		symbol := &Symbol{
			Name: name,
			Type: typ,
			Kind: FunctionSymbol,
		}
		r.globalScope.Define(name, symbol)
	}
	
	// Define console object with log method
	consoleType := &ObjectType{
		Properties: map[string]Type{
			"log": NewVariadicFunctionType([]Type{}, VoidType), // console.log accepts any number of arguments
		},
	}
	
	consoleSymbol := &Symbol{
		Name: "console",
		Type: consoleType,
		Kind: VariableSymbol,
	}
	r.globalScope.Define("console", consoleSymbol)
}

// EnterScope creates and enters a new scope
func (r *Resolver) EnterScope() {
	r.currentScope = NewScope(r.currentScope)
}

// ExitScope exits the current scope
func (r *Resolver) ExitScope() {
	if r.currentScope.Parent != nil {
		r.currentScope = r.currentScope.Parent
	}
}

// Define defines a symbol in the current scope
func (r *Resolver) Define(name string, typ Type, kind SymbolKind, pos lexer.Position) error {
	return r.DefineWithDeclarationKind(name, typ, kind, lexer.ILLEGAL, pos)
}

// DefineWithDeclarationKind defines a symbol with declaration kind (const, let, var)
func (r *Resolver) DefineWithDeclarationKind(name string, typ Type, kind SymbolKind, declKind lexer.Token, pos lexer.Position) error {
	symbol := &Symbol{
		Name:            name,
		Type:            typ,
		Kind:            kind,
		DeclarationKind: declKind,
		Position:        pos,
	}
	
	err := r.currentScope.Define(name, symbol)
	if err != nil {
		r.addError(err)
	}
	
	return err
}

// Lookup looks up a symbol
func (r *Resolver) Lookup(name string) (*Symbol, bool) {
	return r.currentScope.Lookup(name)
}

// LookupLocal looks up a symbol only in the current scope
func (r *Resolver) LookupLocal(name string) (*Symbol, bool) {
	return r.currentScope.LookupLocal(name)
}

// UpdateType updates the type of an existing symbol
func (r *Resolver) UpdateType(name string, typ Type) error {
	if symbol, exists := r.currentScope.Lookup(name); exists {
		symbol.Type = typ
		return nil
	}
	return fmt.Errorf("symbol '%s' not found", name)
}

// ResolveProgram resolves symbols in a program
func (r *Resolver) ResolveProgram(program *ast.Program) error {
	r.errors = nil
	
	for _, stmt := range program.Body {
		r.resolveStatement(stmt)
	}
	
	if len(r.errors) > 0 {
		return fmt.Errorf("resolution failed with %d errors", len(r.errors))
	}
	
	return nil
}

// resolveStatement resolves a statement
func (r *Resolver) resolveStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		r.resolveVariableDeclaration(s)
	case *ast.FunctionDeclaration:
		r.resolveFunctionDeclaration(s)
	case *ast.ExpressionStatement:
		r.resolveExpression(s.Expression)
	case *ast.BlockStatement:
		r.resolveBlockStatement(s)
	case *ast.IfStatement:
		r.resolveIfStatement(s)
	case *ast.WhileStatement:
		r.resolveWhileStatement(s)
	case *ast.ForStatement:
		r.resolveForStatement(s)
	case *ast.ReturnStatement:
		r.resolveReturnStatement(s)
	}
}

// resolveVariableDeclaration resolves a variable declaration
func (r *Resolver) resolveVariableDeclaration(stmt *ast.VariableDeclaration) {
	for _, decl := range stmt.Declarations {
		// Resolve initializer first
		if decl.Init != nil {
			r.resolveExpression(decl.Init)
		}
		
		// For now, we'll use a simple approach for variable names
		// In a full implementation, we'd need to handle destructuring patterns
		if id, ok := decl.Id.(*ast.Identifier); ok {
			// Check for let redeclaration in the same scope
			if stmt.Kind == lexer.LET {
				if symbol, exists := r.currentScope.LookupLocal(id.Name); exists {
					// Only report error if the existing symbol is also a let variable
					if symbol.DeclarationKind == lexer.LET {
						typeErr := &TypeError{
							Position:   id.NamePos,
							Message:    fmt.Sprintf("Identifier '%s' has already been declared", id.Name),
							Code:       LetRedeclarationError,
							Suggestion: "Use a different variable name or remove the duplicate declaration",
							Context:    fmt.Sprintf("Previous declaration was at line %d", symbol.Position.Line),
						}
						r.addError(typeErr)
						continue // Skip defining this variable
					}
				}
			}
			
			r.DefineWithDeclarationKind(id.Name, UndefinedType, VariableSymbol, stmt.Kind, id.NamePos)
		}
	}
}

// resolveFunctionDeclaration resolves a function declaration
func (r *Resolver) resolveFunctionDeclaration(stmt *ast.FunctionDeclaration) {
	// Resolve parameter types
	var paramTypes []Type
	for _, param := range stmt.Parameters {
		var paramType Type = AnyType // Default to AnyType for unannotated parameters
		if param.TypeAnnotation != nil {
			paramType = r.resolveTypeAnnotation(param.TypeAnnotation)
		}
		paramTypes = append(paramTypes, paramType)
	}
	
	// Resolve return type
	var returnType Type = VoidType
	if stmt.ReturnType != nil {
		returnType = r.resolveTypeAnnotation(stmt.ReturnType)
	}
	
	// Define function in current scope
	funcType := NewFunctionType(paramTypes, returnType)
	r.Define(stmt.Name.Name, funcType, FunctionSymbol, stmt.Name.NamePos)
	
	// Enter function scope
	r.EnterScope()
	
	// Define parameters with their resolved types
	for i, param := range stmt.Parameters {
		r.Define(param.Name.Name, paramTypes[i], ParameterSymbol, param.Name.NamePos)
	}
	
	// Resolve function body
	r.resolveBlockStatement(stmt.Body)
	
	// Exit function scope
	r.ExitScope()
}

// resolveExpression resolves an expression
func (r *Resolver) resolveExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		r.resolveIdentifier(e)
	case *ast.CallExpression:
		r.resolveCallExpression(e)
	case *ast.MemberExpression:
		r.resolveMemberExpression(e)
	case *ast.BinaryExpression:
		r.resolveBinaryExpression(e)
	case *ast.UnaryExpression:
		r.resolveUnaryExpression(e)
	case *ast.AssignmentExpression:
		r.resolveAssignmentExpression(e)
	case *ast.ArrayLiteral:
		r.resolveArrayLiteral(e)
	}
}

// resolveIdentifier resolves an identifier
func (r *Resolver) resolveIdentifier(expr *ast.Identifier) {
	// Note: We don't report undefined identifier errors here because
	// the TypeChecker handles this with more detailed error messages
	_, _ = r.Lookup(expr.Name)
}

// resolveCallExpression resolves a call expression
func (r *Resolver) resolveCallExpression(expr *ast.CallExpression) {
	r.resolveExpression(expr.Callee)
	
	for _, arg := range expr.Arguments {
		r.resolveExpression(arg)
	}
}

// resolveMemberExpression resolves a member expression
func (r *Resolver) resolveMemberExpression(expr *ast.MemberExpression) {
	r.resolveExpression(expr.Object)
	
	if expr.Computed {
		r.resolveExpression(expr.Property)
	}
}

// resolveBinaryExpression resolves a binary expression
func (r *Resolver) resolveBinaryExpression(expr *ast.BinaryExpression) {
	r.resolveExpression(expr.Left)
	r.resolveExpression(expr.Right)
}

// resolveUnaryExpression resolves a unary expression
func (r *Resolver) resolveUnaryExpression(expr *ast.UnaryExpression) {
	r.resolveExpression(expr.Operand)
}

// resolveAssignmentExpression resolves an assignment expression
func (r *Resolver) resolveAssignmentExpression(expr *ast.AssignmentExpression) {
	r.resolveExpression(expr.Right)
	r.resolveExpression(expr.Left)
}

// resolveArrayLiteral resolves an array literal
func (r *Resolver) resolveArrayLiteral(expr *ast.ArrayLiteral) {
	for _, element := range expr.Elements {
		if element != nil {
			r.resolveExpression(element)
		}
	}
}

// resolveBlockStatement resolves a block statement
func (r *Resolver) resolveBlockStatement(stmt *ast.BlockStatement) {
	r.EnterScope()
	
	for _, s := range stmt.Body {
		r.resolveStatement(s)
	}
	
	r.ExitScope()
}

// resolveIfStatement resolves an if statement
func (r *Resolver) resolveIfStatement(stmt *ast.IfStatement) {
	r.resolveExpression(stmt.Test)
	r.resolveStatement(stmt.Consequent)
	
	if stmt.Alternate != nil {
		r.resolveStatement(stmt.Alternate)
	}
}

// resolveWhileStatement resolves a while statement
func (r *Resolver) resolveWhileStatement(stmt *ast.WhileStatement) {
	r.resolveExpression(stmt.Test)
	r.resolveStatement(stmt.Body)
}

// resolveForStatement resolves a for statement
func (r *Resolver) resolveForStatement(stmt *ast.ForStatement) {
	r.EnterScope()
	
	if stmt.Init != nil {
		r.resolveStatement(stmt.Init)
	}
	
	if stmt.Test != nil {
		r.resolveExpression(stmt.Test)
	}
	
	if stmt.Update != nil {
		r.resolveExpression(stmt.Update)
	}
	
	r.resolveStatement(stmt.Body)
	
	r.ExitScope()
}

// resolveReturnStatement resolves a return statement
func (r *Resolver) resolveReturnStatement(stmt *ast.ReturnStatement) {
	if stmt.Argument != nil {
		r.resolveExpression(stmt.Argument)
	}
}

// addError adds an error to the resolver
func (r *Resolver) addError(err error) {
	r.errors = append(r.errors, err)
}

// GetErrors returns all resolution errors
func (r *Resolver) GetErrors() []error {
	return r.errors
}

// GetGlobalScope returns the global scope
func (r *Resolver) GetGlobalScope() *Scope {
	return r.globalScope
}

// resolveTypeAnnotation resolves a type annotation to a Type
func (r *Resolver) resolveTypeAnnotation(annotation ast.TypeNode) Type {
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
		elementType := r.resolveTypeAnnotation(t.ElementType)
		return NewArrayType(elementType)
	case *ast.UnionType:
		var types []Type
		for _, typeNode := range t.Types {
			types = append(types, r.resolveTypeAnnotation(typeNode))
		}
		return NewUnionType(types...)
	default:
		return UndefinedType
	}
}