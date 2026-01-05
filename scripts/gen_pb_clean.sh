#!/bin/bash

# jh_app_service 项目专用脚本
# 生成 protobuf 文件并移除 omitempty 标签

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info "=== jh_app_service - 生成并清理 protobuf 文件 ==="

# 获取脚本所在目录的父目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

print_info "项目根目录: $PROJECT_ROOT"

# 进入项目根目录
cd "$PROJECT_ROOT"

# 检查是否是有效的 Go 项目
if [ ! -f "go.mod" ]; then
    print_error "当前目录不是有效的 Go 模块: $(pwd)"
    exit 1
fi

# 检查项目名称
if ! grep -q "jh_app_service" go.mod; then
    print_warning "当前项目可能不是 jh_app_service"
fi

# 检查是否安装了 gf 命令
if ! command -v gf &> /dev/null; then
    print_error "gf 命令未找到，请先安装 GoFrame CLI"
    print_info "安装命令: go install github.com/gogf/gf/cmd/gf/v2@latest"
    exit 1
fi

print_info "=== 第一步: 生成 protobuf 文件 ==="

# 运行 gf gen pb
if gf gen pb; then
    print_success "protobuf 文件生成完成"
else
    print_error "protobuf 文件生成失败"
    exit 1
fi

print_info "=== 第二步: 移除 omitempty 标签 ==="

# 查找所有 .pb.go 文件
pb_files=$(find . -name "*.pb.go" -type f 2>/dev/null)

if [ -z "$pb_files" ]; then
    print_warning "未找到 .pb.go 文件"
    exit 0
fi

# 统计文件数量
file_count=$(echo "$pb_files" | wc -l | tr -d ' ')
print_info "找到 $file_count 个 .pb.go 文件"

# 处理每个文件
processed=0
modified=0
failed=0

while IFS= read -r file; do
    print_info "处理文件: $file"
    
    # 检查文件是否存在
    if [ ! -f "$file" ]; then
        print_error "  文件不存在: $file"
        failed=$((failed + 1))
        continue
    fi
    
    # 备份原文件
    if ! cp "$file" "${file}.bak"; then
        print_error "  备份文件失败: $file"
        failed=$((failed + 1))
        continue
    fi
    
    # 移除 omitempty 标签
    if sed 's/,omitempty//g' "${file}.bak" > "$file"; then
        # 检查是否有变化
        if ! cmp -s "$file" "${file}.bak"; then
            print_success "  ✓ 已移除 omitempty 标签"
            modified=$((modified + 1))
        else
            print_info "  - 无需修改"
        fi
        
        # 删除备份文件
        rm "${file}.bak"
        processed=$((processed + 1))
    else
        print_error "  处理文件失败: $file"
        # 恢复备份文件
        mv "${file}.bak" "$file"
        failed=$((failed + 1))
    fi
done <<< "$pb_files"

echo ""
print_info "=== 处理完成 ==="
print_success "总文件数: $file_count"
print_success "成功处理: $processed"
print_success "已修改: $modified"

if [ $failed -gt 0 ]; then
    print_error "失败: $failed"
    exit 1
else
    print_success "jh_app_service protobuf 文件生成和清理完成！"
fi

# 显示使用建议
echo ""
print_info "=== 使用建议 ==="
print_info "1. 检查修改结果: git diff"
print_info "2. 编译测试: go build"
print_info "3. 提交更改: git add . && git commit -m 'update: regenerate protobuf files'"