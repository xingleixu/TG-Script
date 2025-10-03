#!/bin/bash

echo "=== 列出所有失败的测试文件 ==="
echo

failed_tests=()

for file in tests/*.tg; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        
        # 运行测试并检查退出码
        ./tg run "$file" >/dev/null 2>&1
        exit_code=$?
        
        if [ $exit_code -ne 0 ]; then
            failed_tests+=("$filename")
        fi
    fi
done

echo "失败的测试文件 (${#failed_tests[@]}个):"
for test in "${failed_tests[@]}"; do
    echo "- $test"
done

echo
echo "总计: ${#failed_tests[@]} 个失败测试"