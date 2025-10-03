#!/bin/bash

# 批量测试脚本
# 测试所有tests目录中的.tg文件

echo "=== TG-Script 批量测试 ==="
echo "测试目录: tests/"
echo ""

# 计数器
total_files=0
check_passed=0
check_failed=0
run_passed=0
run_failed=0

# 获取所有.tg文件
for file in tests/*.tg; do
    if [ -f "$file" ]; then
        total_files=$((total_files + 1))
        filename=$(basename "$file")
        
        echo "[$total_files] 测试文件: $filename"
        
        # 测试类型检查
        echo -n "  类型检查: "
        if ./tg check "$file" > /dev/null 2>&1; then
            echo "✓ 通过"
            check_passed=$((check_passed + 1))
            
            # 如果类型检查通过，测试运行
            echo -n "  虚拟机执行: "
            if ./tg run "$file" > /dev/null 2>&1; then
                echo "✓ 通过"
                run_passed=$((run_passed + 1))
            else
                echo "✗ 失败"
                run_failed=$((run_failed + 1))
                echo "    运行错误，查看详细信息:"
                ./tg run "$file" 2>&1 | head -5 | sed 's/^/    /'
            fi
        else
            echo "✗ 失败"
            check_failed=$((check_failed + 1))
            echo "    类型检查错误，查看详细信息:"
            ./tg check "$file" 2>&1 | head -5 | sed 's/^/    /'
        fi
        
        echo ""
    fi
done

echo "=== 测试总结 ==="
echo "总文件数: $total_files"
echo "类型检查: $check_passed 通过, $check_failed 失败"
echo "虚拟机执行: $run_passed 通过, $run_failed 失败"
echo ""

if [ $check_failed -eq 0 ] && [ $run_failed -eq 0 ]; then
    echo "🎉 所有测试都通过了！"
    exit 0
else
    echo "⚠️  有测试失败，需要修复"
    exit 1
fi