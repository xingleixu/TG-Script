# TG-Script Syntax Specification

TG-Script is designed based on a TypeScript subset, using the `.tg` file extension, maintaining TypeScript's familiar syntax while providing performance optimizations.

## 🎯 Design Principles

- **TypeScript Compatibility**: Maintains TypeScript's core syntax and semantics
- **Performance Optimization**: Optimized type system, removes performance overhead
- **Simplified Syntax**: Removes complex and error-prone features
- **Low Learning Cost**: TypeScript developers can migrate seamlessly

## 📝 Syntax Features

### ✅ Retained TypeScript Features

#### 1. Variable Declarations
```typescript
// Maintains TypeScript's variable declaration syntax
let name: string = "hello"
const PI: number = 3.14159
var count: number = 0  // Retained but not recommended

// Type inference
let message = "hello"  // Inferred as string
let count = 42         // Inferred as int
```

#### 2. Function Definitions
```typescript
// Function declaration
function add(a: number, b: number): number {
    return a + b
}

// Arrow function
const multiply = (a: number, b: number): number => a * b

// Optional parameters
function greet(name: string, title?: string): string {
    return title ? `${title} ${name}` : name
}

// Default parameters
function power(base: number, exponent: number = 2): number {
    return Math.pow(base, exponent)
}
```

#### 3. Interfaces and Types
```typescript
// Interface definition
interface Point {
    x: number
    y: number
}

// Type alias
type ID = string | number

// Union types
type Status = "pending" | "success" | "error"
```

#### 4. Class Definitions
```typescript
class Rectangle {
    width: number
    height: number
    
    constructor(width: number, height: number) {
        this.width = width
        this.height = height
    }
    
    area(): number {
        return this.width * this.height
    }
}
```

#### 5. Control Flow
```typescript
// if-else
if (condition) {
    // ...
} else if (otherCondition) {
    // ...
} else {
    // ...
}

// for loops
for (let i = 0; i < 10; i++) {
    console.log(i)
}

// for-of loops
for (const item of items) {
    console.log(item)
}

// while loops
while (condition) {
    // ...
}
```

#### 6. Arrays and Objects
```typescript
// Arrays
let numbers: number[] = [1, 2, 3, 4, 5]
let names: Array<string> = ["Alice", "Bob", "Charlie"]

// Objects
let person = {
    name: "John",
    age: 30,
    city: "New York"
}
```

### 🔧 Optimized Features

#### 1. Fine-grained Numeric Types
```typescript
// Replaces TypeScript's number type
let count: int = 42           // 64-bit integer
let price: float = 19.99      // 64-bit floating point
let byte: int8 = 127          // 8-bit integer
let large: int64 = 1000000    // Explicit 64-bit integer

// Maintains compatibility: number as alias for float
let legacy: number = 3.14     // Equivalent to float
```

#### 2. Optional Semicolons
```typescript
// Semicolons are optional (recommended to omit)
let name = "hello"
let count = 42
console.log(name)

// But still supported when needed
let a = 1; let b = 2; // Multiple statements on same line
```

#### 3. Strict Null Checking
```typescript
// Strict null checking enabled by default
let name: string = "hello"        // Cannot be null
let optional: string | null = null // Explicitly allows null
let maybe: string | undefined      // Explicitly allows undefined
```

### ❌ Removed TypeScript Features

#### 1. Complex Type Operations
```typescript
// ❌ Complex type operations not supported
// type Partial<T> = { [P in keyof T]?: T[P] }
// type Pick<T, K extends keyof T> = { [P in K]: T[P] }
```

#### 2. Decorators
```typescript
// ❌ Decorators not supported
// @Component
// class MyComponent { }
```

#### 3. Namespaces
```typescript
// ❌ Namespaces not supported
// namespace MyNamespace {
//     export function helper() { }
// }
```

#### 4. Complex Enum Features
```typescript
// ✅ Simple enums supported
enum Color {
    Red,
    Green,
    Blue
}

// ❌ Complex enums not supported
// enum Direction {
//     Up = "UP",
//     Down = "DOWN"
// }
```

#### 5. Dynamic Features
```typescript
// ❌ any type not supported
// let value: any = 42

// ❌ eval not supported
// eval("console.log('hello')")

// ❌ with statements not supported
// with (obj) { ... }
```

## 🔄 Type Mapping

### TypeScript → TG-Script
```typescript
// TypeScript          // TG-Script
number               → float (default) or int
string               → string
boolean              → bool
Array<T>             → T[]
object               → object
Function             → function
undefined            → undefined
null                 → null
void                 → void
```

## 📖 Complete Example

```typescript
// example.tg - Demonstrates TG-Script's main features
interface User {
    id: int
    name: string
    email: string
    age?: int
}

class UserService {
    private users: User[] = []
    
    addUser(user: User): void {
        this.users.push(user)
    }
    
    findUser(id: int): User | null {
        for (const user of this.users) {
            if (user.id === id) {
                return user
            }
        }
        return null
    }
    
    getUserCount(): int {
        return this.users.length
    }
}

// Usage example
const service = new UserService()

service.addUser({
    id: 1,
    name: "Alice",
    email: "alice@example.com",
    age: 25
})

const user = service.findUser(1)
if (user) {
    console.log(`Found user: ${user.name}`)
}

console.log(`Total users: ${service.getUserCount()}`)
```

## 🔄 Migration Guide

### Migrating from TypeScript to TG-Script

1. **File Extension**: Rename `.ts` files to `.tg`
2. **Numeric Type Optimization**:
   ```typescript
   // TypeScript
   let count: number = 42
   let price: number = 19.99
   
   // TG-Script
   let count: int = 42
   let price: float = 19.99
   ```
3. **Remove Unsupported Features**:
   - Remove `any` type usage
   - Remove decorators
   - Simplify complex type operations
4. **Optional Syntax Simplification**:
   - Remove unnecessary semicolons
   - Use strict null checking

### Command Line Tools

```bash
# Check syntax compatibility
tg check myfile.tg

# Auto-migrate TypeScript files
tg migrate myfile.ts

# Format code
tg fmt myfile.tg
```

### Compatibility
- Most TypeScript code can run directly
- Only minimal changes needed (mainly numeric types)
- Maintains the same development experience and tool support

This design maintains TypeScript's familiarity while achieving significant performance improvements!