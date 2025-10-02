package ast

import (
	"strings"

	"github.com/tsgo/tg/lexer"
)

// ============================================================================
// FUNCTION RELATED NODES
// ============================================================================

// Parameter represents a function parameter.
type Parameter struct {
	Name         *Identifier // parameter name
	TypeAnnotation TypeNode  // type annotation (optional)
	DefaultValue Expression  // default value (optional)
	Rest         bool        // true for rest parameters (...args)
}

func (p *Parameter) Pos() lexer.Position { return p.Name.Pos() }
func (p *Parameter) End() lexer.Position {
	if p.DefaultValue != nil {
		return p.DefaultValue.End()
	}
	if p.TypeAnnotation != nil {
		return p.TypeAnnotation.End()
	}
	return p.Name.End()
}
func (p *Parameter) String() string {
	result := ""
	if p.Rest {
		result += "..."
	}
	result += p.Name.String()
	if p.TypeAnnotation != nil {
		result += ": " + p.TypeAnnotation.String()
	}
	if p.DefaultValue != nil {
		result += " = " + p.DefaultValue.String()
	}
	return result
}

// FunctionExpression represents a function expression.
type FunctionExpression struct {
	FunctionPos lexer.Position // position of 'function'
	Name        *Identifier    // function name (optional)
	LParen      lexer.Position // position of '('
	Parameters  []*Parameter   // parameters
	RParen      lexer.Position // position of ')'
	ReturnType  TypeNode       // return type annotation (optional)
	Body        *BlockStatement // function body
	Async       bool           // true for async functions
	Generator   bool           // true for generator functions
}

func (fe *FunctionExpression) Pos() lexer.Position { return fe.FunctionPos }
func (fe *FunctionExpression) End() lexer.Position { return fe.Body.End() }
func (fe *FunctionExpression) String() string {
	result := ""
	if fe.Async {
		result += "async "
	}
	result += "function"
	if fe.Generator {
		result += "*"
	}
	if fe.Name != nil {
		result += " " + fe.Name.String()
	}
	
	var params []string
	for _, param := range fe.Parameters {
		params = append(params, param.String())
	}
	result += "(" + strings.Join(params, ", ") + ")"
	
	if fe.ReturnType != nil {
		result += ": " + fe.ReturnType.String()
	}
	
	result += " " + fe.Body.String()
	return result
}
func (fe *FunctionExpression) expressionNode() {}

// FunctionDeclaration represents a function declaration.
type FunctionDeclaration struct {
	FunctionPos lexer.Position // position of 'function'
	Name        *Identifier    // function name
	LParen      lexer.Position // position of '('
	Parameters  []*Parameter   // parameters
	RParen      lexer.Position // position of ')'
	ReturnType  TypeNode       // return type annotation (optional)
	Body        *BlockStatement // function body
	Async       bool           // true for async functions
	Generator   bool           // true for generator functions
}

func (fd *FunctionDeclaration) Pos() lexer.Position { return fd.FunctionPos }
func (fd *FunctionDeclaration) End() lexer.Position { return fd.Body.End() }
func (fd *FunctionDeclaration) String() string {
	result := ""
	if fd.Async {
		result += "async "
	}
	result += "function"
	if fd.Generator {
		result += "*"
	}
	result += " " + fd.Name.String()
	
	var params []string
	for _, param := range fd.Parameters {
		params = append(params, param.String())
	}
	result += "(" + strings.Join(params, ", ") + ")"
	
	if fd.ReturnType != nil {
		result += ": " + fd.ReturnType.String()
	}
	
	result += " " + fd.Body.String()
	return result
}
func (fd *FunctionDeclaration) statementNode()    {}
func (fd *FunctionDeclaration) declarationNode() {}

// ArrowFunctionExpression represents an arrow function.
type ArrowFunctionExpression struct {
	LParen     lexer.Position // position of '(' (optional for single param)
	Parameters []*Parameter   // parameters
	RParen     lexer.Position // position of ')' (optional for single param)
	Arrow      lexer.Position // position of '=>'
	Body       Node           // function body (expression or block)
	ReturnType TypeNode       // return type annotation (optional)
	Async      bool           // true for async arrow functions
}

func (afe *ArrowFunctionExpression) Pos() lexer.Position {
	if len(afe.Parameters) > 0 {
		return afe.Parameters[0].Pos()
	}
	return afe.LParen
}
func (afe *ArrowFunctionExpression) End() lexer.Position { return afe.Body.End() }
func (afe *ArrowFunctionExpression) String() string {
	result := ""
	if afe.Async {
		result += "async "
	}
	
	if len(afe.Parameters) == 1 && afe.LParen.Line == 0 {
		// Single parameter without parentheses
		result += afe.Parameters[0].String()
	} else {
		// Multiple parameters or parentheses
		var params []string
		for _, param := range afe.Parameters {
			params = append(params, param.String())
		}
		result += "(" + strings.Join(params, ", ") + ")"
	}
	
	if afe.ReturnType != nil {
		result += ": " + afe.ReturnType.String()
	}
	
	result += " => " + afe.Body.String()
	return result
}
func (afe *ArrowFunctionExpression) expressionNode() {}

// ============================================================================
// CLASS RELATED NODES
// ============================================================================

// MethodDefinition represents a method in a class.
type MethodDefinition struct {
	Key        Expression      // method name
	Value      *FunctionExpression // method function
	Kind       string          // "method", "constructor", "get", "set"
	Static     bool            // true for static methods
	Computed   bool            // true for computed property names
	Async      bool            // true for async methods
	Generator  bool            // true for generator methods
}

func (md *MethodDefinition) Pos() lexer.Position { return md.Key.Pos() }
func (md *MethodDefinition) End() lexer.Position { return md.Value.End() }
func (md *MethodDefinition) String() string {
	result := ""
	if md.Static {
		result += "static "
	}
	if md.Async {
		result += "async "
	}
	if md.Kind == "get" {
		result += "get "
	} else if md.Kind == "set" {
		result += "set "
	}
	
	if md.Computed {
		result += "[" + md.Key.String() + "]"
	} else {
		result += md.Key.String()
	}
	
	if md.Generator {
		result += "*"
	}
	
	// Extract parameters and body from the function
	var params []string
	for _, param := range md.Value.Parameters {
		params = append(params, param.String())
	}
	result += "(" + strings.Join(params, ", ") + ")"
	
	if md.Value.ReturnType != nil {
		result += ": " + md.Value.ReturnType.String()
	}
	
	result += " " + md.Value.Body.String()
	return result
}

// PropertyDefinition represents a property in a class.
type PropertyDefinition struct {
	Key              Expression     // property name
	Value            Expression     // property value (optional)
	TypeAnnotation   TypeNode       // type annotation (optional)
	Static           bool           // true for static properties
	Computed         bool           // true for computed property names
	Readonly         bool           // true for readonly properties
}

func (pd *PropertyDefinition) Pos() lexer.Position { return pd.Key.Pos() }
func (pd *PropertyDefinition) End() lexer.Position {
	if pd.Value != nil {
		return pd.Value.End()
	}
	if pd.TypeAnnotation != nil {
		return pd.TypeAnnotation.End()
	}
	return pd.Key.End()
}
func (pd *PropertyDefinition) String() string {
	result := ""
	if pd.Static {
		result += "static "
	}
	if pd.Readonly {
		result += "readonly "
	}
	
	if pd.Computed {
		result += "[" + pd.Key.String() + "]"
	} else {
		result += pd.Key.String()
	}
	
	if pd.TypeAnnotation != nil {
		result += ": " + pd.TypeAnnotation.String()
	}
	
	if pd.Value != nil {
		result += " = " + pd.Value.String()
	}
	
	return result
}

// ClassExpression represents a class expression.
type ClassExpression struct {
	ClassPos   lexer.Position        // position of 'class'
	Name       *Identifier           // class name (optional)
	SuperClass Expression            // superclass (optional)
	LBrace     lexer.Position        // position of '{'
	Body       []Node                // class body (methods and properties)
	RBrace     lexer.Position        // position of '}'
}

func (ce *ClassExpression) Pos() lexer.Position { return ce.ClassPos }
func (ce *ClassExpression) End() lexer.Position { return lexer.Position{
	Line:   ce.RBrace.Line,
	Column: ce.RBrace.Column + 1,
	Offset: ce.RBrace.Offset + 1,
} }
func (ce *ClassExpression) String() string {
	result := "class"
	if ce.Name != nil {
		result += " " + ce.Name.String()
	}
	if ce.SuperClass != nil {
		result += " extends " + ce.SuperClass.String()
	}
	result += " {\n"
	
	for _, member := range ce.Body {
		result += "  " + member.String() + "\n"
	}
	
	result += "}"
	return result
}
func (ce *ClassExpression) expressionNode() {}

// ClassDeclaration represents a class declaration.
type ClassDeclaration struct {
	ClassPos   lexer.Position        // position of 'class'
	Name       *Identifier           // class name
	SuperClass Expression            // superclass (optional)
	LBrace     lexer.Position        // position of '{'
	Body       []Node                // class body (methods and properties)
	RBrace     lexer.Position        // position of '}'
}

func (cd *ClassDeclaration) Pos() lexer.Position { return cd.ClassPos }
func (cd *ClassDeclaration) End() lexer.Position { return lexer.Position{
	Line:   cd.RBrace.Line,
	Column: cd.RBrace.Column + 1,
	Offset: cd.RBrace.Offset + 1,
} }
func (cd *ClassDeclaration) String() string {
	result := "class " + cd.Name.String()
	if cd.SuperClass != nil {
		result += " extends " + cd.SuperClass.String()
	}
	result += " {\n"
	
	for _, member := range cd.Body {
		result += "  " + member.String() + "\n"
	}
	
	result += "}"
	return result
}
func (cd *ClassDeclaration) statementNode()    {}
func (cd *ClassDeclaration) declarationNode() {}