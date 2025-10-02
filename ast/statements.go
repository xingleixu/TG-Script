package ast

import (
	"strings"

	"github.com/xingleixu/TG-Script/lexer"
)

// ============================================================================
// STATEMENTS
// ============================================================================

// Program represents the root node of an AST.
type Program struct {
	Body []Statement // top-level statements
}

func (p *Program) Pos() lexer.Position {
	if len(p.Body) > 0 {
		return p.Body[0].Pos()
	}
	return lexer.Position{}
}
func (p *Program) End() lexer.Position {
	if len(p.Body) > 0 {
		return p.Body[len(p.Body)-1].End()
	}
	return lexer.Position{}
}
func (p *Program) String() string {
	var stmts []string
	for _, stmt := range p.Body {
		stmts = append(stmts, stmt.String())
	}
	return strings.Join(stmts, "")
}

// BlockStatement represents a block statement.
type BlockStatement struct {
	LBrace lexer.Position // position of '{'
	Body   []Statement    // statements in the block
	RBrace lexer.Position // position of '}'
}

func (bs *BlockStatement) Pos() lexer.Position { return bs.LBrace }
func (bs *BlockStatement) End() lexer.Position { return lexer.Position{
	Line:   bs.RBrace.Line,
	Column: bs.RBrace.Column + 1,
	Offset: bs.RBrace.Offset + 1,
} }
func (bs *BlockStatement) String() string {
	var stmts []string
	for _, stmt := range bs.Body {
		stmts = append(stmts, stmt.String())
	}
	return "{\n" + strings.Join(stmts, "\n") + "\n}"
}
func (bs *BlockStatement) statementNode() {}

// ExpressionStatement represents an expression statement.
type ExpressionStatement struct {
	Expression Expression     // the expression
	Semicolon  lexer.Position // position of ';' (optional)
}

func (es *ExpressionStatement) Pos() lexer.Position { return es.Expression.Pos() }
func (es *ExpressionStatement) End() lexer.Position {
	if es.Semicolon.Line > 0 {
		return lexer.Position{
			Line:   es.Semicolon.Line,
			Column: es.Semicolon.Column + 1,
			Offset: es.Semicolon.Offset + 1,
		}
	}
	return es.Expression.End()
}
func (es *ExpressionStatement) String() string {
	if es.Expression == nil {
		return ""
	}
	return es.Expression.String()
}
func (es *ExpressionStatement) statementNode() {}

// ============================================================================
// VARIABLE DECLARATIONS
// ============================================================================

// VariableDeclarator represents a single variable declarator.
type VariableDeclarator struct {
	Id             BindingTarget // variable name/pattern
	TypeAnnotation TypeNode      // type annotation (optional)
	Init           Expression    // initializer (optional)
}

func (vd *VariableDeclarator) Pos() lexer.Position { return vd.Id.Pos() }
func (vd *VariableDeclarator) End() lexer.Position {
	if vd.Init != nil {
		return vd.Init.End()
	}
	return vd.Id.End()
}
func (vd *VariableDeclarator) String() string {
	result := vd.Id.String()
	if vd.TypeAnnotation != nil {
		result += ": " + vd.TypeAnnotation.String()
	}
	if vd.Init != nil {
		result += " = " + vd.Init.String()
	}
	return result
}

// VariableDeclaration represents a variable declaration.
type VariableDeclaration struct {
	DeclPos      lexer.Position         // position of 'let', 'const', or 'var'
	Kind         lexer.Token            // LET, CONST, or VAR
	Declarations []*VariableDeclarator  // variable declarators
	Semicolon    lexer.Position         // position of ';' (optional)
}

func (vd *VariableDeclaration) Pos() lexer.Position { return vd.DeclPos }
func (vd *VariableDeclaration) End() lexer.Position {
	if vd.Semicolon.Line > 0 {
		return lexer.Position{
			Line:   vd.Semicolon.Line,
			Column: vd.Semicolon.Column + 1,
			Offset: vd.Semicolon.Offset + 1,
		}
	}
	if len(vd.Declarations) > 0 {
		return vd.Declarations[len(vd.Declarations)-1].End()
	}
	return vd.DeclPos
}
func (vd *VariableDeclaration) String() string {
	var decls []string
	for _, decl := range vd.Declarations {
		decls = append(decls, decl.String())
	}
	return vd.Kind.String() + " " + strings.Join(decls, ", ") + ";"
}
func (vd *VariableDeclaration) statementNode()    {}
func (vd *VariableDeclaration) declarationNode() {}

// ============================================================================
// CONTROL FLOW STATEMENTS
// ============================================================================

// IfStatement represents an if statement.
type IfStatement struct {
	IfPos       lexer.Position // position of 'if'
	LParen      lexer.Position // position of '('
	Test        Expression     // condition
	RParen      lexer.Position // position of ')'
	Consequent  Statement      // then branch
	ElsePos     lexer.Position // position of 'else' (optional)
	Alternate   Statement      // else branch (optional)
}

func (is *IfStatement) Pos() lexer.Position { return is.IfPos }
func (is *IfStatement) End() lexer.Position {
	if is.Alternate != nil {
		return is.Alternate.End()
	}
	return is.Consequent.End()
}
func (is *IfStatement) String() string {
	result := "if (" + is.Test.String() + ") " + is.Consequent.String()
	if is.Alternate != nil {
		result += " else " + is.Alternate.String()
	}
	return result
}
func (is *IfStatement) statementNode() {}

// WhileStatement represents a while loop.
type WhileStatement struct {
	WhilePos lexer.Position // position of 'while'
	LParen   lexer.Position // position of '('
	Test     Expression     // condition
	RParen   lexer.Position // position of ')'
	Body     Statement      // loop body
}

func (ws *WhileStatement) Pos() lexer.Position { return ws.WhilePos }
func (ws *WhileStatement) End() lexer.Position { return ws.Body.End() }
func (ws *WhileStatement) String() string {
	return "while (" + ws.Test.String() + ") " + ws.Body.String()
}
func (ws *WhileStatement) statementNode() {}

// ForStatement represents a for loop.
type ForStatement struct {
	ForPos lexer.Position // position of 'for'
	LParen lexer.Position // position of '('
	Init   Statement      // initialization (optional)
	Test   Expression     // condition (optional)
	Update Expression     // update (optional)
	RParen lexer.Position // position of ')'
	Body   Statement      // loop body
}

func (fs *ForStatement) Pos() lexer.Position { return fs.ForPos }
func (fs *ForStatement) End() lexer.Position { return fs.Body.End() }
func (fs *ForStatement) String() string {
	init := ""
	if fs.Init != nil {
		init = fs.Init.String()
	}
	test := ""
	if fs.Test != nil {
		test = fs.Test.String()
	}
	update := ""
	if fs.Update != nil {
		update = fs.Update.String()
	}
	return "for (" + init + "; " + test + "; " + update + ") " + fs.Body.String()
}
func (fs *ForStatement) statementNode() {}

// ForInStatement represents a for-in loop.
type ForInStatement struct {
	ForPos lexer.Position // position of 'for'
	LParen lexer.Position // position of '('
	Left   BindingTarget  // loop variable
	InPos  lexer.Position // position of 'in'
	Right  Expression     // object to iterate
	RParen lexer.Position // position of ')'
	Body   Statement      // loop body
}

func (fis *ForInStatement) Pos() lexer.Position { return fis.ForPos }
func (fis *ForInStatement) End() lexer.Position { return fis.Body.End() }
func (fis *ForInStatement) String() string {
	return "for (" + fis.Left.String() + " in " + fis.Right.String() + ") " + fis.Body.String()
}
func (fis *ForInStatement) statementNode() {}

// ForOfStatement represents a for-of loop.
type ForOfStatement struct {
	ForPos lexer.Position // position of 'for'
	LParen lexer.Position // position of '('
	Left   BindingTarget  // loop variable
	OfPos  lexer.Position // position of 'of'
	Right  Expression     // iterable to iterate
	RParen lexer.Position // position of ')'
	Body   Statement      // loop body
}

func (fos *ForOfStatement) Pos() lexer.Position { return fos.ForPos }
func (fos *ForOfStatement) End() lexer.Position { return fos.Body.End() }
func (fos *ForOfStatement) String() string {
	return "for (" + fos.Left.String() + " of " + fos.Right.String() + ") " + fos.Body.String()
}
func (fos *ForOfStatement) statementNode() {}

// ============================================================================
// JUMP STATEMENTS
// ============================================================================

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	ReturnPos lexer.Position // position of 'return'
	Argument  Expression     // return value (optional)
	Semicolon lexer.Position // position of ';' (optional)
}

func (rs *ReturnStatement) Pos() lexer.Position { return rs.ReturnPos }
func (rs *ReturnStatement) End() lexer.Position {
	if rs.Semicolon.Line > 0 {
		return lexer.Position{
			Line:   rs.Semicolon.Line,
			Column: rs.Semicolon.Column + 1,
			Offset: rs.Semicolon.Offset + 1,
		}
	}
	if rs.Argument != nil {
		return rs.Argument.End()
	}
	return lexer.Position{
		Line:   rs.ReturnPos.Line,
		Column: rs.ReturnPos.Column + 6, // "return"
		Offset: rs.ReturnPos.Offset + 6,
	}
}
func (rs *ReturnStatement) String() string {
	if rs.Argument != nil {
		return "return " + rs.Argument.String() + ";"
	}
	return "return;"
}
func (rs *ReturnStatement) statementNode() {}

// BreakStatement represents a break statement.
type BreakStatement struct {
	BreakPos  lexer.Position // position of 'break'
	Label     *Identifier    // label (optional)
	Semicolon lexer.Position // position of ';' (optional)
}

func (bs *BreakStatement) Pos() lexer.Position { return bs.BreakPos }
func (bs *BreakStatement) End() lexer.Position {
	if bs.Semicolon.Line > 0 {
		return lexer.Position{
			Line:   bs.Semicolon.Line,
			Column: bs.Semicolon.Column + 1,
			Offset: bs.Semicolon.Offset + 1,
		}
	}
	if bs.Label != nil {
		return bs.Label.End()
	}
	return lexer.Position{
		Line:   bs.BreakPos.Line,
		Column: bs.BreakPos.Column + 5, // "break"
		Offset: bs.BreakPos.Offset + 5,
	}
}
func (bs *BreakStatement) String() string {
	if bs.Label != nil {
		return "break " + bs.Label.String() + ";"
	}
	return "break;"
}
func (bs *BreakStatement) statementNode() {}

// ContinueStatement represents a continue statement.
type ContinueStatement struct {
	ContinuePos lexer.Position // position of 'continue'
	Label       *Identifier    // label (optional)
	Semicolon   lexer.Position // position of ';' (optional)
}

func (cs *ContinueStatement) Pos() lexer.Position { return cs.ContinuePos }
func (cs *ContinueStatement) End() lexer.Position {
	if cs.Semicolon.Line > 0 {
		return lexer.Position{
			Line:   cs.Semicolon.Line,
			Column: cs.Semicolon.Column + 1,
			Offset: cs.Semicolon.Offset + 1,
		}
	}
	if cs.Label != nil {
		return cs.Label.End()
	}
	return lexer.Position{
		Line:   cs.ContinuePos.Line,
		Column: cs.ContinuePos.Column + 8, // "continue"
		Offset: cs.ContinuePos.Offset + 8,
	}
}
func (cs *ContinueStatement) String() string {
	if cs.Label != nil {
		return "continue " + cs.Label.String() + ";"
	}
	return "continue;"
}
func (cs *ContinueStatement) statementNode() {}

// ============================================================================
// OTHER STATEMENTS
// ============================================================================

// EmptyStatement represents an empty statement (just a semicolon).
type EmptyStatement struct {
	Semicolon lexer.Position // position of ';'
}

func (es *EmptyStatement) Pos() lexer.Position { return es.Semicolon }
func (es *EmptyStatement) End() lexer.Position { return lexer.Position{
	Line:   es.Semicolon.Line,
	Column: es.Semicolon.Column + 1,
	Offset: es.Semicolon.Offset + 1,
} }
func (es *EmptyStatement) String() string { return ";" }
func (es *EmptyStatement) statementNode() {}

// LabeledStatement represents a labeled statement.
type LabeledStatement struct {
	Label     *Identifier    // label
	Colon     lexer.Position // position of ':'
	Statement Statement      // labeled statement
}

func (ls *LabeledStatement) Pos() lexer.Position { return ls.Label.Pos() }
func (ls *LabeledStatement) End() lexer.Position { return ls.Statement.End() }
func (ls *LabeledStatement) String() string {
	return ls.Label.String() + ": " + ls.Statement.String()
}
func (ls *LabeledStatement) statementNode() {}