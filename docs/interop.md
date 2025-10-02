# Interop Module

Go interoperability module providing seamless integration between TG-Script and Go code.

## Design Goals

- **Zero-Copy**: Direct memory mapping to avoid data copying
- **Type Safety**: Compile-time verification of type compatibility
- **High Performance**: Minimize cross-language call overhead
- **Ease of Use**: Automated binding generation

## File Structure

- `bridge.go` - Type bridging and conversion
- `binding.go` - Function binding and invocation
- `converter.go` - Automatic type conversion

## Supported Type Mappings

### Basic Types
- `int` ↔ `int64`
- `float` ↔ `float64`
- `string` ↔ `string`
- `bool` ↔ `bool`

### Composite Types
- `[]T` ↔ `[]T` (slices)
- `map[K]V` ↔ `map[K]V`
- `struct` ↔ `struct`

### Function Types
- `func(T) R` ↔ `func(T) R`

## Call Optimizations

- **Batch Calls**: Reduce cross-language call frequency
- **Async Calls**: Support for Go goroutines
- **Error Handling**: Unified error handling mechanism
- **Memory Management**: Automatic lifecycle management