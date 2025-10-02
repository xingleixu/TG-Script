# TG-Script Coding Standards

This document outlines the coding standards and conventions for the TG-Script project to ensure consistency, maintainability, and clarity across the codebase.

## Language Requirements

### Documentation and Comments
- **All documentation must be written in English**
- **All code comments must be written in English**
- Use clear, concise language that can be understood by international developers
- Avoid colloquialisms and region-specific terminology

### Code Comments
```go
// Good: Clear English comment
// ParseExpression parses a TypeScript-compatible expression from the token stream
func ParseExpression(tokens []Token) (Expression, error) {
    // Implementation details...
}

// Bad: Non-English comment
// Parse expression from token stream
func ParseExpression(tokens []Token) (Expression, error) {
    // Implementation details...
}
```

## Project Structure

### Examples
- All example files must be placed in the `examples/` directory
- Example files should use the `.tg` extension for TG-Script code
- Examples should demonstrate real-world usage patterns
- Each example should include comprehensive English comments explaining the code

### Tests
- All test files must be placed in the `tests/` directory or alongside source files with `_test.go` suffix
- Test names should be descriptive and in English
- Test comments should explain the purpose and expected behavior

### Documentation
- All documentation files are located in the `docs/` directory
- Use Markdown format for documentation files
- Follow a consistent structure across documentation files

## Go Code Standards

### Naming Conventions
- Use Go's standard naming conventions (PascalCase for exported, camelCase for unexported)
- Choose descriptive names that clearly indicate purpose
- Avoid abbreviations unless they are widely understood

```go
// Good
type TypeChecker struct {
    symbolTable *SymbolTable
    errorReporter *ErrorReporter
}

// Bad
type TC struct {
    st *ST
    er *ER
}
```

### Function Documentation
- All exported functions and types must have English documentation comments
- Follow Go's documentation conventions (start with the function/type name)

```go
// ParseTypeScript parses TypeScript-compatible source code and returns an AST.
// It supports the TypeScript subset defined in the TG-Script specification.
func ParseTypeScript(source string) (*AST, error) {
    // Implementation...
}
```

### Error Handling
- Use descriptive error messages in English
- Provide context about what operation failed and why
- Use structured error types when appropriate

```go
// Good
return nil, fmt.Errorf("failed to parse expression at line %d: unexpected token %s", line, token.Value)

// Bad
return nil, fmt.Errorf("parse error")
```

## TG-Script Code Standards

### Syntax Examples
- Use TypeScript-compatible syntax in examples
- Demonstrate TG-Script's optimized features (int/float types)
- Include type annotations for clarity

```typescript
// Good: Clear TypeScript-compatible syntax with TG-Script optimizations
function calculateSum(numbers: int[]): int {
    let sum: int = 0
    for (let num of numbers) {
        sum += num
    }
    return sum
}

// Show both TypeScript compatibility and TG-Script optimizations
let count: int = 42        // TG-Script optimization: specific int type
let pi: float = 3.14159    // TG-Script optimization: specific float type
```

### Comments in Examples
- Explain TypeScript compatibility features
- Highlight TG-Script-specific optimizations
- Provide context for language design decisions

## File Organization

### Source Code Structure
```
TG-Script/
├── docs/                 # All documentation
│   ├── SYNTAX.md
│   ├── ARCHITECTURE.md
│   ├── CODING_STANDARDS.md
│   └── *.md
├── examples/             # Example TG-Script programs
│   ├── hello.tg
│   └── *.tg
├── tests/               # Test files and test data
│   ├── integration/
│   └── unit/
├── lexer/               # Lexical analysis
├── parser/              # Syntax analysis
├── types/               # Type system
├── compiler/            # Code generation
├── vm/                  # Virtual machine
├── runtime/             # Runtime system
├── interop/             # Go interoperability
└── stdlib/              # Standard library
```

### File Naming
- Use descriptive names that indicate the file's purpose
- Follow Go conventions for Go source files
- Use lowercase with underscores for multi-word names when necessary

## Version Control

### Commit Messages
- Write commit messages in English
- Use imperative mood ("Add feature" not "Added feature")
- Provide clear, descriptive commit messages

```
Good commit messages:
- "Add TypeScript interface parsing support"
- "Fix memory leak in VM register allocation"
- "Update documentation for type system"

Bad commit messages:
- "fix bug"
- "update"
- "fix"
```

### Branch Naming
- Use English for branch names
- Use descriptive names that indicate the purpose
- Follow pattern: `feature/description`, `bugfix/description`, `docs/description`

## Testing Standards

### Test Documentation
- All test functions must have English comments explaining what they test
- Include examples of expected input and output
- Document edge cases and error conditions

### Test Data
- Use English strings in test data when possible
- Provide clear variable names in test cases
- Include comments explaining complex test scenarios

## Performance Considerations

### Documentation
- Document performance characteristics in English
- Explain optimization decisions and trade-offs
- Provide benchmarking guidelines

### Code Comments
- Explain performance-critical sections
- Document algorithmic complexity where relevant
- Note memory allocation patterns

## Accessibility

### International Development
- Ensure all documentation can be understood by non-native English speakers
- Use simple, clear language structures
- Avoid idioms and cultural references
- Provide examples for complex concepts

### Code Readability
- Write self-documenting code with clear variable and function names
- Use consistent formatting and indentation
- Group related functionality logically

## Enforcement

These standards should be enforced through:
- Code review processes
- Automated linting where possible
- Documentation reviews
- Regular team discussions about code quality

## Updates

This document should be updated as the project evolves. All changes should be discussed with the team and documented in the project's change log.