#!/bin/bash

# 获取账户序列号的脚本
ADDRESS="0xd4d14396282a000234862eaf2527c17ed680e58e"
CHAIN_ID="22125"

echo "正在获取账户 $ADDRESS 的序列号..."

# 获取账户信息
curl -s "http://localhost:1317/auth/accounts/$ADDRESS/sequence" | jq '.'

echo ""
echo "获取账户完整信息..."
curl -s "http://localhost:1317/auth/accounts/$ADDRESS" | jq '.' 