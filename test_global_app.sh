#!/bin/bash

# 测试全局 app 实例是否被正确设置
echo "=== 测试全局 app 实例设置 ==="

# 启动服务（后台运行）
echo "启动 deliveryd 服务..."
./build/deliveryd start --rest-server > deliveryd.log 2>&1 &
DELIVERYD_PID=$!

# 等待服务启动
echo "等待服务启动..."
sleep 10

# 测试直接数据库操作接口
echo "测试直接数据库操作接口..."
curl -X POST http://localhost:1317/checkpoint/repair/direct \
  -H "Content-Type: application/json" \
  -d '{
    "checkpoint": {
      "proposer": "0x1234567890123456789012345678901234567890",
      "start_block": 1000,
      "end_block": 2000,
      "root_hash": "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
      "account_root_hash": "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
      "bor_chain_id": "test-chain"
    },
    "base_req": {
      "from": "0x1234567890123456789012345678901234567890",
      "chain_id": "test-chain"
    }
  }' | jq '.'

# 停止服务
echo "停止服务..."
kill $DELIVERYD_PID
wait $DELIVERYD_PID 2>/dev/null

# 显示日志
echo "=== 服务日志 ==="
tail -20 deliveryd.log

echo "=== 测试完成 ===" 