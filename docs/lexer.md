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

## TypeScript Lexical Rules

### Keywords

TG-Script supports all TypeScript keywords plus TG-Script-specific optimizations:

#### TypeScript Standard Keywords
```typescript
// Control flow
if, else, switch, case, default, break, continue
for, while, do, return, throw, try, catch, finally

// Declarations
var, let, const, function, class, interface, enum, type
import, export, from, as, namespace, module

// Types
string, number, boolean, object, symbol, bigint, undefined, null, void, any, unknown, never

// Modifiers
public, private, protected, readonly, static, abstract, async, await
extends, implements, super, this, new, typeof, instanceof, in, of

// Advanced
yield, with, debugger, delete
```

#### TG-Script Optimized Keywords
```typescript
// Optimized numeric types
int, float, byte, uint, int32, int64, float32, float64

// Memory management
mut, ref, ptr
```

### Identifiers

Following TypeScript identifier rules:
- Must start with letter, underscore (_), or dollar sign ($)
- Can contain letters, digits, underscores, or dollar signs
- Support Unicode characters (following Unicode Standard Annex #31)
- Case-sensitive

```typescript
// Valid identifiers
myVariable, _private, $jquery, 变量名, αβγ, user123, __proto__

// Invalid identifiers
123abc, -name, @symbol, class (keyword)
```

### Literals

#### Numeric Literals
```typescript
// Decimal integers
42, 0, 123456789

// Hexadecimal (0x or 0X prefix)
0xFF, 0x1A2B, 0X123

// Octal (0o or 0O prefix)
0o777, 0O123

// Binary (0b or 0B prefix)
0b1010, 0B1111

// Floating-point
3.14, .5, 2., 1e10, 2.5e-3, 1E+5

// BigInt (n suffix)
123n, 0xFFn, 0b1010n

// Numeric separators (TypeScript 2.7+)
1_000_000, 0xFF_EC_DE_5E, 0b1010_0001
```

#### String Literals
```typescript
// Single quotes
'Hello, world!'
'It\'s a beautiful day'

// Double quotes
"Hello, world!"
"She said \"Hello\""

// Template literals (backticks)
`Hello, ${name}!`
`Line 1
Line 2`

// Escape sequences
'\n', '\t', '\r', '\\', '\'', '\"', '\u0041', '\x41'
```

#### Boolean and Special Literals
```typescript
// Boolean
true, false

// Null and undefined
null, undefined

// Regular expressions
/pattern/flags, /[a-z]+/gi
```

### Operators

#### Arithmetic Operators
```typescript
+, -, *, /, %, **  // Basic arithmetic
++, --             // Increment/decrement
+=, -=, *=, /=, %= // Assignment operators
```

#### Comparison Operators
```typescript
==, !=, ===, !==   // Equality
<, >, <=, >=       // Relational
```

#### Logical Operators
```typescript
&&, ||, !          // Logical
??, ?.             // Nullish coalescing, optional chaining
```

#### Bitwise Operators
```typescript
&, |, ^, ~         // Bitwise
<<, >>, >>>        // Shift operators
&=, |=, ^=, <<=, >>=, >>>= // Bitwise assignment
```

#### Type Operators
```typescript
typeof, instanceof, in, as, is, keyof, infer
```

#### Other Operators
```typescript
? :                // Ternary
=                  // Assignment
=>                 // Arrow function
...                // Spread/rest
```

### Delimiters and Punctuation

```typescript
// Brackets
( )                // Parentheses
[ ]                // Square brackets
{ }                // Curly braces

// Punctuation
;                  // Semicolon
,                  // Comma
.                  // Dot
:                  // Colon

// Type annotations
<T>                // Generic type parameters
```

### Comments

```typescript
// Single-line comment
/* Multi-line comment */
/** JSDoc comment */
```

### Whitespace and Line Terminators

- Space (U+0020)
- Tab (U+0009)
- Vertical Tab (U+000B)
- Form Feed (U+000C)
- Non-breaking Space (U+00A0)
- Line Feed (U+000A)
- Carriage Return (U+000D)
- Line Separator (U+2028)
- Paragraph Separator (U+2029)

## Automatic Semicolon Insertion (ASI)

TG-Script follows TypeScript's ASI rules:
- Semicolons are automatically inserted at line breaks when the next token cannot be parsed as part of the same statement
- No semicolon insertion within parentheses, brackets, or template literals
- Special handling for `return`, `throw`, `break`, `continue`, `yield`, and `debugger` statements

## Error Handling

The lexer provides TypeScript-compatible error messages:

```typescript
// Unterminated string
"Hello world    // Error: Unterminated string literal

// Invalid numeric literal
0x             // Error: Hexadecimal digit expected

// Invalid identifier
123abc         // Error: Identifier cannot start with digit

// Invalid escape sequence
"\q"           // Error: Invalid escape sequence
```

## Performance Optimizations

- **Token Pooling**: Reuse token objects to reduce GC pressure
- **Lookahead Buffering**: Efficient multi-character token recognition
- **Unicode Caching**: Cache Unicode property lookups
- **Streaming**: Support for large files with streaming lexical analysis