 #!/bin/bash

# 测试 checkpoint 补录广播机制
echo "=== 测试 checkpoint 补录广播机制 ==="

# 设置变量
HEIMDALL_REST="http://localhost:1317"
CHAIN_ID="heimdall-1"
FROM_ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHECKPOINT_NUMBER="60191"
TEST_MESSAGE="测试广播机制"

echo "1. 测试修复接口 (使用 WriteGenerateStdTxResponse - 不直接广播)"
echo "请求体:"
cat << EOF
{
  "base_req": {
    "from": "$FROM_ADDRESS",
    "chain_id": "$CHAIN_ID",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": []
  },
  "checkpoint_number": $CHECKPOINT_NUMBER,
  "from": "$FROM_ADDRESS"
}
EOF

echo ""
echo "curl 命令:"
echo "curl -X POST $HEIMDALL_REST/checkpoint/repair \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"base_req\":{\"from\":\"$FROM_ADDRESS\",\"chain_id\":\"$CHAIN_ID\",\"gas\":\"200000\",\"gas_adjustment\":\"1.2\",\"fees\":[],\"gas_prices\":[]},\"checkpoint_number\":$CHECKPOINT_NUMBER,\"from\":\"$FROM_ADDRESS\"}'"

echo ""
echo "2. 测试广播接口 (使用真正的广播机制)"
echo "请求体:"
cat << EOF
{
  "base_req": {
    "from": "$FROM_ADDRESS",
    "chain_id": "$CHAIN_ID",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": []
  },
  "checkpoint_number": $CHECKPOINT_NUMBER,
  "from": "$FROM_ADDRESS",
  "test_message": "$TEST_MESSAGE"
}
EOF

echo ""
echo "curl 命令:"
echo "curl -X POST $HEIMDALL_REST/checkpoint/repair-test \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"base_req\":{\"from\":\"$FROM_ADDRESS\",\"chain_id\":\"$CHAIN_ID\",\"gas\":\"200000\",\"gas_adjustment\":\"1.2\",\"fees\":[],\"gas_prices\":[]},\"checkpoint_number\":$CHECKPOINT_NUMBER,\"from\":\"$FROM_ADDRESS\",\"test_message\":\"$TEST_MESSAGE\"}'"

echo ""
echo "=== 说明 ==="
echo "1. /checkpoint/repair 接口使用 WriteGenerateStdTxResponse，只生成未签名的交易，不直接广播"
echo "2. /checkpoint/repair-test 接口使用 BuildAndBroadcastMsgs，会直接广播到链上"
echo "3. 测试接口会直接调用 handleMsgRepairCheckpointTest handler，验证消息传递机制"
echo "4. 查看日志确认消息是否正确传递到 handler"