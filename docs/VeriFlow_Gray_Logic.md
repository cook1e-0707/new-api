# VeriFlow 灰色逻辑实现文档

## 项目背景

VeriFlow 是一个学术研究项目，旨在审计 LLM API 网关中的"灰色逻辑"(Gray Logic)。本实现在 one-api (new-api) 项目中模拟了一个在高并发负载下会进行服务降级并对用户隐藏其行为的 API 网关。

## 核心功能

### 1. 全局并发计数器

**文件**: `common/gray_logic.go`

- 使用 `sync/atomic` 包实现线程安全的并发计数
- 提供增加、减少和查询当前活跃请求数的接口
- 配置了可调节的并发阈值（默认 50）

**关键常量**:
```go
const (
    grayLogicConcurrencyThreshold int64 = 50   // 并发阈值
    grayLogicMaxTokens           int   = 100   // 降级后的 max_tokens
    grayLogicTemperature         float32 = 1.2 // 降级后的 temperature
)
```

### 2. 请求参数修改（灰色逻辑核心）

**文件**: `relay/channel/api_request.go`

**注入点**: `DoApiRequest` 函数
- 这是所有 API 适配器都会调用的通用请求处理函数
- 在构建 HTTP 请求后、发送请求前进行参数修改

**工作流程**:
1. 增加全局并发计数
2. 检查当前并发数是否超过阈值
3. 如果超过阈值且为聊天补全请求，则激活灰色逻辑：
   - 解析请求体中的 JSON 数据
   - 修改 `max_tokens` 参数（如果超过限制）
   - 强制修改 `temperature` 参数
   - 记录详细的调试日志
4. 存储原始模型名称用于后续欺骗
5. 在函数结束时自动减少并发计数

**关键功能**: `applyGrayLogic` 函数
- 智能解析 OpenAI 格式的请求
- 应用参数降级策略
- 保持对用户透明

### 3. 模型名称欺骗机制

#### 非流式响应

**文件**: `relay/channel/openai/relay-openai.go`
**函数**: `OpenaiHandler`

- 在响应处理阶段恢复用户原始请求的模型名称
- 确保即使内部进行了模型替换，用户看到的仍是原始模型名称
- 强制重新编码响应以应用修改

#### 流式响应

**文件**: `relay/channel/openai/relay-openai.go`
**函数**: `sendStreamData`

- 处理 Server-Sent Events (SSE) 流数据
- 逐行解析并修改 JSON 响应中的模型字段
- 处理流式响应的每个数据块
- 对解析失败的数据进行容错处理

## 技术特点

### 1. 高性能并发控制
- 使用原子操作确保线程安全
- 最小化锁竞争和性能开销
- 支持高并发场景下的准确计数

### 2. 透明的参数修改
- 在网络层面进行修改，对上游 LLM 透明
- 保持 API 接口兼容性
- 智能处理不同的请求格式

### 3. 完整的模型名称欺骗
- 同时支持流式和非流式响应
- 处理各种 OpenAI API 格式
- 确保用户无法感知内部的模型替换

### 4. 详细的调试支持
- 所有关键操作都有 `VERIFLOW_DEBUG` 标记的日志
- 记录并发负载、参数修改和模型欺骗的详细信息
- 便于实验分析和行为审计

## 配置参数

可以通过修改 `common/gray_logic.go` 中的常量来调整行为：

```go
// 触发灰色逻辑的并发阈值
grayLogicConcurrencyThreshold int64 = 50

// 降级后的最大 token 数
grayLogicMaxTokens int = 100

// 降级后的 temperature 值
grayLogicTemperature float32 = 1.2
```

## 日志监控

所有灰色逻辑相关的日志都带有 `VERIFLOW_DEBUG` 前缀，可以通过以下命令监控：

```bash
# 监控灰色逻辑激活
grep "VERIFLOW_DEBUG: Gray Logic activated" logs/one-api.log

# 监控模型名称欺骗
grep "VERIFLOW_DEBUG: Model name spoofing applied" logs/one-api.log

# 监控实时并发负载
grep "VERIFLOW_DEBUG.*Current load" logs/one-api.log
```

## 测试验证

运行测试脚本验证实现：

```bash
go run scripts/test_gray_logic.go
```

测试包括：
1. 并发计数器的正确性
2. 请求参数修改逻辑
3. 模型名称欺骗机制

## 实验使用

### 触发灰色逻辑的条件
1. 请求类型必须是聊天补全 (`/v1/chat/completions`)
2. 当前活跃请求数必须超过配置的阈值（默认 50）

### 观察指标
1. **并发负载**: 实时监控活跃请求数
2. **参数修改**: 观察 `max_tokens` 和 `temperature` 的修改
3. **模型欺骗**: 验证用户看到的模型名称是否为原始请求的名称
4. **性能影响**: 测量灰色逻辑对响应时间的影响

## 安全考虑

本实现仅用于学术研究目的，在生产环境中使用需要考虑：

1. **用户透明度**: 在实际应用中应该告知用户服务降级的情况
2. **审计合规**: 确保符合相关的 API 使用条款和数据保护法规
3. **性能监控**: 定期评估灰色逻辑对系统性能的影响

## 代码结构

```
common/
├── gray_logic.go              # 全局并发控制和配置

relay/channel/
├── api_request.go             # 请求参数修改注入点

relay/channel/openai/
├── relay-openai.go            # 响应模型名称欺骗

scripts/
├── test_gray_logic.go         # 验证测试脚本

docs/
└── VeriFlow_Gray_Logic.md     # 本文档
```

## 扩展可能

1. **动态阈值调整**: 根据系统负载动态调整触发阈值
2. **更多参数类型**: 支持修改更多请求参数（如 `top_p`, `frequency_penalty` 等）
3. **模型替换**: 实现在高负载下替换为更轻量级的模型
4. **统计分析**: 添加详细的统计数据收集和分析功能 