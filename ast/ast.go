package ast

import (
	"strings"

	"github.com/xingleixu/TG-Script/lexer"
)

// Node represents a node in the AST.
// All AST nodes implement this interface.
type Node interface {
	// Pos returns the position of the first character belonging to the node.
	Pos() lexer.Position
	// End returns the position of the first character immediately after the node.
	End() lexer.Position
	// String returns a string representation of the node.
	String() string
}

// Expression represents all expression nodes.
type Expression interface {
	Node
	expressionNode()
}

// Statement represents all statement nodes.
type Statement interface {
	Node
	statementNode()
}

// Declaration represents all declaration nodes.
type Declaration interface {
	Statement
	declarationNode()
}

// BindingTarget represents nodes that can be used as binding targets.
type BindingTarget interface {
	Expression
	bindingTarget()
}

// Pattern represents destructuring patterns.
type Pattern interface {
	BindingTarget
	pattern()
}

// TypeNode represents TypeScript type annotations.
type TypeNode interface {
	Node
	typeNode()
}

// ============================================================================
// BASIC NODES
// ============================================================================

// Identifier represents an identifier.
type Identifier struct {
	NamePos lexer.Position // position of the identifier
	Name    string         // identifier name
}

func (i *Identifier) Pos() lexer.Position { return i.NamePos }
func (i *Identifier) End() lexer.Position { return lexer.Position{
	Line:   i.NamePos.Line,
	Column: i.NamePos.Column + len(i.Name),
	Offset: i.NamePos.Offset + len(i.Name),
} }
func (i *Identifier) String() string { return i.Name }
func (i *Identifier) expressionNode() {}
func (i *Identifier) bindingTarget()  {}

// ============================================================================
// LITERALS
// ============================================================================

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	ValuePos lexer.Position // position of the literal
	Value    int64          // the integer value
	Raw      string         // the raw literal string
}

func (il *IntegerLiteral) Pos() lexer.Position { return il.ValuePos }
func (il *IntegerLiteral) End() lexer.Position { return lexer.Position{
	Line:   il.ValuePos.Line,
	Column: il.ValuePos.Column + len(il.Raw),
	Offset: il.ValuePos.Offset + len(il.Raw),
} }
func (il *IntegerLiteral) String() string { return il.Raw }
func (il *IntegerLiteral) expressionNode() {}

// FloatLiteral represents a floating-point literal.
type FloatLiteral struct {
	ValuePos lexer.Position // position of the literal
	Value    float64        // the float value
	Raw      string         // the raw literal string
}

func (fl *FloatLiteral) Pos() lexer.Position { return fl.ValuePos }
func (fl *FloatLiteral) End() lexer.Position { return lexer.Position{
	Line:   fl.ValuePos.Line,
	Column: fl.ValuePos.Column + len(fl.Raw),
	Offset: fl.ValuePos.Offset + len(fl.Raw),
} }
func (fl *FloatLiteral) String() string { return fl.Raw }
func (fl *FloatLiteral) expressionNode() {}

// StringLiteral represents a string literal.
type StringLiteral struct {
	ValuePos lexer.Position // position of the literal
	Value    string         // the string value (unescaped)
	Raw      string         // the raw literal string (with quotes)
}

func (sl *StringLiteral) Pos() lexer.Position { return sl.ValuePos }
func (sl *StringLiteral) End() lexer.Position { return lexer.Position{
	Line:   sl.ValuePos.Line,
	Column: sl.ValuePos.Column + len(sl.Raw),
	Offset: sl.ValuePos.Offset + len(sl.Raw),
} }
func (sl *StringLiteral) String() string { return sl.Raw }
func (sl *StringLiteral) expressionNode() {}

// BooleanLiteral represents a boolean literal.
type BooleanLiteral struct {
	ValuePos lexer.Position // position of the literal
	Value    bool           // the boolean value
	Raw      string         // the raw literal string ("true" or "false")
}

func (bl *BooleanLiteral) Pos() lexer.Position { return bl.ValuePos }
func (bl *BooleanLiteral) End() lexer.Position { return lexer.Position{
	Line:   bl.ValuePos.Line,
	Column: bl.ValuePos.Column + len(bl.Raw),
	Offset: bl.ValuePos.Offset + len(bl.Raw),
} }
func (bl *BooleanLiteral) String() string { return bl.Raw }
func (bl *BooleanLiteral) expressionNode() {}

// NullLiteral represents a null literal.
type NullLiteral struct {
	ValuePos lexer.Position // position of the literal
}

func (nl *NullLiteral) Pos() lexer.Position { return nl.ValuePos }
func (nl *NullLiteral) End() lexer.Position { return lexer.Position{
	Line:   nl.ValuePos.Line,
	Column: nl.ValuePos.Column + 4, // "null"
	Offset: nl.ValuePos.Offset + 4,
} }
func (nl *NullLiteral) String() string { return "null" }
func (nl *NullLiteral) expressionNode() {}

// UndefinedLiteral represents an undefined literal.
type UndefinedLiteral struct {
	ValuePos lexer.Position // position of the literal
}

func (ul *UndefinedLiteral) Pos() lexer.Position { return ul.ValuePos }
func (ul *UndefinedLiteral) End() lexer.Position { return lexer.Position{
	Line:   ul.ValuePos.Line,
	Column: ul.ValuePos.Column + 9, // "undefined"
	Offset: ul.ValuePos.Offset + 9,
} }
func (ul *UndefinedLiteral) String() string { return "undefined" }
func (ul *UndefinedLiteral) expressionNode() {}

// VoidLiteral represents a void literal.
type VoidLiteral struct {
	ValuePos lexer.Position // position of the literal
}

func (vl *VoidLiteral) Pos() lexer.Position { return vl.ValuePos }
func (vl *VoidLiteral) End() lexer.Position {
	return lexer.Position{
		Line:   vl.ValuePos.Line,
		Column: vl.ValuePos.Column + 4, // length of "void"
		Offset: vl.ValuePos.Offset + 4,
	}
}
func (vl *VoidLiteral) String() string { return "void" }
func (vl *VoidLiteral) expressionNode() {}

// ============================================================================
// EXPRESSIONS
// ============================================================================

// BinaryExpression represents a binary expression.
type BinaryExpression struct {
	Left     Expression     // left operand
	OpPos    lexer.Position // position of the operator
	Operator lexer.Token    // operator
	Right    Expression     // right operand
}

func (be *BinaryExpression) Pos() lexer.Position { return be.Left.Pos() }
func (be *BinaryExpression) End() lexer.Position { return be.Right.End() }
func (be *BinaryExpression) String() string {
	leftStr := be.Left.String()
	rightStr := be.Right.String()
	
	// Add parentheses around member expressions
	if _, isMember := be.Left.(*MemberExpression); isMember {
		leftStr = "(" + leftStr + ")"
	}
	if _, isMember := be.Right.(*MemberExpression); isMember {
		rightStr = "(" + rightStr + ")"
	}
	
	return "(" + leftStr + " " + be.Operator.String() + " " + rightStr + ")"
}
func (be *BinaryExpression) expressionNode() {}

// UnaryExpression represents a unary expression.
type UnaryExpression struct {
	OpPos    lexer.Position // position of the operator
	Operator lexer.Token    // operator
	Operand  Expression     // operand
	Postfix  bool           // true if postfix (e.g., x++)
}

func (ue *UnaryExpression) Pos() lexer.Position {
	if ue.Postfix {
		return ue.Operand.Pos()
	}
	return ue.OpPos
}
func (ue *UnaryExpression) End() lexer.Position {
	if ue.Postfix {
		return lexer.Position{
			Line:   ue.OpPos.Line,
			Column: ue.OpPos.Column + len(ue.Operator.String()),
			Offset: ue.OpPos.Offset + len(ue.Operator.String()),
		}
	}
	return ue.Operand.End()
}
func (ue *UnaryExpression) String() string {
	if ue.Postfix {
		return ue.Operand.String() + ue.Operator.String()
	}
	return "(" + ue.Operator.String() + ue.Operand.String() + ")"
}
func (ue *UnaryExpression) expressionNode() {}

// AssignmentExpression represents an assignment expression.
type AssignmentExpression struct {
	Left     Expression     // left-hand side
	OpPos    lexer.Position // position of the operator
	Operator lexer.Token    // assignment operator
	Right    Expression     // right-hand side
}

func (ae *AssignmentExpression) Pos() lexer.Position { return ae.Left.Pos() }
func (ae *AssignmentExpression) End() lexer.Position { return ae.Right.End() }
func (ae *AssignmentExpression) String() string {
	return ae.Left.String() + " " + ae.Operator.String() + " " + ae.Right.String()
}
func (ae *AssignmentExpression) expressionNode() {}

// CallExpression represents a function call.
type CallExpression struct {
	Callee    Expression     // function being called
	LParen    lexer.Position // position of '('
	Arguments []Expression   // arguments
	RParen    lexer.Position // position of ')'
}

func (ce *CallExpression) Pos() lexer.Position { return ce.Callee.Pos() }
func (ce *CallExpression) End() lexer.Position { return lexer.Position{
	Line:   ce.RParen.Line,
	Column: ce.RParen.Column + 1,
	Offset: ce.RParen.Offset + 1,
} }
func (ce *CallExpression) String() string {
	var args []string
	for _, arg := range ce.Arguments {
		argStr := arg.String()
		// Add parentheses around member expressions when used as arguments
		if _, isMember := arg.(*MemberExpression); isMember {
			argStr = "(" + argStr + ")"
		}
		args = append(args, argStr)
	}
	return ce.Callee.String() + "(" + strings.Join(args, ", ") + ")"
}
func (ce *CallExpression) expressionNode() {}

// MemberExpression represents property access (obj.prop or obj[prop]).
type MemberExpression struct {
	Object   Expression     // object being accessed
	Property Expression     // property name
	Computed bool           // true for obj[prop], false for obj.prop
	LBracket lexer.Position // position of '[' (if computed)
	RBracket lexer.Position // position of ']' (if computed)
	Dot      lexer.Position // position of '.' (if not computed)
}

func (me *MemberExpression) Pos() lexer.Position { return me.Object.Pos() }
func (me *MemberExpression) End() lexer.Position {
	if me.Computed {
		return lexer.Position{
			Line:   me.RBracket.Line,
			Column: me.RBracket.Column + 1,
			Offset: me.RBracket.Offset + 1,
		}
	}
	return me.Property.End()
}
func (me *MemberExpression) String() string {
	if me.Computed {
		return me.Object.String() + "[" + me.Property.String() + "]"
	}
	return me.Object.String() + "." + me.Property.String()
}
func (me *MemberExpression) expressionNode() {}

// ConditionalExpression represents a ternary conditional expression (test ? consequent : alternate).
type ConditionalExpression struct {
	Test       Expression     // condition
	Question   lexer.Position // position of '?'
	Consequent Expression     // value if true
	Colon      lexer.Position // position of ':'
	Alternate  Expression     // value if false
}

func (ce *ConditionalExpression) Pos() lexer.Position { return ce.Test.Pos() }
func (ce *ConditionalExpression) End() lexer.Position { return ce.Alternate.End() }
func (ce *ConditionalExpression) String() string {
	return ce.Test.String() + " ? " + ce.Consequent.String() + " : " + ce.Alternate.String()
}
func (ce *ConditionalExpression) expressionNode() {}

// ============================================================================
// ARRAY AND OBJECT LITERALS
// ============================================================================

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	LBracket lexer.Position // position of '['
	Elements []Expression   // array elements (nil for holes)
	RBracket lexer.Position // position of ']'
}

func (al *ArrayLiteral) Pos() lexer.Position { return al.LBracket }
func (al *ArrayLiteral) End() lexer.Position { return lexer.Position{
	Line:   al.RBracket.Line,
	Column: al.RBracket.Column + 1,
	Offset: al.RBracket.Offset + 1,
} }
func (al *ArrayLiteral) String() string {
	var elements []string
	for _, elem := range al.Elements {
		if elem == nil {
			elements = append(elements, "")
		} else {
			elements = append(elements, elem.String())
		}
	}
	return "[" + strings.Join(elements, ", ") + "]"
}
func (al *ArrayLiteral) expressionNode() {}

// Property represents a property in an object literal.
type Property struct {
	Key      Expression     // property key
	Colon    lexer.Position // position of ':'
	Value    Expression     // property value
	Computed bool           // true for [key]: value
	Method   bool           // true for method shorthand
	Shorthand bool          // true for {x} shorthand
}

func (p *Property) Pos() lexer.Position { return p.Key.Pos() }
func (p *Property) End() lexer.Position { return p.Value.End() }
func (p *Property) String() string {
	if p.Shorthand {
		return p.Key.String()
	}
	if p.Computed {
		return "[" + p.Key.String() + "]: " + p.Value.String()
	}
	return p.Key.String() + ": " + p.Value.String()
}

// ObjectLiteral represents an object literal.
type ObjectLiteral struct {
	LBrace     lexer.Position // position of '{'
	Properties []*Property    // object properties
	RBrace     lexer.Position // position of '}'
}

func (ol *ObjectLiteral) Pos() lexer.Position { return ol.LBrace }
func (ol *ObjectLiteral) End() lexer.Position { return lexer.Position{
	Line:   ol.RBrace.Line,
	Column: ol.RBrace.Column + 1,
	Offset: ol.RBrace.Offset + 1,
} }
func (ol *ObjectLiteral) String() string {
	var props []string
	for _, prop := range ol.Properties {
		props = append(props, prop.String())
	}
	return "{" + strings.Join(props, ", ") + "}"
}
func (ol *ObjectLiteral) expressionNode() {}