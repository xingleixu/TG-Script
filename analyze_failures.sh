#!/bin/bash

echo "=== 分析失败测试用例的详细错误信息 ==="
echo

failed_tests=()
passed_tests=()

for file in tests/*.tg; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        echo "测试文件: $filename"
        
        # 运行测试并捕获输出
        output=$(./tg "$file" 2>&1)
        exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            echo "  ✅ 通过"
            passed_tests+=("$filename")
        else
            echo "  ❌ 失败"
            echo "  错误信息:"
            echo "$output" | sed 's/^/    /'
            failed_tests+=("$filename")
        fi
        echo "----------------------------------------"
    fi
done

echo
echo "=== 统计结果 ==="
echo "通过测试: ${#passed_tests[@]}"
echo "失败测试: ${#failed_tests[@]}"
echo "总测试数: $((${#passed_tests[@]} + ${#failed_tests[@]}))"
echo

echo "=== 失败测试列表 ==="
for test in "${failed_tests[@]}"; do
    echo "- $test"
done