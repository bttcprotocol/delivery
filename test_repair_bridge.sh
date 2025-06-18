#!/bin/bash

# 测试使用 bridge 广播方式的 repair-test 接口
echo "=== 测试 repair-test 接口（使用 bridge 广播方式）==="

# 设置变量
REST_URL="http://localhost:1317"
ACCOUNT_ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHECKPOINT_NUMBER="1"
TEST_MESSAGE="test_bridge_broadcast"

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
else
    echo ""
    echo "❌ 测试失败！"
    echo "错误信息: $(echo "$RESPONSE" | jq -r '.error // .')"
fi

echo ""
echo "=== 测试完成 ===" 