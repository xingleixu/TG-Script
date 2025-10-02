package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Lexer represents the lexical analyzer
type Lexer struct {
	input        string   // the input source code
	position     int      // current position in input (points to current char)
	readPosition int      // current reading position in input (after current char)
	ch           byte     // current char under examination
	line         int      // current line number (1-based)
	column       int      // current column number (1-based)
	offset       int      // current byte offset (0-based)
	errors       []string // collection of lexer errors
}

// New creates a new lexer instance
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
		offset: 0,
		errors: make([]string, 0),
	}
	l.readChar() // initialize the lexer by reading the first character
	return l
}

// readChar reads the next character and advances position in the input
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL character represents "EOF"
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	// Update line and column tracking
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.offset = l.position
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekCharAt returns the character at the specified offset from current position
func (l *Lexer) peekCharAt(offset int) byte {
	pos := l.readPosition + offset - 1
	if pos >= len(l.input) || pos < 0 {
		return 0
	}
	return l.input[pos]
}

// currentPosition returns the current position information
func (l *Lexer) currentPosition() Position {
	return Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.offset,
	}
}

// addError adds an error message to the lexer's error collection
func (l *Lexer) addError(msg string) {
	pos := l.currentPosition()
	errorMsg := fmt.Sprintf("Lexer error at line %d, column %d: %s", pos.Line, pos.Column, msg)
	l.errors = append(l.errors, errorMsg)
}

// GetErrors returns all lexer errors
func (l *Lexer) GetErrors() []string {
	return l.errors
}

// HasErrors returns true if there are any lexer errors
func (l *Lexer) HasErrors() bool {
	return len(l.errors) > 0
}

// skipWhitespace skips whitespace characters (space, tab, newline, carriage return)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readSingleLineComment reads a single line comment starting with //
func (l *Lexer) readSingleLineComment() string {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readMultiLineComment reads a multi-line comment starting with /* and ending with */
func (l *Lexer) readMultiLineComment() string {
	position := l.position
	l.readChar() // consume '/'
	l.readChar() // consume '*'

	for {
		if l.ch == 0 {
			l.addError("unterminated multi-line comment")
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume '*'
			l.readChar() // consume '/'
			break
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

// isLetter checks if the character is a letter or underscore (valid for identifiers)
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$'
}

// isDigit checks if the character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// isHexDigit checks if the character is a hexadecimal digit
func isHexDigit(ch byte) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

// isBinaryDigit checks if the character is a binary digit
func isBinaryDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}

// isOctalDigit checks if the character is an octal digit
func isOctalDigit(ch byte) bool {
	return '0' <= ch && ch <= '7'
}

// isAlphaNumeric checks if the character is alphanumeric or underscore
func isAlphaNumeric(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

// readIdentifier reads an identifier (variable name, function name, etc.)
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isAlphaNumeric(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a numeric literal (integer or float)
func (l *Lexer) readNumber() (string, Token) {
	position := l.position
	tokenType := INT

	// Handle different number formats
	if l.ch == '0' {
		if l.peekChar() == 'x' || l.peekChar() == 'X' {
			// Hexadecimal number
			l.readChar() // consume '0'
			l.readChar() // consume 'x' or 'X'
			for isHexDigit(l.ch) {
				l.readChar()
			}
			return l.input[position:l.position], INT
		} else if l.peekChar() == 'b' || l.peekChar() == 'B' {
			// Binary number
			l.readChar() // consume '0'
			l.readChar() // consume 'b' or 'B'
			for isBinaryDigit(l.ch) {
				l.readChar()
			}
			return l.input[position:l.position], INT
		} else if l.peekChar() == 'o' || l.peekChar() == 'O' {
			// Octal number
			l.readChar() // consume '0'
			l.readChar() // consume 'o' or 'O'
			for isOctalDigit(l.ch) {
				l.readChar()
			}
			return l.input[position:l.position], INT
		}
	}

	// Regular decimal number
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = FLOAT
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for scientific notation
	if l.ch == 'e' || l.ch == 'E' {
		tokenType = FLOAT
		l.readChar() // consume 'e' or 'E'
		if l.ch == '+' || l.ch == '-' {
			l.readChar() // consume sign
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position], tokenType
}

// readString reads a string literal (single or double quoted)
func (l *Lexer) readString(delimiter byte) string {
	position := l.position + 1 // skip opening quote
	for {
		l.readChar()
		if l.ch == delimiter || l.ch == 0 {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar() // skip escape character
		}
	}
	return l.input[position:l.position]
}

// readTemplateString reads a template string literal (backtick quoted)
func (l *Lexer) readTemplateString() string {
	position := l.position + 1 // skip opening backtick
	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar() // skip escape character
		}
		// Note: Template expressions ${...} would need special handling
		// For now, we'll treat them as part of the template string
	}
	return l.input[position:l.position]
}

// NextToken scans the input and returns the next token
func (l *Lexer) NextToken() TokenInfo {
	var tok TokenInfo

	l.skipWhitespace()

	tok.Position = l.currentPosition()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = TokenInfo{Type: STRICT_EQ, Literal: "===", Position: tok.Position}
			} else {
				tok = TokenInfo{Type: EQ, Literal: string(ch) + string(l.ch), Position: tok.Position}
			}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = TokenInfo{Type: ARROW, Literal: "=>", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: ASSIGN, Literal: string(l.ch), Position: tok.Position}
		}
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			tok = TokenInfo{Type: INCREMENT, Literal: "++", Position: tok.Position}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: ADD_ASSIGN, Literal: "+=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: ADD, Literal: string(l.ch), Position: tok.Position}
		}
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = TokenInfo{Type: DECREMENT, Literal: "--", Position: tok.Position}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: SUB_ASSIGN, Literal: "-=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: SUB, Literal: string(l.ch), Position: tok.Position}
		}
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = TokenInfo{Type: POW_ASSIGN, Literal: "**=", Position: tok.Position}
			} else {
				tok = TokenInfo{Type: POW, Literal: "**", Position: tok.Position}
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: MUL_ASSIGN, Literal: "*=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: MUL, Literal: string(l.ch), Position: tok.Position}
		}
	case '/':
		if l.peekChar() == '/' {
			// Single line comment
			tok.Type = COMMENT
			tok.Literal = l.readSingleLineComment()
			return tok // early return to avoid readChar() call
		} else if l.peekChar() == '*' {
			// Multi-line comment
			tok.Type = COMMENT
			tok.Literal = l.readMultiLineComment()
			return tok // early return to avoid readChar() call
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: DIV_ASSIGN, Literal: "/=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: DIV, Literal: string(l.ch), Position: tok.Position}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = TokenInfo{Type: STRICT_NE, Literal: "!==", Position: tok.Position}
			} else {
				tok = TokenInfo{Type: NE, Literal: string(ch) + string(l.ch), Position: tok.Position}
			}
		} else {
			tok = TokenInfo{Type: LOGICAL_NOT, Literal: string(l.ch), Position: tok.Position}
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: LE, Literal: "<=", Position: tok.Position}
		} else if l.peekChar() == '<' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = TokenInfo{Type: LSHIFT_ASSIGN, Literal: "<<=", Position: tok.Position}
			} else {
				tok = TokenInfo{Type: BIT_LSHIFT, Literal: "<<", Position: tok.Position}
			}
		} else {
			tok = TokenInfo{Type: LT, Literal: string(l.ch), Position: tok.Position}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: GE, Literal: ">=", Position: tok.Position}
		} else if l.peekChar() == '>' {
			l.readChar()
			if l.peekChar() == '>' {
				l.readChar()
				if l.peekChar() == '=' {
					l.readChar()
					tok = TokenInfo{Type: URSHIFT_ASSIGN, Literal: ">>>=", Position: tok.Position}
				} else {
					tok = TokenInfo{Type: BIT_URSHIFT, Literal: ">>>", Position: tok.Position}
				}
			} else if l.peekChar() == '=' {
				l.readChar()
				tok = TokenInfo{Type: RSHIFT_ASSIGN, Literal: ">>=", Position: tok.Position}
			} else {
				tok = TokenInfo{Type: BIT_RSHIFT, Literal: ">>", Position: tok.Position}
			}
		} else {
			tok = TokenInfo{Type: GT, Literal: string(l.ch), Position: tok.Position}
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = TokenInfo{Type: LOGICAL_AND, Literal: "&&", Position: tok.Position}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: BIT_AND_ASSIGN, Literal: "&=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: BIT_AND, Literal: string(l.ch), Position: tok.Position}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = TokenInfo{Type: LOGICAL_OR, Literal: "||", Position: tok.Position}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: BIT_OR_ASSIGN, Literal: "|=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: BIT_OR, Literal: string(l.ch), Position: tok.Position}
		}
	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: BIT_XOR_ASSIGN, Literal: "^=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: BIT_XOR, Literal: string(l.ch), Position: tok.Position}
		}
	case '~':
		tok = TokenInfo{Type: BIT_NOT, Literal: string(l.ch), Position: tok.Position}
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok = TokenInfo{Type: MOD_ASSIGN, Literal: "%=", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: MOD, Literal: string(l.ch), Position: tok.Position}
		}
	case '?':
		if l.peekChar() == '?' {
			l.readChar()
			tok = TokenInfo{Type: NULLISH, Literal: "??", Position: tok.Position}
		} else if l.peekChar() == '.' {
			l.readChar()
			tok = TokenInfo{Type: OPTIONAL, Literal: "?.", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: QUESTION, Literal: string(l.ch), Position: tok.Position}
		}
	case '.':
		if l.peekChar() == '.' && l.peekCharAt(2) == '.' {
			l.readChar()
			l.readChar()
			tok = TokenInfo{Type: SPREAD, Literal: "...", Position: tok.Position}
		} else {
			tok = TokenInfo{Type: DOT, Literal: string(l.ch), Position: tok.Position}
		}
	case ',':
		tok = TokenInfo{Type: COMMA, Literal: string(l.ch), Position: tok.Position}
	case ';':
		tok = TokenInfo{Type: SEMICOLON, Literal: string(l.ch), Position: tok.Position}
	case ':':
		tok = TokenInfo{Type: COLON, Literal: string(l.ch), Position: tok.Position}
	case '(':
		tok = TokenInfo{Type: LPAREN, Literal: string(l.ch), Position: tok.Position}
	case ')':
		tok = TokenInfo{Type: RPAREN, Literal: string(l.ch), Position: tok.Position}
	case '{':
		tok = TokenInfo{Type: LBRACE, Literal: string(l.ch), Position: tok.Position}
	case '}':
		tok = TokenInfo{Type: RBRACE, Literal: string(l.ch), Position: tok.Position}
	case '[':
		tok = TokenInfo{Type: LBRACKET, Literal: string(l.ch), Position: tok.Position}
	case ']':
		tok = TokenInfo{Type: RBRACKET, Literal: string(l.ch), Position: tok.Position}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString('"')
	case '\'':
		tok.Type = STRING
		tok.Literal = l.readString('\'')
	case '`':
		tok.Type = TEMPLATE
		tok.Literal = l.readTemplateString()
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = Lookup(tok.Literal)
			return tok // early return to avoid readChar() call
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			return tok // early return to avoid readChar() call
		} else {
			// Handle Unicode characters
			if l.ch > 127 {
				r, _ := utf8.DecodeRuneInString(l.input[l.position:])
				if unicode.IsLetter(r) || r == '_' {
					// Unicode identifier
					tok.Literal = l.readIdentifier()
					tok.Type = Lookup(tok.Literal)
					return tok
				}
			}
			l.addError(fmt.Sprintf("unexpected character: %q", l.ch))
			tok = TokenInfo{Type: ILLEGAL, Literal: string(l.ch), Position: tok.Position}
		}
	}

	l.readChar()
	return tok
}

// TokenizeAll returns all tokens from the input
func (l *Lexer) TokenizeAll() []TokenInfo {
	var tokens []TokenInfo
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}
