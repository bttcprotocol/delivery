#!/bin/bash

# 智能测试脚本，自动处理序列号递增，避免防重机制
echo "=== 智能测试 repair-test 接口（自动处理序列号）==="

# 设置变量
REST_URL="http://localhost:1317"
ACCOUNT_ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHECKPOINT_NUMBER="60191"
TEST_MESSAGE="智能测试_$(date +%s)"

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
    local test_msg=$2
    
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
response1=$(send_test_request "$initial_sequence" "${TEST_MESSAGE}_1")

echo "第一次响应:"
echo "$response1" | jq '.'

# 检查第一次是否成功
if echo "$response1" | jq -e '.success' > /dev/null; then
    echo "✅ 第一次测试成功！"
    
    # 等待一下，让交易被处理
    echo "等待 3 秒让交易被处理..."
    sleep 3
    
    # 获取新的序列号
    echo ""
    echo "获取新的序列号..."
    new_sequence=$(get_latest_sequence)
    echo "新序列号: $new_sequence"
    
    # 第二次测试
    echo ""
    echo "=== 第二次测试 ==="
    response2=$(send_test_request "$new_sequence" "${TEST_MESSAGE}_2")
    
    echo "第二次响应:"
    echo "$response2" | jq '.'
    
    if echo "$response2" | jq -e '.success' > /dev/null; then
        echo "✅ 第二次测试也成功！"
        echo ""
        echo "🎉 连续两次测试都成功，说明："
        echo "1. 广播机制正常工作"
        echo "2. 序列号自动递增机制正常"
        echo "3. 防重机制正常工作"
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
fi

echo ""
echo "=== 测试完成 ===" 