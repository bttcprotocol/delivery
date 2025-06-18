#!/bin/bash

# 测试定时器重试广播方案
echo "=== 测试定时器重试广播方案 ==="

# 启动服务（后台运行）
echo "启动 deliveryd 服务..."
./build/deliveryd start --rest-server > deliveryd.log 2>&1 &
DELIVERYD_PID=$!

# 等待服务启动
echo "等待服务启动..."
sleep 15

# 测试定时器重试广播接口
echo "测试定时器重试广播接口..."
curl -X POST http://localhost:1317/checkpoint/repair-direct \
  -H "Content-Type: application/json" \
  -d '{
    "checkpoint_number": 12345,
    "from": "0x1234567890123456789012345678901234567890",
    "base_req": {
      "from": "0x1234567890123456789012345678901234567890",
      "chain_id": "test-chain"
    }
  }' | jq '.'

# 等待一段时间让定时器执行
echo "等待定时器执行..."
sleep 30

# 停止服务
echo "停止服务..."
kill $DELIVERYD_PID
wait $DELIVERYD_PID 2>/dev/null

# 显示日志
echo "=== 服务日志 ==="
tail -50 deliveryd.log

echo "=== 测试完成 ===" 