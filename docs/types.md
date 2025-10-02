# Types Module

TG-Script's static type system, designed based on the TypeScript type system, providing compile-time type checking and inference.

## Design Goals

- **TypeScript Compatibility**: Support TypeScript's core type features and syntax
- **Performance Optimization**: Optimized numeric types (int/float) with zero runtime type overhead
- **Strong Type Inference**: Maintain TypeScript's type inference capabilities
- **Generic Support**: Support TypeScript-style generic functions and interfaces
- **Precise Error Diagnosis**: Provide TypeScript-style type error messages

## Type Hierarchy

```
Type (Base type interface)
├── PrimitiveType (Primitive types - TypeScript compatible)
│   ├── NumberType (Numeric types)
│   │   ├── IntType (Integer types: int, int8, int16, int32, int64)
│   │   └── FloatType (Float types: float, float32, float64)
│   ├── StringType (String type)
│   ├── BooleanType (Boolean type)
│   ├── UndefinedType (undefined type)
│   ├── NullType (null type)
│   └── VoidType (void type)
├── ObjectType (Object types - TypeScript compatible)
│   ├── ArrayType (Array type: T[])
│   ├── TupleType (Tuple type: [T1, T2, ...])
│   ├── InterfaceType (Interface type)
│   ├── ClassType (Class type)
│   └── FunctionType (Function type)
├── GenericType (Generic types - TypeScript compatible)
│   ├── TypeParameter (Type parameters: T, U, K, V)
│   └── GenericInstance (Generic instances: Array<T>)
└── AdvancedType (Advanced types - TypeScript compatible)
    ├── UnionType (Union types: T | U)
    ├── IntersectionType (Intersection types: T & U)
    ├── LiteralType (Literal types)
    └── NeverType (never type)
```

## File Structure

- `types.go` - Type definitions and interfaces
- `checker.go` - Type checker implementation
- `inference.go` - Type inference algorithms
- `resolver.go` - Symbol resolution and scope management

## Type Checking Process

1. **Symbol Resolution**: Build symbol table, resolve identifiers
2. **Type Inference**: Infer types not explicitly declared
3. **Type Checking**: Verify type compatibility
4. **Error Reporting**: Generate detailed error messages