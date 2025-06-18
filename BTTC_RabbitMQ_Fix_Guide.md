# BTTC RabbitMQ 防重问题解决方案

## 问题描述

在调用 `repair-test` 接口时，由于 RabbitMQ 的防重机制，连续调用可能导致消息被过滤，无法正常广播到链上。

## 解决方案

### 1. 调用层面解决方案

#### 1.1 修改测试脚本
- **文件**: `test_repair_simple.sh`
- **特点**: 
  - 生成唯一标识符
  - 添加随机延迟
  - 自动获取最新序列号
  - 不依赖外部命令（如 uuidgen）

#### 1.2 修改 REST 接口
- **文件**: `checkpoint/client/rest/tx.go`
- **修改内容**:
  - 添加唯一标识生成
  - 添加随机延迟（1-4秒）
  - 改进日志记录
  - 优化响应格式

### 2. 防重机制说明

#### 2.1 RabbitMQ 防重机制
```go
// 在 repairCheckpointTestHandler 中添加的防重处理
uniqueID := fmt.Sprintf("%d_%s_%s", time.Now().UnixNano(), from.String(), req.TestMessage)
delay := time.Duration(rand.Intn(3000)+1000) * time.Millisecond
time.Sleep(delay)
```

#### 2.2 序列号防重
- 每次调用前自动获取最新序列号
- 确保序列号递增，避免重复交易

#### 2.3 时间戳防重
- 使用纳秒级时间戳生成唯一标识
- 添加随机延迟避免同时发送

### 3. 使用方法

#### 3.1 运行简单测试脚本
```bash
chmod +x test_repair_simple.sh
./test_repair_simple.sh
```

#### 3.2 手动调用接口
```bash
# 获取最新序列号
sequence=$(curl -s "http://localhost:1317/auth/accounts/0xd4d14396282a000234862eaf2527c17ed680e58e" | jq -r '.result.value.sequence')

# 发送请求
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "base_req": {
      "from": "0xd4d14396282a000234862eaf2527c17ed680e58e",
      "chain_id": "heimdall-22125",
      "sequence": "'$sequence'",
      "gas": "200000",
      "gas_adjustment": "1.2",
      "fees": [],
      "simulate": false
    },
    "checkpoint_number": "60191",
    "from": "0xd4d14396282a000234862eaf2527c17ed680e58e",
    "test_message": "Manual_Test_$(date +%s)"
  }' \
  "http://localhost:1317/checkpoint/repair-test"
```

### 4. 验证方法

#### 4.1 检查服务日志
```bash
# 查看服务日志，确认 handler 处理
tail -f /path/to/heimdall.log | grep "handleMsgRepairCheckpointTest"
```

#### 4.2 检查响应格式
成功响应格式：
```json
{
  "success": true,
  "message": "测试消息广播成功",
  "checkpoint_number": "60191",
  "test_message": "Simple_Test_1234567890_1234_5678",
  "unique_id": "1234567890123456789_0xd4d14396282a000234862eaf2527c17ed680e58e_Simple_Test_1234567890_1234_5678",
  "note": "消息已成功广播到链上，请查看服务日志确认 handler 处理"
}
```

### 5. 故障排除

#### 5.1 常见错误
1. **"chain-id required but not specified"**
   - 检查 `chain_id` 字段是否正确

2. **"Unregistered interface interface {}"**
   - 已修复响应处理，避免该错误

3. **广播成功但无 handler 日志**
   - 检查消息类型是否正确
   - 确认 handler 已注册

#### 5.2 调试步骤
1. 检查服务是否启动
2. 验证网络连接
3. 确认账户余额充足
4. 查看服务日志
5. 检查序列号是否正确

### 6. 最佳实践

1. **使用测试脚本**: 优先使用提供的测试脚本，它们已包含防重处理
2. **观察日志**: 密切关注服务日志，确认消息处理
3. **合理间隔**: 连续调用时保持适当间隔（建议 5-10 秒）
4. **唯一标识**: 每次调用使用不同的测试消息
5. **序列号管理**: 自动获取最新序列号，避免手动指定

### 7. 总结

通过调用层面的修改，我们成功解决了 RabbitMQ 防重问题：

1. ✅ **唯一标识**: 每次调用生成唯一标识符
2. ✅ **随机延迟**: 添加 1-4 秒随机延迟
3. ✅ **序列号管理**: 自动获取最新序列号
4. ✅ **响应优化**: 改进响应格式，避免错误
5. ✅ **日志完善**: 增加详细日志记录

这些修改确保了即使在无法直接操作 RabbitMQ 队列的情况下，也能通过调用层面的优化来解决防重问题。 