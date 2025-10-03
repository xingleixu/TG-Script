#!/bin/bash

# 更全面的批量修复类型错误脚本
echo "开始批量修复类型错误..."

# 修复函数参数类型注解的更多模式
echo "修复函数参数类型注解..."

for file in tests/*.tg; do
    if [ -f "$file" ]; then
        echo "处理文件: $file"
        
        # 修复各种函数参数模式
        sed -i '' 's/function test(x, y)/function test(x: number, y: number): void/g' "$file"
        sed -i '' 's/function test(a, b)/function test(a: number, b: number): void/g' "$file"
        sed -i '' 's/function test(n)/function test(n: number): void/g' "$file"
        sed -i '' 's/function test(value)/function test(value: number): void/g' "$file"
        sed -i '' 's/function test(num)/function test(num: number): void/g' "$file"
        sed -i '' 's/function test(str)/function test(str: string): void/g' "$file"
        sed -i '' 's/function test(name, age)/function test(name: string, age: number): void/g' "$file"
        
        # 修复其他常见函数名
        sed -i '' 's/function func(a, b)/function func(a: number, b: number): number/g' "$file"
        sed -i '' 's/function calc(x, y)/function calc(x: number, y: number): number/g' "$file"
        sed -i '' 's/function compute(a, b)/function compute(a: number, b: number): number/g' "$file"
        sed -i '' 's/function process(data)/function process(data: number): number/g' "$file"
        sed -i '' 's/function validate(input)/function validate(input: number): boolean/g' "$file"
        sed -i '' 's/function check(value)/function check(value: number): boolean/g' "$file"
        
        # 修复带有单个参数的函数
        sed -i '' 's/function \([a-zA-Z_][a-zA-Z0-9_]*\)(\([a-zA-Z_][a-zA-Z0-9_]*\)) {/function \1(\2: number): void {/g' "$file"
        
        # 修复更多变量声明模式
        sed -i '' 's/^x = /let x = /g' "$file"
        sed -i '' 's/^y = /let y = /g' "$file"
        sed -i '' 's/^z = /let z = /g' "$file"
        sed -i '' 's/^n = /let n = /g' "$file"
        sed -i '' 's/^i = /let i = /g' "$file"
        sed -i '' 's/^j = /let j = /g' "$file"
        sed -i '' 's/^k = /let k = /g' "$file"
        sed -i '' 's/^data = /let data = /g' "$file"
        sed -i '' 's/^input = /let input = /g' "$file"
        sed -i '' 's/^output = /let output = /g' "$file"
        sed -i '' 's/^num = /let num = /g' "$file"
        sed -i '' 's/^str = /let str = /g' "$file"
        sed -i '' 's/^val = /let val = /g' "$file"
        sed -i '' 's/^res = /let res = /g' "$file"
        sed -i '' 's/^ret = /let ret = /g' "$file"
    fi
done

echo "类型错误修复完成！"