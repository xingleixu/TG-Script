# 扩展数值类型支持

TG-Script 现在支持更精确的数值类型，提供了比基本的 `int` 和 `float` 更细粒度的类型控制。

## 支持的扩展数值类型

### 整数类型

| 类型 | 位数 | 范围 | 用法示例 |
|------|------|------|----------|
| `int8` | 8位 | -128 到 127 | `let small: int8 = 100;` |
| `int16` | 16位 | -32,768 到 32,767 | `let medium: int16 = 30000;` |
| `int32` | 32位 | -2,147,483,648 到 2,147,483,647 | `let large: int32 = 2000000000;` |
| `int64` | 64位 | -9,223,372,036,854,775,808 到 9,223,372,036,854,775,807 | `let huge: int64 = 9000000000000000000;` |

### 浮点类型

| 类型 | 精度 | 用法示例 |
|------|------|----------|
| `float32` | 单精度 (32位) | `let pi: float32 = 3.14159;` |
| `float64` | 双精度 (64位) | `let e: float64 = 2.718281828459045;` |

## 基本用法

### 变量声明

```typescript
// 整数类型
let byte: int8 = 127;
let short: int16 = 32767;
let int: int32 = 2147483647;
let long: int64 = 9223372036854775807;

// 浮点类型
let singleFloat: float32 = 3.14159;
let doubleFloat: float64 = 2.718281828459045;
```

### 函数参数和返回值

```typescript
function addInt8(a: int8, b: int8): int8 {
    return a + b;
}

function processFloat32(value: float32): float32 {
    return value * 2.0;
}
```

### 算术运算

```typescript
let int8Sum: int8 = 10 + 20;
let int16Product: int16 = 100 * 200;
let float32Result: float32 = 1.5 + 2.5;
let float64Result: float64 = 3.14159 * 2.0;
```

### 比较运算

```typescript
let int8Compare: boolean = byte > 100;
let float32Compare: boolean = singleFloat >= 3.0;
```

## 实现细节

### 词法分析器 (Lexer)
- 添加了新的 token 类型：`INT8_T`, `INT16_T`, `INT32_T`, `INT64_T`, `FLOAT32_T`, `FLOAT64_T`
- 更新了关键字映射以识别新的类型关键字

### 语法分析器 (Parser)
- 更新了 `parseTypeReference` 和 `parsePrimaryType` 函数以支持新的类型 token
- 扩展了类型注解解析功能

### 类型检查器 (Type Checker)
- 添加了新的类型对象：`Int8Type`, `Int16Type`, `Int32Type`, `Int64Type`, `Float32Type`, `Float64Type`
- 更新了 `resolveTypeAnnotation` 方法以正确映射新的类型

## 测试覆盖

项目包含了全面的测试用例：

1. **基本类型测试** (`test_simple_extended_types.tg`) - 验证基本的变量声明和操作
2. **类型兼容性测试** (`test_type_compatibility.tg`) - 验证类型兼容性规则
3. **函数调试测试** (`test_function_debug.tg`) - 验证函数参数类型处理
4. **最终验证测试** (`test_final_validation.tg`) - 综合测试所有功能

## 注意事项

1. **类型安全**: 扩展数值类型提供了更严格的类型检查，有助于在编译时发现类型相关的错误。

2. **性能考虑**: 不同的数值类型在内存使用和计算性能上可能有差异，选择合适的类型可以优化程序性能。

3. **兼容性**: 新的扩展类型与现有的 `int` 和 `float` 类型保持兼容。

4. **边界值**: 使用时需要注意各类型的数值范围，避免溢出。

## 未来扩展

可能的未来改进包括：
- 无符号整数类型支持 (`uint8`, `uint16`, `uint32`, `uint64`)
- 类型转换函数
- 更复杂的数值运算支持
- 数组和集合的类型化支持