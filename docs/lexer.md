# Lexer Module

TG-Script's lexical analyzer, responsible for converting TypeScript-compatible source code into a token stream.

## Design Goals

- **TypeScript Compatibility**: Support TypeScript lexical rules and syntactic sugar
- **High Performance**: Zero-allocation token generation with memory pool reuse
- **Error Recovery**: Intelligent error recovery mechanism providing TypeScript-style error messages
- **Unicode Support**: Complete Unicode identifier support (TypeScript compatible)
- **Extensible**: Support for TG-Script-specific optimized tokens (such as int/float types)

## File Structure

- `token.go` - Token type definitions and constants
- `lexer.go` - Core lexical analyzer implementation
- `lexer_test.go` - Unit tests

## Token Types

Supported token types include:
- Keywords: `let`, `mut`, `const`, `func`, `struct`, `if`, `else`, etc.
- Identifiers: variable names, function names, etc.
- Literals: numbers, strings, boolean values
- Operators: `+`, `-`, `*`, `/`, `==`, `!=`, etc.
- Delimiters: `(`, `)`, `{`, `}`, `;`, `,`, etc.