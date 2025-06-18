#!/bin/bash

# 清空 RabbitMQ 队列脚本
echo "=== 清空 RabbitMQ 队列 ==="

# RabbitMQ 配置
RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
QUEUE_NAME="machinery_tasks"

# 检查 RabbitMQ 是否运行
echo "检查 RabbitMQ 连接..."
if ! curl -s -u guest:guest http://localhost:15672/api/overview > /dev/null 2>&1; then
    echo "❌ RabbitMQ 未运行或无法连接"
    echo "请确保 RabbitMQ 服务已启动："
    echo "  brew services start rabbitmq  # macOS"
    echo "  sudo systemctl start rabbitmq-server  # Linux"
    exit 1
fi

echo "✅ RabbitMQ 连接正常"

# 方法1：使用 RabbitMQ Management API 清空队列
echo ""
echo "方法1：使用 RabbitMQ Management API 清空队列"
echo "清空队列: $QUEUE_NAME"

RESPONSE=$(curl -s -u guest:guest \
    -H "Content-Type: application/json" \
    -X DELETE \
    "http://localhost:15672/api/queues/%2F/$QUEUE_NAME/contents")

if echo "$RESPONSE" | jq -e '.messages_purged' > /dev/null 2>&1; then
    PURGED_COUNT=$(echo "$RESPONSE" | jq -r '.messages_purged')
    echo "✅ 成功清空队列，删除了 $PURGED_COUNT 条消息"
else
    echo "❌ 清空队列失败"
    echo "响应: $RESPONSE"
fi

# 方法2：使用 AMQP 协议清空队列
echo ""
echo "方法2：使用 AMQP 协议清空队列"
echo "注意：这需要安装 amqp-tools"

if command -v amqp-delete-queue > /dev/null 2>&1; then
    echo "删除队列: $QUEUE_NAME"
    amqp-delete-queue --url="$RABBITMQ_URL" --queue="$QUEUE_NAME"
    echo "✅ 队列已删除"
else
    echo "⚠️  amqp-tools 未安装，跳过 AMQP 方法"
    echo "安装方法："
    echo "  brew install amqp-tools  # macOS"
    echo "  sudo apt-get install amqp-tools  # Ubuntu"
fi

# 检查队列状态
echo ""
echo "检查队列状态..."
QUEUE_INFO=$(curl -s -u guest:guest \
    "http://localhost:15672/api/queues/%2F/$QUEUE_NAME")

if echo "$QUEUE_INFO" | jq -e '.messages' > /dev/null 2>&1; then
    MESSAGE_COUNT=$(echo "$QUEUE_INFO" | jq -r '.messages')
    echo "队列 $QUEUE_NAME 当前消息数: $MESSAGE_COUNT"
else
    echo "队列 $QUEUE_NAME 不存在或无法访问"
fi

echo ""
echo "=== 清空完成 ===" 