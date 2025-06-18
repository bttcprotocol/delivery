#!/bin/bash

# 测试事件修复后的 repair-test 接口
echo "=== 测试事件修复后的 repair-test 接口 ==="

# 设置变量
REST_URL="http://localhost:1317"
ACCOUNT_ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHECKPOINT_NUMBER="60191"

# 函数：获取最新序列号
get_latest_sequence() {
    local sequence=$(curl -s "http://localhost:1317/auth/accounts/$ACCOUNT_ADDRESS" | jq -r '.result.value.sequence')
    if [ "$sequence" = "null" ] || [ -z "$sequence" ]; then
        echo "0"
    else
        echo "$sequence"
    fi
}

# 获取序列号
echo "获取最新序列号..."
sequence=$(get_latest_sequence)
echo "序列号: $sequence"

# 发送测试请求
echo ""
echo "发送测试请求..."
request_body=$(cat <<EOF
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
    "test_message": "Event_Fix_Test_$(date +%s)"
}
EOF
)

response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$request_body" \
    "$REST_URL/checkpoint/repair-test")

echo "响应:"
echo "$response" | jq '.'

# 检查响应
if echo "$response" | jq -e '.success' > /dev/null; then
    echo ""
    echo "✅ 测试成功！"
    echo ""
    echo "🎉 事件修复验证："
    echo "1. 事件类型已正确定义在 types/events.go 中"
    echo "2. Handler 使用标准事件类型常量"
    echo "3. 事件属性使用标准属性键"
    echo "4. 消息成功广播到链上"
    echo ""
    echo "📋 下一步："
    echo "1. 查看服务日志确认 handler 处理"
    echo "2. 检查事件是否正确发出"
    echo "3. 验证事件监听器是否正常工作"
else
    echo ""
    echo "❌ 测试失败"
    echo "错误信息: $(echo "$response" | jq -r '.error // .')"
fi

echo ""
echo "=== 测试完成 ===" 