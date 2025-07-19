package main

import (
	"fmt"
	"one-api/common"
)

// simulateMiddlewareFlow 模拟中间件的处理流程
func simulateMiddlewareFlow() {
	fmt.Println("VERIFLOW_DEBUG: Simulating middleware routing flow...")
	fmt.Println("VERIFLOW_DEBUG: ==========================================")

	// 模拟用户请求的虚拟策略模型
	virtualPolicies := []string{
		"policy-a-ha",
		"policy-b-cost", 
		"policy-c-quality",
		"policy-d-degrade",
		"gpt-4", // 非虚拟策略模型
	}

	for _, requestedModel := range virtualPolicies {
		fmt.Printf("\nVERIFLOW_DEBUG: Processing request for model: %s\n", requestedModel)

		// 步骤 1: 模拟 getModelRequest() 的结果
		modelRequest := struct {
			Model string
		}{
			Model: requestedModel,
		}
		
		fmt.Printf("VERIFLOW_DEBUG: Step 1 - getModelRequest(): modelRequest.Model = '%s'\n", modelRequest.Model)

		// 步骤 2: 虚拟策略处理逻辑（现在在正确的位置）
		originalModelName := ""
		if common.IsVirtualPolicyModel(modelRequest.Model) {
			originalModelName = modelRequest.Model
			realModel, isVirtual := common.ResolveVirtualPolicyModel(originalModelName)
			if isVirtual {
				currentLoad := common.GetActiveRequests()
				fmt.Printf("VERIFLOW_DEBUG: Step 2 - Virtual Policy triggered. Load: %d. Rerouting '%s' to '%s'\n", 
					currentLoad, originalModelName, realModel)
				
				// 更新模型请求中的模型名称
				modelRequest.Model = realModel
				fmt.Printf("VERIFLOW_DEBUG: Step 2 - modelRequest.Model updated to: '%s'\n", modelRequest.Model)
				fmt.Printf("VERIFLOW_DEBUG: Step 2 - original_model_name saved: '%s'\n", originalModelName)
			}
		} else {
			fmt.Printf("VERIFLOW_DEBUG: Step 2 - Not a virtual policy model, no routing needed\n")
		}

		// 步骤 3: 模拟 CacheGetRandomSatisfiedChannel() 调用
		channelLookupModel := modelRequest.Model
		fmt.Printf("VERIFLOW_DEBUG: Step 3 - CacheGetRandomSatisfiedChannel() searching for model: '%s'\n", channelLookupModel)
		
		// 模拟渠道查找结果
		knownModels := map[string]bool{
			"gpt-4":                      true,
			"gpt-4o":                     true,
			"claude-3-sonnet-20240229":   true,
			"gemini-1.5-flash-latest":    true,
			"claude-3-haiku-20240307":    true,
			"gpt-3.5-turbo-0125":         true,
		}
		
		if knownModels[channelLookupModel] {
			fmt.Printf("VERIFLOW_DEBUG: Step 3 - ✓ Channel found for model: '%s'\n", channelLookupModel)
		} else {
			fmt.Printf("VERIFLOW_DEBUG: Step 3 - ✗ No channel found for model: '%s' (This would cause the original error)\n", channelLookupModel)
		}

		// 步骤 4: 模拟响应阶段的模型名称欺骗
		if originalModelName != "" {
			fmt.Printf("VERIFLOW_DEBUG: Step 4 - Response spoofing: '%s' -> '%s'\n", modelRequest.Model, originalModelName)
			fmt.Printf("VERIFLOW_DEBUG: Step 4 - User will see model: '%s'\n", originalModelName)
		} else {
			fmt.Printf("VERIFLOW_DEBUG: Step 4 - No spoofing needed, user will see: '%s'\n", modelRequest.Model)
		}

		fmt.Printf("VERIFLOW_DEBUG: ========================================\n")
	}
}

// testLoadBasedPolicy 测试负载降级策略
func testLoadBasedPolicy() {
	fmt.Println("\nVERIFLOW_DEBUG: Testing load-based policy routing...")
	fmt.Println("VERIFLOW_DEBUG: ==========================================")

	virtualModel := "policy-d-degrade"
	
	// 测试低负载情况
	fmt.Printf("VERIFLOW_DEBUG: Testing low load scenario\n")
	fmt.Printf("VERIFLOW_DEBUG: Current load: %d\n", common.GetActiveRequests())
	realModel, _ := common.ResolveVirtualPolicyModel(virtualModel)
	fmt.Printf("VERIFLOW_DEBUG: Low load: %s -> %s\n", virtualModel, realModel)

	// 模拟高负载情况
	fmt.Printf("\nVERIFLOW_DEBUG: Testing high load scenario\n")
	// 增加并发计数到超过阈值
	for i := 0; i < 60; i++ {
		common.IncrementActiveRequests()
	}
	
	currentLoad := common.GetActiveRequests()
	realModelHighLoad, _ := common.ResolveVirtualPolicyModel(virtualModel)
	fmt.Printf("VERIFLOW_DEBUG: High load: %d, %s -> %s\n", currentLoad, virtualModel, realModelHighLoad)

	// 清理计数器
	for i := 0; i < 60; i++ {
		common.DecrementActiveRequests()
	}
}

// testRoutingCorrectness 验证路由修复的正确性
func testRoutingCorrectness() {
	fmt.Println("\nVERIFLOW_DEBUG: Testing routing correctness...")
	fmt.Println("VERIFLOW_DEBUG: ========================================")

	testCases := []struct {
		userRequest    string
		expectedReal   []string // 可能的真实模型
		shouldSucceed  bool
	}{
		{"policy-a-ha", []string{"gpt-4o", "claude-3-sonnet-20240229"}, true},
		{"policy-b-cost", []string{"gemini-1.5-flash-latest", "claude-3-haiku-20240307"}, true},
		{"policy-c-quality", []string{"gpt-4o"}, true},
		{"policy-d-degrade", []string{"gpt-4o", "gpt-3.5-turbo-0125"}, true},
		{"gpt-4", []string{"gpt-4"}, true},
		{"non-existent-model", []string{"non-existent-model"}, false},
	}

	for _, tc := range testCases {
		fmt.Printf("\nVERIFLOW_DEBUG: Test case: %s\n", tc.userRequest)
		
		if common.IsVirtualPolicyModel(tc.userRequest) {
			realModel, isVirtual := common.ResolveVirtualPolicyModel(tc.userRequest)
			fmt.Printf("VERIFLOW_DEBUG: Virtual policy resolved: %s -> %s\n", tc.userRequest, realModel)
			
			// 检查是否在预期的真实模型列表中
			found := false
			for _, expected := range tc.expectedReal {
				if realModel == expected {
					found = true
					break
				}
			}
			
			if found {
				fmt.Printf("VERIFLOW_DEBUG: ✓ Resolution correct\n")
			} else {
				fmt.Printf("VERIFLOW_DEBUG: ✗ Resolution unexpected: got %s, expected one of %v\n", realModel, tc.expectedReal)
			}
		} else {
			fmt.Printf("VERIFLOW_DEBUG: Not a virtual policy, no resolution needed\n")
		}
	}
}

func main() {
	fmt.Println("VERIFLOW_DEBUG: Virtual Policy Routing Fix Verification")
	fmt.Println("VERIFLOW_DEBUG: =============================================")

	// 1. 模拟中间件流程
	simulateMiddlewareFlow()

	// 2. 测试负载策略
	testLoadBasedPolicy()

	// 3. 验证路由正确性
	testRoutingCorrectness()

	fmt.Println("\nVERIFLOW_DEBUG: All routing fix tests completed!")
	fmt.Println("VERIFLOW_DEBUG: The fix should resolve the 'no available channel' error")
	fmt.Println("VERIFLOW_DEBUG: by converting virtual policy models to real models")
	fmt.Println("VERIFLOW_DEBUG: BEFORE channel lookup, not after.")
} 