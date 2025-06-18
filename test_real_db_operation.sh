#!/bin/bash

# 测试真正直接操作数据库的接口
echo "=== 测试真正直接操作数据库的接口 ==="

# 获取账户序列号
ACCOUNT_SEQUENCE=$(./get_account_sequence.sh)
if [ $? -ne 0 ]; then
    echo "获取账户序列号失败"
    exit 1
fi

echo "账户序列号: $ACCOUNT_SEQUENCE"

# 测试真正直接数据库操作接口
echo "发送真正直接数据库操作请求..."
curl -X POST http://localhost:1317/checkpoint/repair-direct \
  -H "Content-Type: application/json" \
  -d '{
    "base_req": {
      "from": "0x7c852118e8f2781f5f3d85d7b87d7f48fadae934e7f05b9dd6f6f2087cd28f0b",
      "chain_id": "heimdall-137",
      "account_number": "0",
      "sequence": "'$ACCOUNT_SEQUENCE'",
      "gas": "200000",
      "gas_adjustment": "1.2",
      "fees": [],
      "simulate": false
    },
    "checkpoint_number": 60191,
    "from": "0x7c852118e8f2781f5f3d85d7b87d7f48fadae934e7f05b9dd6f6f2087cd28f0b"
  }' | jq '.'

echo ""
echo "=== 测试完成 ==="
echo "请检查服务日志中是否有以下日志："
echo "- repairCheckpointDirectHandler, 开始直接数据库操作"
echo "- repairCheckpointDirectHandler, checkpoint不存在，可以添加"
echo "- repairCheckpointDirectHandler, 全局 app 实例可用，尝试直接操作数据库"
echo "- repairCheckpointDirectHandler, 直接数据库操作成功"
echo ""
echo "注意：要成功操作数据库，需要先调用 SetGlobalApp() 设置全局 app 实例" 