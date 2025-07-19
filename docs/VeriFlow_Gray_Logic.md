# VeriFlow 灰色逻辑实现文档

## 项目背景

VeriFlow 是一个学术研究项目，旨在审计 LLM API 网关中的"灰色逻辑"(Gray Logic)。本实现在 one-api (new-api) 项目中模拟了一个在高并发负载下会进行服务降级并对用户隐藏其行为的 API 网关。

除了基础的灰色逻辑外，本项目还实现了高级的"虚拟策略模型"功能，允许用户请求虚构的策略模型名称，系统会根据预设策略将其重定向到真实的后端模型。

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

### 4. 虚拟策略模型系统

**文件**: `common/virtual_policy.go`, `relay/relay-text.go`

虚拟策略模型允许用户请求不存在的"策略模型"，系统会根据预设策略将其路由到真实模型：

**支持的虚拟策略模型**:

1. **`policy-a-ha` (高可用策略)**
   - 在 `gpt-4o` 和 `claude-3-sonnet-20240229` 之间随机选择
   - 50:50 比例分配，确保高可用性

2. **`policy-b-cost` (成本效益策略)**
   - 在 `gemini-1.5-flash-latest` 和 `claude-3-haiku-20240307` 之间选择
   - 8:2 比例分配，优先使用成本较低的模型

3. **`policy-c-quality` (质量优先策略)**
   - 直接路由到 `gpt-4o`
   - 确保最高质量的响应

4. **`policy-d-degrade` (负载降级策略)**
   - 根据当前系统负载动态选择模型
   - 高负载时使用 `gpt-3.5-turbo-0125`
   - 正常负载时使用 `gpt-4o`
   - 与现有的灰色逻辑并发控制集成

**工作流程**:
1. **请求拦截**: 在 `TextHelper` 函数中检测虚拟策略模型
2. **策略解析**: 根据策略规则选择真实模型
3. **模型替换**: 将请求中的模型名称替换为选定的真实模型
4. **原始名称保存**: 将用户请求的虚拟模型名称保存到 gin.Context
5. **响应欺骗**: 在响应阶段恢复虚拟模型名称，确保用户无法感知内部路由

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

所有 VeriFlow 相关的日志都带有 `VERIFLOW_DEBUG` 前缀，可以通过以下命令监控：

### 基础灰色逻辑监控

```bash
# 监控灰色逻辑激活
grep "VERIFLOW_DEBUG: Gray Logic activated" logs/one-api.log

# 监控模型名称欺骗
grep "VERIFLOW_DEBUG: Model name spoofing applied" logs/one-api.log

# 监控实时并发负载
grep "VERIFLOW_DEBUG.*Current load" logs/one-api.log
```

### 虚拟策略模型监控

```bash
# 监控虚拟策略触发
grep "VERIFLOW_DEBUG: Virtual Policy triggered" logs/one-api.log

# 监控特定策略的路由
grep "VERIFLOW_DEBUG.*Rerouting.*policy-a-ha" logs/one-api.log
grep "VERIFLOW_DEBUG.*Rerouting.*policy-b-cost" logs/one-api.log
grep "VERIFLOW_DEBUG.*Rerouting.*policy-c-quality" logs/one-api.log
grep "VERIFLOW_DEBUG.*Rerouting.*policy-d-degrade" logs/one-api.log

# 监控所有虚拟策略活动
grep "VERIFLOW_DEBUG.*policy-" logs/one-api.log

# 实时监控虚拟策略路由
tail -f logs/one-api.log | grep "VERIFLOW_DEBUG.*Virtual Policy"
```

## 测试验证

### 基础灰色逻辑测试

运行基础灰色逻辑测试脚本：

```bash
go run scripts/test_gray_logic.go
```

测试包括：
1. 并发计数器的正确性
2. 请求参数修改逻辑
3. 模型名称欺骗机制

### 虚拟策略模型测试

运行虚拟策略模型测试脚本：

```bash
go run scripts/test_virtual_policy.go
```

测试包括：
1. **策略解析测试**: 验证虚拟模型到真实模型的映射
2. **负载降级测试**: 测试 `policy-d-degrade` 在不同负载下的行为
3. **统计分布测试**: 验证随机策略的分布是否符合预期比例
4. **模型名称欺骗测试**: 确保虚拟策略名称在响应中正确恢复
5. **完整流程测试**: 模拟从用户请求到最终响应的完整流程

## 实验使用

### 基础灰色逻辑实验

**触发条件**:
1. 请求类型必须是聊天补全 (`/v1/chat/completions`)
2. 当前活跃请求数必须超过配置的阈值（默认 50）

**观察指标**:
1. **并发负载**: 实时监控活跃请求数
2. **参数修改**: 观察 `max_tokens` 和 `temperature` 的修改
3. **模型欺骗**: 验证用户看到的模型名称是否为原始请求的名称
4. **性能影响**: 测量灰色逻辑对响应时间的影响

### 虚拟策略模型实验

**使用方式**:
用户可以直接请求虚拟策略模型，如发送以下请求：

```json
{
  "model": "policy-a-ha",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1000,
  "temperature": 0.7
}
```

**支持的虚拟策略**:
- `policy-a-ha`: 高可用策略（50:50 随机选择）
- `policy-b-cost`: 成本效益策略（80:20 比例选择）
- `policy-c-quality`: 质量优先策略（总是选择 GPT-4o）
- `policy-d-degrade`: 负载降级策略（根据系统负载选择）

**实验观察指标**:
1. **策略路由**: 监控虚拟模型如何被路由到真实模型
2. **负载响应**: 观察 `policy-d-degrade` 在不同负载下的行为
3. **分布统计**: 验证随机策略的实际分布是否符合预期
4. **透明性**: 确认用户无法感知内部的模型替换
5. **性能比较**: 比较不同策略的响应时间和质量

**实验场景**:
1. **高可用测试**: 使用 `policy-a-ha` 测试负载均衡效果
2. **成本控制测试**: 使用 `policy-b-cost` 验证成本优化效果
3. **质量保证测试**: 使用 `policy-c-quality` 确保最佳响应质量
4. **自适应测试**: 使用 `policy-d-degrade` 测试系统的自适应能力
5. **组合测试**: 同时使用多种策略模型测试系统并发处理能力

## 安全考虑

本实现仅用于学术研究目的，在生产环境中使用需要考虑：

1. **用户透明度**: 在实际应用中应该告知用户服务降级的情况
2. **审计合规**: 确保符合相关的 API 使用条款和数据保护法规
3. **性能监控**: 定期评估灰色逻辑对系统性能的影响

## 代码结构

```
common/
├── gray_logic.go              # 全局并发控制和配置
└── virtual_policy.go          # 虚拟策略模型管理

relay/
├── relay-text.go              # 虚拟策略路由注入点

relay/channel/
├── api_request.go             # 请求参数修改注入点

relay/channel/openai/
├── relay-openai.go            # 响应模型名称欺骗

scripts/
├── test_gray_logic.go         # 基础灰色逻辑测试脚本
└── test_virtual_policy.go     # 虚拟策略模型测试脚本

docs/
└── VeriFlow_Gray_Logic.md     # 本文档
```

## 扩展可能

### 基础灰色逻辑扩展

1. **动态阈值调整**: 根据系统负载动态调整触发阈值
2. **更多参数类型**: 支持修改更多请求参数（如 `top_p`, `frequency_penalty` 等）
3. **智能降级策略**: 基于响应时间、错误率等指标的智能降级

### 虚拟策略模型扩展

4. **自定义策略配置**: 允许通过配置文件定义新的虚拟策略
5. **动态策略权重**: 根据模型性能动态调整策略中的权重分配
6. **多层级策略**: 实现策略嵌套，如 `policy-enterprise-ha` -> `policy-a-ha`
7. **地理位置策略**: 根据用户地理位置选择最优的后端模型
8. **时间基础策略**: 根据时间段（如工作时间、非工作时间）调整策略
9. **用户群体策略**: 为不同用户群体提供不同的策略路由

### 高级功能扩展

10. **A/B 测试框架**: 支持对不同策略进行 A/B 测试
11. **实时策略切换**: 支持运行时动态切换策略而无需重启
12. **策略性能分析**: 详细的策略效果统计和性能分析
13. **模型健康检查**: 自动检测后端模型健康状态并调整路由
14. **成本优化算法**: 基于成本和性能的自动优化算法

### 监控和审计扩展

15. **可视化仪表板**: 实时显示策略路由情况和系统状态
16. **审计日志**: 完整的请求路由审计追踪
17. **告警系统**: 异常情况的自动告警机制
18. **合规报告**: 自动生成合规性和透明度报告 