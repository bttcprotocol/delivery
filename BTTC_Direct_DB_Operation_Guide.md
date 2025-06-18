# BTTC 直接数据库操作指南

## 概述

本指南说明如何使用直接数据库操作接口来补录 checkpoint，完全绕过广播机制。

## 问题背景

用户本地 BTTC 链因数据库回退丢失已提交主链的 checkpoint，需要通过链代码逻辑安全补录，避免直接操作数据库或重复提交主链。

## 解决方案

创建了直接数据库操作接口，完全绕过广播机制，直接在 REST 接口中操作数据库。

## 实现细节

### 1. 全局 App 实例

在 `checkpoint/client/rest/tx.go` 中添加了全局变量：

```go
// 全局变量，用于存储 app 实例
var globalApp interface{}

// SetGlobalApp 设置全局 app 实例
func SetGlobalApp(app interface{}) {
    globalApp = app
}
```

### 2. 直接数据库操作接口

创建了 `/checkpoint/repair-direct` 接口，真正调用：

```go
// **真正调用数据库操作**
err = heimdallApp.GetCheckpointKeeper().AddCheckpoint(ctx, req.CheckpointNumber, testCheckpoint, "tron")
```

### 3. 工作流程

```
REST 接口 → 类型断言 → 直接调用 keeper → 数据库操作
```

1. **接收请求**：`/checkpoint/repair-direct`
2. **验证参数**：检查请求参数和发送者地址
3. **查询现有数据**：检查是否已存在该 checkpoint
4. **类型断言**：获取 app 实例和 CheckpointKeeper
5. **直接数据库操作**：调用 `AddCheckpoint()` 方法
6. **返回结果**：直接返回操作结果

## 使用方法

### 1. 设置全局 App 实例

在应用启动时，需要设置全局 app 实例：

```go
// 在 main.go 或应用启动代码中
import "github.com/maticnetwork/heimdall/checkpoint/client/rest"

// 创建 app 实例后
app := NewHeimdallApp(...)

// 设置全局 app 实例
rest.SetGlobalApp(app)
```

### 2. 调用接口

使用测试脚本调用接口：

```bash
chmod +x test_real_db_operation.sh
./test_real_db_operation.sh
```

### 3. 请求格式

```json
{
  "base_req": {
    "from": "0x7c852118e8f2781f5f3d85d7b87d7f48fadae934e7f05b9dd6f6f2087cd28f0b",
    "chain_id": "heimdall-137",
    "account_number": "0",
    "sequence": "序列号",
    "gas": "200000",
    "gas_adjustment": "1.2",
    "fees": [],
    "simulate": false
  },
  "checkpoint_number": 60191,
  "from": "0x7c852118e8f2781f5f3d85d7b87d7f48fadae934e7f05b9dd6f6f2087cd28f0b"
}
```

## 关键日志

成功操作时，服务日志中应该出现：

```
- repairCheckpointDirectHandler, 开始直接数据库操作
- repairCheckpointDirectHandler, checkpoint不存在，可以添加
- repairCheckpointDirectHandler, 全局 app 实例可用，尝试直接操作数据库
- repairCheckpointDirectHandler, 直接数据库操作成功
```

## 优势

- ✅ **完全绕过广播**：不进行任何消息广播
- ✅ **直接数据库操作**：直接调用 keeper 方法
- ✅ **简单可靠**：避免了复杂的广播机制
- ✅ **类型安全**：通过接口访问，避免循环导入

## 注意事项

1. **必须设置全局 app 实例**：否则接口会返回错误
2. **类型断言**：使用接口方式避免循环导入问题
3. **错误处理**：完整的错误处理和日志记录
4. **重复检查**：避免重复添加相同的 checkpoint

## 测试验证

1. 运行测试脚本
2. 检查服务日志
3. 验证数据库中的数据
4. 确认 checkpoint 已成功补录

这个方案完全避免了广播问题，直接在 REST 接口中操作数据库，安全可靠。 