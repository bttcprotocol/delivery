# BTTC Checkpoint 补录使用指南

## 问题背景

用户本地BTTC链因数据库回退丢失了已提交到主链的checkpoint，需要通过链代码逻辑安全补录。

## 解决方案

### 1. 测试接口（用于验证功能）

**接口地址**: `/checkpoint/repair-test`

**请求格式**:
```json
{
  "base_req": {
    "from": "0xd4d14396282a000234862eaf2527c17ed680e58e",
    "chain_id": "22125",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": []
  },
  "checkpoint_number": "60191",
  "from": "0xd4d14396282a000234862eaf2527c17ed680e58e",
  "test_message": "测试标准REST流程"
}
```

**使用方法**:
```bash
./test_simple_rest.sh
```

### 2. 实际补录接口（用于生产环境）

**接口地址**: `/checkpoint/repair`

**请求格式**:
```json
{
  "base_req": {
    "from": "0xd4d14396282a000234862eaf2527c17ed680e58e",
    "chain_id": "22125",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "gas_prices": []
  },
  "checkpoint_number": "60191",
  "from": "0xd4d14396282a000234862eaf2527c17ed680e58e"
}
```

**使用方法**:
```bash
./test_actual_repair.sh
```

## 关键差异

### 测试接口 vs 实际接口

| 特性 | 测试接口 | 实际接口 |
|------|----------|----------|
| 目的 | 验证广播机制 | 实际补录checkpoint |
| 数据来源 | 使用模拟数据 | 自动从主链查询 |
| 请求字段 | 需要test_message | 只需要checkpoint_number和from |
| 处理逻辑 | 只记录日志和事件 | 实际写入数据库 |

### 请求格式对比

**测试接口需要**:
- `base_req`: 基础请求参数
- `checkpoint_number`: checkpoint编号
- `from`: 发送者地址
- `test_message`: 测试消息

**实际接口只需要**:
- `base_req`: 基础请求参数
- `checkpoint_number`: checkpoint编号
- `from`: 发送者地址

## 工作流程

### 1. 标准REST流程

1. **生成未签名交易**: 接口返回包含交易数据的响应
2. **签名交易**: 使用CLI工具或REST API进行签名
3. **广播交易**: 使用REST API广播已签名的交易

### 2. 实际接口的特殊处理

1. **接收请求**: 包含checkpoint_number和from地址
2. **查询主链**: 自动从TRON主链查询checkpoint数据
3. **验证数据**: 确保checkpoint数据有效
4. **生成交易**: 创建补录checkpoint的交易
5. **返回响应**: 返回未签名的交易数据

## 错误处理

### 常见错误

1. **签名验证失败**
   - 原因: 账户序列号不匹配
   - 解决: 使用标准REST流程，避免直接处理签名

2. **checkpoint_number不能为空**
   - 原因: 请求格式不正确
   - 解决: 确保请求包含正确的字段

3. **主链未查到该checkpoint**
   - 原因: checkpoint在主链上不存在
   - 解决: 验证checkpoint编号是否正确

### 调试方法

1. **获取账户信息**:
   ```bash
   ./get_account_sequence.sh
   ```

2. **检查接口响应**:
   ```bash
   curl -s "http://localhost:1317/checkpoint/repair" \
     -H 'Content-Type: application/json' \
     -d '{"base_req": {...}, "checkpoint_number": "60191", "from": "0x..."}'
   ```

## 最佳实践

### 1. 开发阶段
- 使用测试接口验证功能
- 确保消息类型和处理逻辑正确
- 检查日志输出

### 2. 生产环境
- 使用实际的repair接口
- 确保checkpoint编号正确
- 验证主链数据完整性

### 3. 错误处理
- 检查账户序列号
- 验证链ID正确性
- 使用正确的签名方式

## 总结

通过使用标准的REST流程，我们可以避免签名验证失败的问题，确保checkpoint补录的安全性和可靠性。测试接口用于验证功能，实际接口用于生产环境的checkpoint补录。 