package lexer

import "strconv"

// Token represents a token type in TG-Script
type Token int

// Position represents the position of a token in the source code
type Position struct {
	Line   int // line number (1-based)
	Column int // column number (1-based)
	Offset int // byte offset (0-based)
}

// TokenInfo contains token information including position
type TokenInfo struct {
	Type     Token
	Literal  string
	Position Position
}

// List of tokens - organized by categories for better maintainability
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT

	// Literals
	literal_beg
	IDENT     // identifiers: variable names, function names, etc.
	INT       // integer literals: 123, 0x1A, 0b1010, 0o777
	FLOAT     // floating point literals: 123.45, 1.23e-4
	STRING    // string literals: "hello", 'world', `template`
	TEMPLATE  // template string literals: `hello ${name}`
	BOOLEAN   // boolean literals: true, false
	NULL      // null literal
	UNDEFINED // undefined literal
	REGEX     // regular expression literals: /pattern/flags
	literal_end

	// Operators
	operator_beg
	// Arithmetic operators
	ADD // +
	SUB // -
	MUL // *
	DIV // /
	MOD // %
	POW // ** (exponentiation)

	// Bitwise operators
	BIT_AND     // &
	BIT_OR      // |
	BIT_XOR     // ^
	BIT_NOT     // ~
	BIT_LSHIFT  // <<
	BIT_RSHIFT  // >>
	BIT_URSHIFT // >>> (unsigned right shift)

	// Logical operators
	LOGICAL_AND // &&
	LOGICAL_OR  // ||
	LOGICAL_NOT // !

	// Comparison operators
	EQ        // ==
	NE        // !=
	STRICT_EQ // ===
	STRICT_NE // !==
	LT        // <
	LE        // <=
	GT        // >
	GE        // >=

	// Assignment operators
	ASSIGN          // =
	ADD_ASSIGN      // +=
	SUB_ASSIGN      // -=
	MUL_ASSIGN      // *=
	DIV_ASSIGN      // /=
	MOD_ASSIGN      // %=
	POW_ASSIGN      // **=
	BIT_AND_ASSIGN  // &=
	BIT_OR_ASSIGN   // |=
	BIT_XOR_ASSIGN  // ^=
	LSHIFT_ASSIGN   // <<=
	RSHIFT_ASSIGN   // >>=
	URSHIFT_ASSIGN  // >>>=

	// Increment/Decrement
	INCREMENT // ++
	DECREMENT // --

	// Type operators
	TYPEOF     // typeof
	INSTANCEOF // instanceof
	IN         // in
	AS         // as (type assertion)

	// Other operators
	QUESTION    // ? (ternary operator)
	NULLISH     // ?? (nullish coalescing)
	OPTIONAL    // ?. (optional chaining)
	SPREAD      // ... (spread operator)
	operator_end

	// Delimiters
	delimiter_beg
	LPAREN     // (
	RPAREN     // )
	LBRACE     // {
	RBRACE     // }
	LBRACKET   // [
	RBRACKET   // ]
	SEMICOLON  // ;
	COMMA      // ,
	DOT        // .
	COLON      // :
	ARROW      // => (arrow function)
	delimiter_end

	// Keywords
	keyword_beg
	// Control flow
	IF       // if
	ELSE     // else
	SWITCH   // switch
	CASE     // case
	DEFAULT  // default
	FOR      // for
	WHILE    // while
	DO       // do
	BREAK    // break
	CONTINUE // continue
	RETURN   // return
	THROW    // throw
	TRY      // try
	CATCH    // catch
	FINALLY  // finally

	// Function and class
	FUNCTION    // function
	CLASS       // class
	EXTENDS     // extends
	IMPLEMENTS  // implements
	CONSTRUCTOR // constructor
	SUPER       // super
	THIS        // this
	NEW         // new
	STATIC      // static
	ABSTRACT    // abstract
	ASYNC       // async
	AWAIT       // await

	// Variable declarations
	VAR   // var
	LET   // let
	CONST // const

	// Types (TypeScript specific)
	TYPE      // type
	INTERFACE // interface
	ENUM      // enum
	NAMESPACE // namespace
	MODULE    // module
	DECLARE   // declare

	// Access modifiers
	PUBLIC    // public
	PRIVATE   // private
	PROTECTED // protected
	READONLY  // readonly

	// Type keywords
	ANY      // any
	UNKNOWN  // unknown
	NEVER    // never
	VOID     // void
	OBJECT   // object
	STRING_T // string (type)
	NUMBER_T // number (type)
	BOOLEAN_T // boolean (type)
	SYMBOL   // symbol

	// TG-Script optimized types
	INT_T   // int (optimized integer type)
	FLOAT_T // float (optimized floating point type)

	// Import/Export
	IMPORT // import
	EXPORT // export
	FROM   // from

	// Other keywords
	DELETE // delete
	WITH   // with
	YIELD  // yield

	// Reserved for future use
	PACKAGE // package
	keyword_end
)

// Token names for debugging and error messages
var tokenNames = [...]string{
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
	COMMENT:   "COMMENT",
	IDENT:     "IDENT",
	INT:       "INT",
	FLOAT:     "FLOAT",
	STRING:    "STRING",
	TEMPLATE:  "TEMPLATE",
	BOOLEAN:   "BOOLEAN",
	NULL:      "NULL",
	UNDEFINED: "UNDEFINED",
	REGEX:     "REGEX",

	// Operators
	ADD:             "+",
	SUB:             "-",
	MUL:             "*",
	DIV:             "/",
	MOD:             "%",
	POW:             "**",
	BIT_AND:         "&",
	BIT_OR:          "|",
	BIT_XOR:         "^",
	BIT_NOT:         "~",
	BIT_LSHIFT:      "<<",
	BIT_RSHIFT:      ">>",
	BIT_URSHIFT:     ">>>",
	LOGICAL_AND:     "&&",
	LOGICAL_OR:      "||",
	LOGICAL_NOT:     "!",
	EQ:              "==",
	NE:              "!=",
	STRICT_EQ:       "===",
	STRICT_NE:       "!==",
	LT:              "<",
	LE:              "<=",
	GT:              ">",
	GE:              ">=",
	ASSIGN:          "=",
	ADD_ASSIGN:      "+=",
	SUB_ASSIGN:      "-=",
	MUL_ASSIGN:      "*=",
	DIV_ASSIGN:      "/=",
	MOD_ASSIGN:      "%=",
	POW_ASSIGN:      "**=",
	BIT_AND_ASSIGN:  "&=",
	BIT_OR_ASSIGN:   "|=",
	BIT_XOR_ASSIGN:  "^=",
	LSHIFT_ASSIGN:   "<<=",
	RSHIFT_ASSIGN:   ">>=",
	URSHIFT_ASSIGN:  ">>>=",
	INCREMENT:       "++",
	DECREMENT:       "--",
	TYPEOF:          "typeof",
	INSTANCEOF:      "instanceof",
	IN:              "in",
	AS:              "as",
	QUESTION:        "?",
	NULLISH:         "??",
	OPTIONAL:        "?.",
	SPREAD:          "...",

	// Delimiters
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACE:    "{",
	RBRACE:    "}",
	LBRACKET:  "[",
	RBRACKET:  "]",
	SEMICOLON: ";",
	COMMA:     ",",
	DOT:       ".",
	COLON:     ":",
	ARROW:     "=>",

	// Keywords
	IF:          "if",
	ELSE:        "else",
	SWITCH:      "switch",
	CASE:        "case",
	DEFAULT:     "default",
	FOR:         "for",
	WHILE:       "while",
	DO:          "do",
	BREAK:       "break",
	CONTINUE:    "continue",
	RETURN:      "return",
	THROW:       "throw",
	TRY:         "try",
	CATCH:       "catch",
	FINALLY:     "finally",
	FUNCTION:    "function",
	CLASS:       "class",
	EXTENDS:     "extends",
	IMPLEMENTS:  "implements",
	CONSTRUCTOR: "constructor",
	SUPER:       "super",
	THIS:        "this",
	NEW:         "new",
	STATIC:      "static",
	ABSTRACT:    "abstract",
	ASYNC:       "async",
	AWAIT:       "await",
	VAR:         "var",
	LET:         "let",
	CONST:       "const",
	TYPE:        "type",
	INTERFACE:   "interface",
	ENUM:        "enum",
	NAMESPACE:   "namespace",
	MODULE:      "module",
	DECLARE:     "declare",
	PUBLIC:      "public",
	PRIVATE:     "private",
	PROTECTED:   "protected",
	READONLY:    "readonly",
	ANY:         "any",
	UNKNOWN:     "unknown",
	NEVER:       "never",
	VOID:        "void",
	OBJECT:      "object",
	STRING_T:    "string",
	NUMBER_T:    "number",
	BOOLEAN_T:   "boolean",
	SYMBOL:      "symbol",
	INT_T:       "int",
	FLOAT_T:     "float",
	IMPORT:      "import",
	EXPORT:      "export",
	FROM:        "from",
	DELETE:      "delete",
	WITH:        "with",
	YIELD:       "yield",
	PACKAGE:     "package",
}

// Keywords map for quick lookup
var keywords map[string]Token

// String returns the string representation of the token
func (tok Token) String() string {
	if int(tok) < len(tokenNames) {
		return tokenNames[tok]
	}
	return "token(" + strconv.Itoa(int(tok)) + ")"
}

// IsLiteral returns true if the token is a literal
func (tok Token) IsLiteral() bool {
	return literal_beg < tok && tok < literal_end
}

// IsOperator returns true if the token is an operator
func (tok Token) IsOperator() bool {
	return operator_beg < tok && tok < operator_end
}

// IsKeyword returns true if the token is a keyword
func (tok Token) IsKeyword() bool {
	return keyword_beg < tok && tok < keyword_end
}

// IsDelimiter returns true if the token is a delimiter
func (tok Token) IsDelimiter() bool {
	return delimiter_beg < tok && tok < delimiter_end
}

// Lookup returns the token associated with the given identifier.
// If the identifier is a keyword, it returns the keyword token.
// Otherwise, it returns IDENT.
func Lookup(ident string) Token {
	if tok, exists := keywords[ident]; exists {
		return tok
	}
	return IDENT
}

// Precedence returns the operator precedence of the token.
// Higher numbers indicate higher precedence.
func (tok Token) Precedence() int {
	switch tok {
	case LOGICAL_OR:
		return 1
	case LOGICAL_AND:
		return 2
	case BIT_OR:
		return 3
	case BIT_XOR:
		return 4
	case BIT_AND:
		return 5
	case EQ, NE, STRICT_EQ, STRICT_NE:
		return 6
	case LT, LE, GT, GE, INSTANCEOF, IN:
		return 7
	case BIT_LSHIFT, BIT_RSHIFT, BIT_URSHIFT:
		return 8
	case ADD, SUB:
		return 9
	case MUL, DIV, MOD:
		return 10
	case POW:
		return 11
	default:
		return 0
	}
}

// IsAssignment returns true if the token is an assignment operator
func (tok Token) IsAssignment() bool {
	switch tok {
	case ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, DIV_ASSIGN, MOD_ASSIGN,
		POW_ASSIGN, BIT_AND_ASSIGN, BIT_OR_ASSIGN, BIT_XOR_ASSIGN,
		LSHIFT_ASSIGN, RSHIFT_ASSIGN, URSHIFT_ASSIGN:
		return true
	default:
		return false
	}
}

// IsUnaryOperator returns true if the token can be used as a unary operator
func (tok Token) IsUnaryOperator() bool {
	switch tok {
	case ADD, SUB, LOGICAL_NOT, BIT_NOT, INCREMENT, DECREMENT, TYPEOF, DELETE:
		return true
	default:
		return false
	}
}

// init initializes the keywords map
func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokenNames[i]] = i
	}
	
	// Add special literal keywords
	keywords["true"] = BOOLEAN
	keywords["false"] = BOOLEAN
	keywords["null"] = NULL
	keywords["undefined"] = UNDEFINED
	
	// Add operator keywords (these are operators but should be recognized as keywords in lookup)
	keywords["typeof"] = TYPEOF
	keywords["instanceof"] = INSTANCEOF
	keywords["in"] = IN
	keywords["as"] = AS
	keywords["delete"] = DELETE
}