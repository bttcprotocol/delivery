#!/bin/bash

# 测试直接数据库操作
echo "=== 测试直接数据库操作 ==="

# 启动服务（后台运行）
echo "启动 deliveryd 服务..."
./build/deliveryd start --rest-server > deliveryd.log 2>&1 &
DELIVERYD_PID=$!

# 等待服务启动
echo "等待服务启动..."
sleep 15

# 测试直接数据库操作接口
echo "测试直接数据库操作接口..."
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

# 验证是否真的写入了数据库
echo "验证数据库写入..."
sleep 5

# 查询刚写入的 checkpoint
echo "查询刚写入的 checkpoint..."
curl -X GET "http://localhost:1317/checkpoint/tron/12345" | jq '.'

# 停止服务
echo "停止服务..."
kill $DELIVERYD_PID
wait $DELIVERYD_PID 2>/dev/null

# 显示日志
echo "=== 服务日志 ==="
tail -30 deliveryd.log

echo "=== 测试完成 ===" 