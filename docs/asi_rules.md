# 自动分号插入 (ASI) 规则

TG-Script 实现了类似 TypeScript/JavaScript 的自动分号插入 (Automatic Semicolon Insertion, ASI) 规则，使分号在大多数情况下成为可选的语句终止符。

## 基本规则

### 1. 分号可选的情况

在以下情况下，分号是可选的：

- **变量声明**
  ```tg
  let a = 10        // ✅ 无分号
  let b = 20;       // ✅ 有分号
  ```

- **表达式语句**
  ```tg
  print("Hello")    // ✅ 无分号
  console.log("Hi"); // ✅ 有分号
  ```

- **return 语句**
  ```tg
  function test() {
      return 42     // ✅ 无分号
      return 0;     // ✅ 有分号
  }
  ```

- **break 和 continue 语句**
  ```tg
  for (let i = 0; i < 10; i++) {
      if (i === 5) break     // ✅ 无分号
      if (i === 3) continue; // ✅ 有分号
  }
  ```

### 2. 分号必需的情况

在以下情况下，分号仍然是必需的：

- **for 循环中的分隔符**
  ```tg
  for (let i = 0; i < 10; i++) {  // ✅ 分号必需
      // 循环体
  }
  ```

- **同一行多个语句**
  ```tg
  let x = 1; let y = 2; print(x, y)  // ✅ 分号必需
  ```

## ASI 触发条件

自动分号插入在以下情况下触发：

1. **文件结束 (EOF)**
   ```tg
   let a = 10  // 文件结束，自动插入分号
   ```

2. **遇到右大括号 '}'**
   ```tg
   if (true) {
       let a = 10  // 遇到 '}' 前，自动插入分号
   }
   ```

3. **换行符后**
   ```tg
   let a = 10  // 换行后下一个token，自动插入分号
   let b = 20
   ```

## 限制性产生式

某些语句（如 `return`、`break`、`continue`）遵循限制性产生式规则：

```tg
function test() {
    return    // ✅ 自动插入分号，返回 undefined
    42        // 这行代码不会被执行
}

function test2() {
    return 42 // ✅ 返回 42
}
```

## 最佳实践

### 推荐做法

1. **省略行尾分号**
   ```tg
   let name = "TG-Script"
   let version = "1.0.0"
   print(name, version)
   ```

2. **保留 for 循环中的分号**
   ```tg
   for (let i = 0; i < 10; i++) {
       print(i)
   }
   ```

3. **同行多语句使用分号**
   ```tg
   let a = 1; let b = 2; let c = 3
   ```

### 避免的做法

1. **不要在限制性产生式后换行**
   ```tg
   // ❌ 避免
   return
       42
   
   // ✅ 推荐
   return 42
   ```

2. **不要依赖复杂的 ASI 规则**
   ```tg
   // ❌ 可能引起混淆
   let a = b
   (c + d).print()
   
   // ✅ 更清晰
   let a = b;
   (c + d).print()
   ```

## 与 TypeScript 的兼容性

TG-Script 的 ASI 规则与 TypeScript 保持高度兼容：

- 支持相同的 ASI 触发条件
- 遵循相同的限制性产生式规则
- 保持相同的语义行为

这确保了从 TypeScript 迁移到 TG-Script 的平滑过渡。

## 示例

### 完整示例

```tg
// 变量声明 - 无分号
let name = "TG-Script"
let version = 1.0
let isStable = true

// 函数定义
function greet(user) {
    if (!user) {
        return "Hello, World!"  // return 语句无分号
    }
    return `Hello, ${user}!`
}

// 循环 - for 中保留分号
for (let i = 0; i < 3; i++) {
    print(`Iteration ${i}`)
    
    if (i === 1) {
        continue  // continue 语句无分号
    }
    
    if (i === 2) {
        break     // break 语句无分号
    }
}

// 表达式语句
print(greet())
print(greet("Developer"))

// 同行多语句 - 使用分号
let x = 10; let y = 20; print("Sum:", x + y)
```

这个示例展示了 TG-Script 中分号使用的各种情况，既保持了代码的简洁性，又确保了语法的正确性。