#!/bin/bash

# 快速测试脚本 - 验证chain_id修复

ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHAIN_ID="22125"
CHECKPOINT_NUMBER="60191"

echo "=== 快速测试 - 验证chain_id修复 ==="
echo ""

# 测试修复后的请求格式
REQUEST_BODY='{
  "base_req": {
    "from": "'$ADDRESS'",
    "chain_id": "'$CHAIN_ID'",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": []
  },
  "checkpoint_number": "'$CHECKPOINT_NUMBER'",
  "from": "'$ADDRESS'",
  "test_message": "快速测试"
}'

echo "发送请求到 /checkpoint/repair-test..."
echo "请求体:"
echo "$REQUEST_BODY" | jq '.' 2>/dev/null || echo "$REQUEST_BODY"
echo ""

RESPONSE=$(curl -X POST http://localhost:1317/checkpoint/repair-test \
  -H 'Content-Type: application/json' \
  -d "$REQUEST_BODY" \
  -s)

echo "响应:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
echo ""

if echo "$RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo "✅ 成功！测试消息广播成功"
elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo "❌ 仍有错误: $(echo $RESPONSE | jq -r '.error' 2>/dev/null)"
else
    echo "⚠️  响应格式异常: $RESPONSE"
fi 