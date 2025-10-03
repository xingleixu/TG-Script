#!/bin/bash

# 批量编译测试脚本
echo "开始批量编译测试..."

passed=0
failed=0
failed_files=()

for file in tests/*.tg; do
    if [ -f "$file" ]; then
        echo "编译测试: $file"
        if ./tg compile "$file" > /dev/null 2>&1; then
            echo "✓ 编译成功: $file"
            ((passed++))
        else
            echo "✗ 编译失败: $file"
            ((failed++))
            failed_files+=("$file")
        fi
    fi
done

echo ""
echo "编译测试总结:"
echo "通过: $passed"
echo "失败: $failed"

if [ $failed -gt 0 ]; then
    echo ""
    echo "失败的文件:"
    for file in "${failed_files[@]}"; do
        echo "  - $file"
    done
    exit 1
else
    echo "所有编译测试通过!"
    exit 0
fi