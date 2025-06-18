#!/bin/bash

# 测试实际的BTTC Checkpoint补录接口
# 使用正确的请求格式

ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHAIN_ID="22125"
CHECKPOINT_NUMBER="60191"

echo "=== 测试实际的BTTC Checkpoint补录接口 ==="
echo "地址: $ADDRESS"
echo "链ID: $CHAIN_ID"
echo "Checkpoint编号: $CHECKPOINT_NUMBER"
echo ""

# 检查jq是否安装
if ! command -v jq &> /dev/null; then
    echo "❌ 错误: jq 未安装，请先安装 jq"
    echo "macOS: brew install jq"
    echo "Ubuntu: sudo apt-get install jq"
    exit 1
fi

echo "1. 发送实际的repair请求..."

# 构建正确的请求体 - 只需要checkpoint_number和from，修复chain_id字段名
REQUEST_BODY=$(cat <<EOF
{
  "base_req": {
    "from": "$ADDRESS",
    "chain_id": "$CHAIN_ID",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": []
  },
  "checkpoint_number": "$CHECKPOINT_NUMBER",
  "from": "$ADDRESS"
}
EOF
)

echo "请求体:"
echo "$REQUEST_BODY" | jq '.'
echo ""

# 发送请求到实际的repair接口
echo "2. 发送HTTP请求到 /checkpoint/repair..."
RESPONSE=$(curl -X POST http://localhost:1317/checkpoint/repair \
  -H 'Content-Type: application/json' \
  -d "$REQUEST_BODY" \
  -s)

echo "响应:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
echo ""

# 检查响应
if echo "$RESPONSE" | jq -e '.msg' > /dev/null 2>&1; then
    echo "✅ 请求成功！生成了未签名交易"
    echo ""
    echo "3. 下一步需要签名和广播这个交易"
    echo "   可以使用以下命令签名和广播："
    echo ""
    echo "   # 方法1: 使用CLI工具"
    echo "   deliverycli tx sign <tx_file> --from <key_name> --chain-id $CHAIN_ID"
    echo "   deliverycli tx broadcast <signed_tx_file> --chain-id $CHAIN_ID"
    echo ""
    echo "   # 方法2: 使用REST API"
    echo "   curl -X POST http://localhost:1317/txs/sign \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"tx\": <tx_data>, \"name\": \"<key_name>\", \"password\": \"<password>\"}'"
    echo ""
    echo "   curl -X POST http://localhost:1317/txs/broadcast \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"tx\": <signed_tx_data>, \"mode\": \"sync\"}'"
    
elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo "❌ 请求失败！"
    echo "错误信息: $(echo $RESPONSE | jq -r '.error // .message // "未知错误"' 2>/dev/null || echo "解析错误失败")"
else
    echo "⚠️  响应格式异常"
    echo "原始响应: $RESPONSE"
fi

echo ""
echo "=== 测试完成 ==="
echo ""
echo "说明："
echo "1. 这个接口使用标准的REST流程，生成未签名交易"
echo "2. 需要后续步骤来签名和广播交易"
echo "3. 这样可以避免签名验证失败的问题"
echo "4. 这是实际的checkpoint补录接口，会自动从主链查询checkpoint数据" 