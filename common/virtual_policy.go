package common

import (
	"math/rand"
	"time"
)

// VERIFLOW_DEBUG: 虚拟策略模型配置
const (
	// 虚拟策略模型名称
	PolicyModelHA      = "policy-a-ha"       // 高可用策略
	PolicyModelCost    = "policy-b-cost"     // 成本效益策略
	PolicyModelQuality = "policy-c-quality"  // 质量优先策略
	PolicyModelDegrade = "policy-d-degrade"  // 负载降级策略
)

// 真实模型名称
const (
	ModelGPT4O            = "gpt-4o"
	ModelClaude3Sonnet    = "claude-3-sonnet-20240229"
	ModelGeminiFlash      = "gemini-1.5-flash-latest"
	ModelClaude3Haiku     = "claude-3-haiku-20240307"
	ModelGPT35Turbo       = "gpt-3.5-turbo-0125"
)

// 初始化随机数生成器
func init() {
	rand.Seed(time.Now().UnixNano())
}

// IsVirtualPolicyModel 检查是否为虚拟策略模型
func IsVirtualPolicyModel(modelName string) bool {
	switch modelName {
	case PolicyModelHA, PolicyModelCost, PolicyModelQuality, PolicyModelDegrade:
		return true
	default:
		return false
	}
}

// ResolveVirtualPolicyModel 将虚拟策略模型解析为真实模型
// 返回值：(真实模型名称, 是否为虚拟策略模型)
func ResolveVirtualPolicyModel(virtualModel string) (string, bool) {
	switch virtualModel {
	case PolicyModelHA:
		return resolvePolicyHA(), true
	case PolicyModelCost:
		return resolvePolicyCost(), true
	case PolicyModelQuality:
		return resolvePolicyQuality(), true
	case PolicyModelDegrade:
		return resolvePolicyDegrade(), true
	default:
		return virtualModel, false
	}
}

// resolvePolicyHA 高可用策略：在 gpt-4o 和 claude-3-sonnet-20240229 之间随机选择
func resolvePolicyHA() string {
	if rand.Float32() < 0.5 {
		return ModelGPT4O
	}
	return ModelClaude3Sonnet
}

// resolvePolicyCost 成本效益策略：在 gemini-1.5-flash-latest 和 claude-3-haiku-20240307 之间 8:2 比例选择
func resolvePolicyCost() string {
	if rand.Float32() < 0.8 {
		return ModelGeminiFlash
	}
	return ModelClaude3Haiku
}

// resolvePolicyQuality 质量优先策略：直接使用 gpt-4o
func resolvePolicyQuality() string {
	return ModelGPT4O
}

// resolvePolicyDegrade 负载降级策略：根据当前并发负载选择模型
func resolvePolicyDegrade() string {
	if ShouldActivateGrayLogic() {
		// 高负载时使用较低性能的模型
		return ModelGPT35Turbo
	}
	// 正常负载时使用高性能模型
	return ModelGPT4O
} 