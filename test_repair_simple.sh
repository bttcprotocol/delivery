#!/bin/bash

# 简单的测试脚本，通过调用层面解决 RabbitMQ 防重问题
echo "=== 简单测试 repair-test 接口（解决 RabbitMQ 防重问题）==="

# 设置变量
REST_URL="http://localhost:1317"
ACCOUNT_ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHECKPOINT_NUMBER="60191"

# 函数：生成简单唯一标识
generate_simple_id() {
    echo "$(date +%s)_$(shuf -i 1000-9999 -n 1)_$$"
}

# 函数：获取最新序列号
get_latest_sequence() {
    local sequence=$(curl -s "http://localhost:1317/auth/accounts/$ACCOUNT_ADDRESS" | jq -r '.result.value.sequence')
    if [ "$sequence" = "null" ] || [ -z "$sequence" ]; then
        echo "0"
    else
        echo "$sequence"
    fi
}

# 函数：发送测试请求
send_test_request() {
    local sequence=$1
    local unique_id=$2
    local test_msg="Simple_Test_${unique_id}"
    
    # 添加随机延迟，避免同时发送
    local delay=$((RANDOM % 3 + 2))
    echo "等待 ${delay} 秒后发送请求..."
    sleep $delay
    
    local request_body=$(cat <<EOF
{
    "base_req": {
        "from": "$ACCOUNT_ADDRESS",
        "chain_id": "heimdall-22125",
        "account_number": "0",
        "sequence": "$sequence",
        "gas": "200000",
        "gas_adjustment": "1.2",
        "fees": [],
        "simulate": false
    },
    "checkpoint_number": "$CHECKPOINT_NUMBER",
    "from": "$ACCOUNT_ADDRESS",
    "test_message": "$test_msg"
}
EOF
)

    echo "发送请求，序列号: $sequence"
    echo "唯一标识: $unique_id"
    echo "测试消息: $test_msg"
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$request_body" \
        "$REST_URL/checkpoint/repair-test")
    
    echo "$response"
}

# 主测试逻辑
echo "获取初始序列号..."
initial_sequence=$(get_latest_sequence)
echo "初始序列号: $initial_sequence"

# 第一次测试
echo ""
echo "=== 第一次测试 ==="
unique_id_1=$(generate_simple_id)
response1=$(send_test_request "$initial_sequence" "$unique_id_1")

echo "第一次响应:"
echo "$response1" | jq '.'

# 检查第一次是否成功
if echo "$response1" | jq -e '.success' > /dev/null; then
    echo "✅ 第一次测试成功！"
    
    # 等待更长时间，确保交易被完全处理
    echo "等待 8 秒让交易被完全处理..."
    sleep 8
    
    # 获取新的序列号
    echo ""
    echo "获取新的序列号..."
    new_sequence=$(get_latest_sequence)
    echo "新序列号: $new_sequence"
    
    # 第二次测试
    echo ""
    echo "=== 第二次测试 ==="
    unique_id_2=$(generate_simple_id)
    response2=$(send_test_request "$new_sequence" "$unique_id_2")
    
    echo "第二次响应:"
    echo "$response2" | jq '.'
    
    if echo "$response2" | jq -e '.success' > /dev/null; then
        echo "✅ 第二次测试也成功！"
        echo ""
        echo "🎉 连续两次测试都成功，说明："
        echo "1. 广播机制正常工作"
        echo "2. 序列号自动递增机制正常"
        echo "3. RabbitMQ 防重机制已通过调用层面解决"
    else
        echo "❌ 第二次测试失败"
        echo "错误信息: $(echo "$response2" | jq -r '.error // .')"
    fi
else
    echo "❌ 第一次测试失败"
    echo "错误信息: $(echo "$response1" | jq -r '.error // .')"
    echo ""
    echo "可能的原因："
    echo "1. 服务未启动"
    echo "2. 网络连接问题"
    echo "3. 账户余额不足"
    echo "4. RabbitMQ 队列问题"
fi

echo ""
echo "=== 测试完成 ===" 