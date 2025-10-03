#!/bin/bash

# 简化版批量虚拟机执行测试脚本
echo "开始批量虚拟机执行测试..."

passed=0
failed=0

for file in tests/*.tg; do
    if [ -f "$file" ]; then
        echo -n "执行测试: $file ... "
        if ./tg run "$file" > /dev/null 2>&1; then
            echo "✓"
            ((passed++))
        else
            echo "✗"
            ((failed++))
        fi
    fi
done

echo ""
echo "虚拟机执行测试总结:"
echo "通过: $passed"
echo "失败: $failed"
echo "总计: $((passed + failed))"

if [ $failed -gt 0 ]; then
    exit 1
else
    echo "所有虚拟机执行测试通过!"
    exit 0
fi