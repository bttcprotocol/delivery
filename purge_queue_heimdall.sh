#!/bin/bash

# 使用 Heimdall 内置命令清空队列
echo "=== 使用 Heimdall 内置命令清空队列 ==="

# 检查 deliveryd 是否在运行
echo "检查 deliveryd 进程..."
if ! pgrep -f "deliveryd" > /dev/null; then
    echo "❌ deliveryd 未运行"
    echo "请先启动 deliveryd 服务"
    exit 1
fi

echo "✅ deliveryd 正在运行"

# 使用 Heimdall 的 purge-queue 命令
echo ""
echo "执行 Heimdall purge-queue 命令..."
cd bridge

# 检查是否存在 purge-queue 命令
if ! ./build/bridge purge-queue --help > /dev/null 2>&1; then
    echo "❌ purge-queue 命令不可用"
    echo "请确保 bridge 已正确编译"
    exit 1
fi

# 执行清空命令
echo "清空 machinery_tasks 队列..."
./build/bridge purge-queue

if [ $? -eq 0 ]; then
    echo "✅ 队列清空成功"
else
    echo "❌ 队列清空失败"
    exit 1
fi

cd ..

echo ""
echo "=== 清空完成 ===" 