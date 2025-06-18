#!/bin/bash

# 测试使用 bridge 广播方式的 repair-test 接口（使用 MsgRepairCheckpointTest）
echo "=== 测试 repair-test 接口（使用 MsgRepairCheckpointTest 消息类型）==="

# 设置变量
REST_URL="http://localhost:1317"
ACCOUNT_ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHECKPOINT_NUMBER="60191"
TEST_MESSAGE="快速测试_MsgRepairCheckpointTest_$(date +%s)"

# 获取账户序列号
echo "获取账户序列号..."
SEQUENCE=$(curl -s "http://localhost:1317/auth/accounts/$ACCOUNT_ADDRESS" | jq -r '.result.value.sequence')
if [ "$SEQUENCE" = "null" ] || [ -z "$SEQUENCE" ]; then
    echo "错误：无法获取账户序列号"
    exit 1
fi
echo "账户序列号: $SEQUENCE"

# 构建请求体
REQUEST_BODY=$(cat <<EOF
{
    "base_req": {
        "from": "$ACCOUNT_ADDRESS",
        "chain_id": "heimdall-22125",
        "account_number": "0",
        "sequence": "$SEQUENCE",
        "gas": "200000",
        "gas_adjustment": "1.2",
        "fees": [],
        "simulate": false
    },
    "checkpoint_number": "$CHECKPOINT_NUMBER",
    "from": "$ACCOUNT_ADDRESS",
    "test_message": "$TEST_MESSAGE"
}
EOF
)

echo "请求体:"
echo "$REQUEST_BODY" | jq '.'

# 发送请求
echo ""
echo "发送 repair-test 请求..."
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$REQUEST_BODY" \
    "$REST_URL/checkpoint/repair-test")

echo "响应:"
echo "$RESPONSE" | jq '.'

# 检查响应
if echo "$RESPONSE" | jq -e '.success' > /dev/null; then
    echo ""
    echo "✅ 测试成功！"
    echo "checkpoint_number: $(echo "$RESPONSE" | jq -r '.checkpoint_number')"
    echo "test_message: $(echo "$RESPONSE" | jq -r '.test_message')"
    echo ""
    echo "请检查服务日志，应该能看到以下日志："
    echo "1. repairCheckpointTestHandler, 开始广播测试消息"
    echo "2. repairCheckpointTestHandler, 广播成功"
    echo "3. handleMsgRepairCheckpointTest 中的日志"
else
    echo ""
    echo "❌ 测试失败！"
    echo "错误信息: $(echo "$RESPONSE" | jq -r '.error // .')"
    echo ""
    echo "可能的原因："
    echo "1. 序列号已被使用（防重机制）"
    echo "2. 服务未启动"
    echo "3. 网络连接问题"
fi

echo ""
echo "=== 测试完成 ===" 