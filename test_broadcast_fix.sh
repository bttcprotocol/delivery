#!/bin/bash

# BTTC Checkpoint 补录测试脚本 - 修复版本
# 解决签名验证失败问题

ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHAIN_ID="22125"
CHECKPOINT_NUMBER="60191"

echo "=== BTTC Checkpoint 补录测试 - 修复版本 ==="
echo "地址: $ADDRESS"
echo "链ID: $CHAIN_ID"
echo "Checkpoint编号: $CHECKPOINT_NUMBER"
echo ""

# 第一步：获取账户当前序列号
echo "1. 获取账户当前序列号..."
ACCOUNT_INFO=$(curl -s "http://localhost:1317/auth/accounts/$ADDRESS/sequence")
echo "账户信息: $ACCOUNT_INFO"

# 解析序列号
SEQUENCE=$(echo $ACCOUNT_INFO | jq -r '.sequence // 0')
ACCOUNT_NUMBER=$(echo $ACCOUNT_INFO | jq -r '.account_number // 0')

echo "当前序列号: $SEQUENCE"
echo "账户编号: $ACCOUNT_NUMBER"
echo ""

# 第二步：使用正确的序列号发送测试请求
echo "2. 发送修复后的测试请求..."

# 构建请求体，包含正确的序列号
REQUEST_BODY=$(cat <<EOF
{
  "base_req": {
    "from": "$ADDRESS",
    "chain_id": "$CHAIN_ID",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": [],
    "account_number": "$ACCOUNT_NUMBER",
    "sequence": "$SEQUENCE"
  },
  "checkpoint_number": "$CHECKPOINT_NUMBER",
  "from": "$ADDRESS",
  "test_message": "测试修复后的广播机制 - 使用正确序列号"
}
EOF
)

echo "请求体:"
echo "$REQUEST_BODY" | jq '.'
echo ""

# 发送请求
echo "3. 发送HTTP请求..."
RESPONSE=$(curl -X POST http://localhost:1317/checkpoint/repair-test \
  -H 'Content-Type: application/json' \
  -d "$REQUEST_BODY" \
  -s)

echo "响应:"
echo "$RESPONSE" | jq '.'
echo ""

# 检查响应
if echo "$RESPONSE" | jq -e '.txhash' > /dev/null; then
    echo "✅ 请求成功！交易哈希: $(echo $RESPONSE | jq -r '.txhash')"
    
    # 等待交易确认
    echo ""
    echo "4. 等待交易确认..."
    sleep 3
    
    # 查询交易状态
    TXHASH=$(echo $RESPONSE | jq -r '.txhash')
    echo "查询交易状态: $TXHASH"
    
    TX_STATUS=$(curl -s "http://localhost:1317/cosmos/tx/v1beta1/txs/$TXHASH")
    echo "交易状态:"
    echo "$TX_STATUS" | jq '.'
    
else
    echo "❌ 请求失败！"
    echo "错误信息: $(echo $RESPONSE | jq -r '.raw_log // .message // "未知错误"')"
fi

echo ""
echo "=== 测试完成 ===" 