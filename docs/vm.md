# VM Module

TG-Script's virtual machine runtime, using a register-based architecture designed for high-performance execution of TypeScript-compatible code.

## Design Goals

- **TypeScript Semantics**: Accurately implement TypeScript's runtime semantics
- **High-Performance Execution**: Register-based architecture optimized for common TypeScript patterns
- **Compact Bytecode**: Instruction encoding optimized for TypeScript syntax
- **Direct Threading**: Instruction dispatch optimization to improve TypeScript code execution efficiency
- **Debug Support**: TypeScript-compatible debugger interface with source mapping support

## File Structure

- `vm.go` - Core virtual machine implementation
- `instruction.go` - Instruction definitions and encoding
- `stack.go` - Call stack management
- `register.go` - Register allocation and management

## Instruction Set Design

### Basic Instructions
- `LOAD_CONST` - Load constant to register
- `LOAD_LOCAL` - Load local variable
- `STORE_LOCAL` - Store to local variable
- `MOVE` - Move between registers

### Arithmetic Instructions
- `ADD`, `SUB`, `MUL`, `DIV` - Basic arithmetic operations
- `MOD`, `POW` - Modulo and power operations

### Control Flow Instructions
- `JUMP` - Unconditional jump
- `JUMP_IF_TRUE`, `JUMP_IF_FALSE` - Conditional jumps
- `CALL` - Function call
- `RETURN` - Function return

## Performance Optimizations

- **Instruction Fusion**: Merge common instruction sequences
- **Branch Prediction**: Optimize conditional jump performance
- **Inline Caching**: Optimize property access
- **JIT Preparation**: Prepare for future JIT compilation