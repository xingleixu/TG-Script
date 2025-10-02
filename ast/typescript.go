package ast

import (
	"strings"

	"github.com/tsgo/tg/lexer"
)

// ============================================================================
// TYPESCRIPT TYPE NODES
// ============================================================================

// TypeReference represents a type reference (e.g., string, number, MyClass).
type TypeReference struct {
	TypePos   lexer.Position // position of type name
	Name      *Identifier    // type name
	TypeArgs  []TypeNode     // type arguments (for generics)
}

func (tr *TypeReference) Pos() lexer.Position { return tr.TypePos }
func (tr *TypeReference) End() lexer.Position {
	if len(tr.TypeArgs) > 0 {
		return tr.TypeArgs[len(tr.TypeArgs)-1].End()
	}
	return tr.Name.End()
}
func (tr *TypeReference) String() string {
	result := tr.Name.String()
	if len(tr.TypeArgs) > 0 {
		var args []string
		for _, arg := range tr.TypeArgs {
			args = append(args, arg.String())
		}
		result += "<" + strings.Join(args, ", ") + ">"
	}
	return result
}
func (tr *TypeReference) typeNode() {}

// ArrayType represents an array type (e.g., string[], Array<string>).
type ArrayType struct {
	ElementType TypeNode       // element type
	LBracket    lexer.Position // position of '['
	RBracket    lexer.Position // position of ']'
}

func (at *ArrayType) Pos() lexer.Position { return at.ElementType.Pos() }
func (at *ArrayType) End() lexer.Position { return lexer.Position{
	Line:   at.RBracket.Line,
	Column: at.RBracket.Column + 1,
	Offset: at.RBracket.Offset + 1,
} }
func (at *ArrayType) String() string {
	return at.ElementType.String() + "[]"
}
func (at *ArrayType) typeNode() {}

// UnionType represents a union type (e.g., string | number).
type UnionType struct {
	Types []TypeNode // union member types
}

func (ut *UnionType) Pos() lexer.Position { return ut.Types[0].Pos() }
func (ut *UnionType) End() lexer.Position { return ut.Types[len(ut.Types)-1].End() }
func (ut *UnionType) String() string {
	var types []string
	for _, t := range ut.Types {
		types = append(types, t.String())
	}
	return strings.Join(types, " | ")
}
func (ut *UnionType) typeNode() {}

// IntersectionType represents an intersection type (e.g., A & B).
type IntersectionType struct {
	Types []TypeNode // intersection member types
}

func (it *IntersectionType) Pos() lexer.Position { return it.Types[0].Pos() }
func (it *IntersectionType) End() lexer.Position { return it.Types[len(it.Types)-1].End() }
func (it *IntersectionType) String() string {
	var types []string
	for _, t := range it.Types {
		types = append(types, t.String())
	}
	return strings.Join(types, " & ")
}
func (it *IntersectionType) typeNode() {}

// FunctionType represents a function type (e.g., (x: number) => string).
type FunctionType struct {
	LParen     lexer.Position // position of '('
	Parameters []*Parameter   // parameters
	RParen     lexer.Position // position of ')'
	Arrow      lexer.Position // position of '=>'
	ReturnType TypeNode       // return type
}

func (ft *FunctionType) Pos() lexer.Position { return ft.LParen }
func (ft *FunctionType) End() lexer.Position { return ft.ReturnType.End() }
func (ft *FunctionType) String() string {
	var params []string
	for _, param := range ft.Parameters {
		params = append(params, param.String())
	}
	return "(" + strings.Join(params, ", ") + ") => " + ft.ReturnType.String()
}
func (ft *FunctionType) typeNode() {}

// ObjectType represents an object type (e.g., { x: number; y: string }).
type ObjectType struct {
	LBrace   lexer.Position      // position of '{'
	Members  []*TypeMember       // object members
	RBrace   lexer.Position      // position of '}'
}

func (ot *ObjectType) Pos() lexer.Position { return ot.LBrace }
func (ot *ObjectType) End() lexer.Position { return lexer.Position{
	Line:   ot.RBrace.Line,
	Column: ot.RBrace.Column + 1,
	Offset: ot.RBrace.Offset + 1,
} }
func (ot *ObjectType) String() string {
	result := "{ "
	var members []string
	for _, member := range ot.Members {
		members = append(members, member.String())
	}
	result += strings.Join(members, "; ")
	result += " }"
	return result
}
func (ot *ObjectType) typeNode() {}

// TypeMember represents a member in an object type.
type TypeMember struct {
	Key        Expression // property key
	Type       TypeNode   // property type
	Optional   bool       // true for optional properties
	Readonly   bool       // true for readonly properties
	Computed   bool       // true for computed property names
}

func (tm *TypeMember) Pos() lexer.Position { return tm.Key.Pos() }
func (tm *TypeMember) End() lexer.Position { return tm.Type.End() }
func (tm *TypeMember) String() string {
	result := ""
	if tm.Readonly {
		result += "readonly "
	}
	
	if tm.Computed {
		result += "[" + tm.Key.String() + "]"
	} else {
		result += tm.Key.String()
	}
	
	if tm.Optional {
		result += "?"
	}
	
	result += ": " + tm.Type.String()
	return result
}

// TupleType represents a tuple type (e.g., [string, number]).
type TupleType struct {
	LBracket lexer.Position // position of '['
	Elements []TypeNode     // tuple element types
	RBracket lexer.Position // position of ']'
}

func (tt *TupleType) Pos() lexer.Position { return tt.LBracket }
func (tt *TupleType) End() lexer.Position { return lexer.Position{
	Line:   tt.RBracket.Line,
	Column: tt.RBracket.Column + 1,
	Offset: tt.RBracket.Offset + 1,
} }
func (tt *TupleType) String() string {
	var elements []string
	for _, elem := range tt.Elements {
		elements = append(elements, elem.String())
	}
	return "[" + strings.Join(elements, ", ") + "]"
}
func (tt *TupleType) typeNode() {}

// ============================================================================
// TYPESCRIPT DECLARATIONS
// ============================================================================

// TypeParameter represents a generic type parameter.
type TypeParameter struct {
	Name       *Identifier // parameter name
	Constraint TypeNode    // constraint (extends clause)
	Default    TypeNode    // default type
}

func (tp *TypeParameter) Pos() lexer.Position { return tp.Name.Pos() }
func (tp *TypeParameter) End() lexer.Position {
	if tp.Default != nil {
		return tp.Default.End()
	}
	if tp.Constraint != nil {
		return tp.Constraint.End()
	}
	return tp.Name.End()
}
func (tp *TypeParameter) String() string {
	result := tp.Name.String()
	if tp.Constraint != nil {
		result += " extends " + tp.Constraint.String()
	}
	if tp.Default != nil {
		result += " = " + tp.Default.String()
	}
	return result
}

// InterfaceDeclaration represents an interface declaration.
type InterfaceDeclaration struct {
	InterfacePos   lexer.Position    // position of 'interface'
	Name           *Identifier       // interface name
	TypeParameters []*TypeParameter  // generic type parameters
	Extends        []TypeNode        // extended interfaces
	LBrace         lexer.Position    // position of '{'
	Body           []*TypeMember     // interface members
	RBrace         lexer.Position    // position of '}'
}

func (id *InterfaceDeclaration) Pos() lexer.Position { return id.InterfacePos }
func (id *InterfaceDeclaration) End() lexer.Position { return lexer.Position{
	Line:   id.RBrace.Line,
	Column: id.RBrace.Column + 1,
	Offset: id.RBrace.Offset + 1,
} }
func (id *InterfaceDeclaration) String() string {
	result := "interface " + id.Name.String()
	
	if len(id.TypeParameters) > 0 {
		var params []string
		for _, param := range id.TypeParameters {
			params = append(params, param.String())
		}
		result += "<" + strings.Join(params, ", ") + ">"
	}
	
	if len(id.Extends) > 0 {
		var extends []string
		for _, ext := range id.Extends {
			extends = append(extends, ext.String())
		}
		result += " extends " + strings.Join(extends, ", ")
	}
	
	result += " {\n"
	for _, member := range id.Body {
		result += "  " + member.String() + ";\n"
	}
	result += "}"
	
	return result
}
func (id *InterfaceDeclaration) statementNode()    {}
func (id *InterfaceDeclaration) declarationNode() {}

// TypeAliasDeclaration represents a type alias declaration.
type TypeAliasDeclaration struct {
	TypePos        lexer.Position    // position of 'type'
	Name           *Identifier       // alias name
	TypeParameters []*TypeParameter  // generic type parameters
	Assign         lexer.Position    // position of '='
	Type           TypeNode          // aliased type
}

func (tad *TypeAliasDeclaration) Pos() lexer.Position { return tad.TypePos }
func (tad *TypeAliasDeclaration) End() lexer.Position { return tad.Type.End() }
func (tad *TypeAliasDeclaration) String() string {
	result := "type " + tad.Name.String()
	
	if len(tad.TypeParameters) > 0 {
		var params []string
		for _, param := range tad.TypeParameters {
			params = append(params, param.String())
		}
		result += "<" + strings.Join(params, ", ") + ">"
	}
	
	result += " = " + tad.Type.String()
	return result
}
func (tad *TypeAliasDeclaration) statementNode()    {}
func (tad *TypeAliasDeclaration) declarationNode() {}

// EnumDeclaration represents an enum declaration.
type EnumDeclaration struct {
	EnumPos lexer.Position  // position of 'enum'
	Name    *Identifier     // enum name
	LBrace  lexer.Position  // position of '{'
	Members []*EnumMember   // enum members
	RBrace  lexer.Position  // position of '}'
}

func (ed *EnumDeclaration) Pos() lexer.Position { return ed.EnumPos }
func (ed *EnumDeclaration) End() lexer.Position { return lexer.Position{
	Line:   ed.RBrace.Line,
	Column: ed.RBrace.Column + 1,
	Offset: ed.RBrace.Offset + 1,
} }
func (ed *EnumDeclaration) String() string {
	result := "enum " + ed.Name.String() + " {\n"
	
	for i, member := range ed.Members {
		result += "  " + member.String()
		if i < len(ed.Members)-1 {
			result += ","
		}
		result += "\n"
	}
	
	result += "}"
	return result
}
func (ed *EnumDeclaration) statementNode()    {}
func (ed *EnumDeclaration) declarationNode() {}

// EnumMember represents a member in an enum.
type EnumMember struct {
	Name  *Identifier // member name
	Value Expression  // member value (optional)
}

func (em *EnumMember) Pos() lexer.Position { return em.Name.Pos() }
func (em *EnumMember) End() lexer.Position {
	if em.Value != nil {
		return em.Value.End()
	}
	return em.Name.End()
}
func (em *EnumMember) String() string {
	result := em.Name.String()
	if em.Value != nil {
		result += " = " + em.Value.String()
	}
	return result
}

// ============================================================================
// TYPESCRIPT EXPRESSIONS
// ============================================================================

// TypeAssertion represents a type assertion (e.g., value as Type).
type TypeAssertion struct {
	Expression Expression     // expression being asserted
	AsPos      lexer.Position // position of 'as'
	Type       TypeNode       // target type
}

func (ta *TypeAssertion) Pos() lexer.Position { return ta.Expression.Pos() }
func (ta *TypeAssertion) End() lexer.Position { return ta.Type.End() }
func (ta *TypeAssertion) String() string {
	return ta.Expression.String() + " as " + ta.Type.String()
}
func (ta *TypeAssertion) expressionNode() {}

// NonNullAssertion represents a non-null assertion (e.g., value!).
type NonNullAssertion struct {
	Expression Expression     // expression being asserted
	Bang       lexer.Position // position of '!'
}

func (nna *NonNullAssertion) Pos() lexer.Position { return nna.Expression.Pos() }
func (nna *NonNullAssertion) End() lexer.Position { return lexer.Position{
	Line:   nna.Bang.Line,
	Column: nna.Bang.Column + 1,
	Offset: nna.Bang.Offset + 1,
} }
func (nna *NonNullAssertion) String() string {
	return nna.Expression.String() + "!"
}
func (nna *NonNullAssertion) expressionNode() {}

// ============================================================================
// TYPESCRIPT MODIFIERS
// ============================================================================

// Modifier represents TypeScript modifiers.
type Modifier struct {
	Kind string         // "public", "private", "protected", "static", "readonly", "abstract", etc.
	Pos  lexer.Position // position of modifier
}

func (m *Modifier) String() string {
	return m.Kind
}