#!/bin/bash

# 核心测试脚本 - 测试01-06基础功能
echo "=== TG-Script 核心功能测试 ==="
echo ""

# 测试文件列表
test_files=(
    "01_variables.tg"
    "02_arithmetic.tg" 
    "03_comparison.tg"
    "04_logical.tg"
    "05_conditionals.tg"
    "06_loops.tg"
)

passed=0
failed=0

for file in "${test_files[@]}"; do
    echo "测试: $file"
    if ./bin/tg run "tests/$file" > /dev/null 2>&1; then
        echo "  ✓ 通过"
        passed=$((passed + 1))
    else
        echo "  ✗ 失败"
        failed=$((failed + 1))
        echo "  错误信息:"
        ./bin/tg run "tests/$file" 2>&1 | head -3 | sed 's/^/    /'
    fi
    echo ""
done

echo "=== 测试结果 ==="
echo "通过: $passed"
echo "失败: $failed"

if [ $failed -eq 0 ]; then
    echo "🎉 所有核心测试都通过了！"
    exit 0
else
    echo "⚠️  有测试失败"
    exit 1
fi