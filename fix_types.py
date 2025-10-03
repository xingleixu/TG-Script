#!/usr/bin/env python3

import os
import re
import glob

def fix_function_signatures(content):
    """修复函数签名中缺少的类型注解"""
    
    # 匹配没有类型注解的函数定义
    patterns = [
        # function name(param1, param2) { -> function name(param1: type, param2: type): returnType {
        (r'function\s+(\w+)\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)\s*{', 
         r'function \1(\2: number, \3: number): void {'),
        
        # function name(param) { -> function name(param: type): returnType {
        (r'function\s+(\w+)\s*\(\s*(\w+)\s*\)\s*{', 
         r'function \1(\2: number): void {'),
        
        # 特殊情况：add, multiply等数学函数应该返回number
        (r'function\s+(add|multiply|subtract|divide|max|min|calc|compute)\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)\s*:\s*void\s*{', 
         r'function \1(\2: number, \3: number): number {'),
        
        (r'function\s+(factorial|fibonacci|square|abs)\s*\(\s*(\w+)\s*\)\s*:\s*void\s*{', 
         r'function \1(\2: number): number {'),
    ]
    
    for pattern, replacement in patterns:
        content = re.sub(pattern, replacement, content)
    
    return content

def fix_variable_declarations(content):
    """修复缺少let关键字的变量声明"""
    
    # 匹配行首的赋值语句（不是已经有let/const/var的）
    lines = content.split('\n')
    fixed_lines = []
    
    for line in lines:
        # 如果行首是标识符 = 表达式，且不是已经声明的变量
        if re.match(r'^[a-zA-Z_]\w*\s*=\s*', line.strip()) and not re.match(r'^\s*(let|const|var)\s+', line.strip()):
            # 添加let关键字
            fixed_line = re.sub(r'^(\s*)([a-zA-Z_]\w*\s*=)', r'\1let \2', line)
            fixed_lines.append(fixed_line)
        else:
            fixed_lines.append(line)
    
    return '\n'.join(fixed_lines)

def fix_file(filepath):
    """修复单个文件"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        # 应用修复
        content = fix_function_signatures(content)
        content = fix_variable_declarations(content)
        
        # 如果内容有变化，写回文件
        if content != original_content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"✓ 修复了 {filepath}")
            return True
        else:
            print(f"- 跳过 {filepath} (无需修复)")
            return False
            
    except Exception as e:
        print(f"✗ 修复 {filepath} 时出错: {e}")
        return False

def main():
    print("开始批量修复TG-Script类型错误...")
    
    # 获取所有.tg文件
    test_files = glob.glob('tests/*.tg')
    
    fixed_count = 0
    total_count = len(test_files)
    
    for filepath in sorted(test_files):
        if fix_file(filepath):
            fixed_count += 1
    
    print(f"\n修复完成！")
    print(f"总文件数: {total_count}")
    print(f"修复文件数: {fixed_count}")
    print(f"跳过文件数: {total_count - fixed_count}")

if __name__ == "__main__":
    main()