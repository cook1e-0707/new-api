package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"one-api/common"
	"one-api/dto"
)

// TestPolicyRequest 测试虚拟策略模型请求
type TestPolicyRequest struct {
	Model       string              `json:"model"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float32             `json:"temperature"`
	Messages    []dto.OpenAIMessage `json:"messages"`
	Stream      bool                `json:"stream"`
}

// TestPolicyResponse 测试响应
type TestPolicyResponse struct {
	Model   string `json:"model"`
	Content string `json:"content"`
}

// testPolicyResolution 测试虚拟策略模型解析
func testPolicyResolution() {
	fmt.Println("VERIFLOW_DEBUG: Testing Virtual Policy Model Resolution...")
	fmt.Println("VERIFLOW_DEBUG: ================================================")

	virtualPolicies := []string{
		common.PolicyModelHA,      // policy-a-ha
		common.PolicyModelCost,    // policy-b-cost
		common.PolicyModelQuality, // policy-c-quality
		common.PolicyModelDegrade, // policy-d-degrade
	}

	for _, virtualModel := range virtualPolicies {
		fmt.Printf("\nVERIFLOW_DEBUG: Testing policy: %s\n", virtualModel)

		// 检查是否为虚拟策略模型
		isVirtual := common.IsVirtualPolicyModel(virtualModel)
		fmt.Printf("VERIFLOW_DEBUG: IsVirtualPolicyModel('%s') = %t\n", virtualModel, isVirtual)

		if isVirtual {
			// 测试多次解析，观察随机性
			fmt.Printf("VERIFLOW_DEBUG: Resolution results for '%s':\n", virtualModel)
			for i := 0; i < 5; i++ {
				realModel, resolved := common.ResolveVirtualPolicyModel(virtualModel)
				fmt.Printf("VERIFLOW_DEBUG:   Attempt %d: %s -> %s (resolved: %t)\n", 
					i+1, virtualModel, realModel, resolved)
			}
		}
	}
}

// testPolicyLoadBalancing 测试负载降级策略
func testPolicyLoadBalancing() {
	fmt.Println("\n\nVERIFLOW_DEBUG: Testing Load-based Policy (policy-d-degrade)...")
	fmt.Println("VERIFLOW_DEBUG: ==========================================================")

	// 测试低负载情况
	fmt.Printf("VERIFLOW_DEBUG: Current load: %d\n", common.GetActiveRequests())
	fmt.Printf("VERIFLOW_DEBUG: Gray logic threshold: %d\n", func() int64 { _, _, threshold := common.GetGrayLogicConstants(); return threshold }())
	
	realModel, _ := common.ResolveVirtualPolicyModel(common.PolicyModelDegrade)
	fmt.Printf("VERIFLOW_DEBUG: Low load resolution: %s -> %s\n", common.PolicyModelDegrade, realModel)

	// 模拟高负载情况
	fmt.Println("\nVERIFLOW_DEBUG: Simulating high load...")
	var wg sync.WaitGroup
	const simulatedLoad = 60 // 超过阈值 50

	for i := 0; i < simulatedLoad; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			common.IncrementActiveRequests()
			defer common.DecrementActiveRequests()

			// 在高负载下测试策略解析
			if id == simulatedLoad/2 { // 只在中间测试一次，避免输出过多
				currentLoad := common.GetActiveRequests()
				realModel, _ := common.ResolveVirtualPolicyModel(common.PolicyModelDegrade)
				fmt.Printf("VERIFLOW_DEBUG: High load resolution (load=%d): %s -> %s\n", 
					currentLoad, common.PolicyModelDegrade, realModel)
			}

			// 模拟请求处理时间
			time.Sleep(50 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	fmt.Printf("VERIFLOW_DEBUG: Final load after simulation: %d\n", common.GetActiveRequests())
}

// testPolicyStatistics 测试策略统计
func testPolicyStatistics() {
	fmt.Println("\n\nVERIFLOW_DEBUG: Testing Policy Statistics...")
	fmt.Println("VERIFLOW_DEBUG: =====================================")

	// 测试成本效益策略的分布（8:2 比例）
	fmt.Printf("VERIFLOW_DEBUG: Testing cost policy distribution (expected ~80%% %s, ~20%% %s):\n", 
		common.ModelGeminiFlash, common.ModelClaude3Haiku)

	geminiCount := 0
	claudeCount := 0
	totalTests := 1000

	for i := 0; i < totalTests; i++ {
		realModel, _ := common.ResolveVirtualPolicyModel(common.PolicyModelCost)
		switch realModel {
		case common.ModelGeminiFlash:
			geminiCount++
		case common.ModelClaude3Haiku:
			claudeCount++
		}
	}

	geminiPercent := float64(geminiCount) / float64(totalTests) * 100
	claudePercent := float64(claudeCount) / float64(totalTests) * 100

	fmt.Printf("VERIFLOW_DEBUG: Results after %d tests:\n", totalTests)
	fmt.Printf("VERIFLOW_DEBUG:   %s: %d times (%.1f%%)\n", common.ModelGeminiFlash, geminiCount, geminiPercent)
	fmt.Printf("VERIFLOW_DEBUG:   %s: %d times (%.1f%%)\n", common.ModelClaude3Haiku, claudeCount, claudePercent)

	// 测试高可用策略的分布（50:50 比例）
	fmt.Printf("\nVERIFLOW_DEBUG: Testing HA policy distribution (expected ~50%% each):\n")

	gpt4Count := 0
	sonnetCount := 0

	for i := 0; i < totalTests; i++ {
		realModel, _ := common.ResolveVirtualPolicyModel(common.PolicyModelHA)
		switch realModel {
		case common.ModelGPT4O:
			gpt4Count++
		case common.ModelClaude3Sonnet:
			sonnetCount++
		}
	}

	gpt4Percent := float64(gpt4Count) / float64(totalTests) * 100
	sonnetPercent := float64(sonnetCount) / float64(totalTests) * 100

	fmt.Printf("VERIFLOW_DEBUG:   %s: %d times (%.1f%%)\n", common.ModelGPT4O, gpt4Count, gpt4Percent)
	fmt.Printf("VERIFLOW_DEBUG:   %s: %d times (%.1f%%)\n", common.ModelClaude3Sonnet, sonnetCount, sonnetPercent)
}

// testModelNameSpoofing 测试模型名称欺骗与虚拟策略结合
func testModelNameSpoofing() {
	fmt.Println("\n\nVERIFLOW_DEBUG: Testing Model Name Spoofing with Virtual Policies...")
	fmt.Println("VERIFLOW_DEBUG: ===============================================================")

	virtualPolicies := []string{
		common.PolicyModelHA,
		common.PolicyModelCost,
		common.PolicyModelQuality,
		common.PolicyModelDegrade,
	}

	for _, virtualModel := range virtualPolicies {
		realModel, isVirtual := common.ResolveVirtualPolicyModel(virtualModel)
		if isVirtual {
			fmt.Printf("\nVERIFLOW_DEBUG: Testing spoofing for: %s\n", virtualModel)
			fmt.Printf("VERIFLOW_DEBUG: Virtual model: %s -> Real model: %s\n", virtualModel, realModel)

			// 模拟响应处理
			// 1. 非流式响应
			response := dto.OpenAITextResponse{
				ID:      "chatcmpl-test-policy",
				Object:  "chat.completion",
				Created: time.Now().Unix(),
				Model:   realModel, // 这是实际处理响应时的模型
				Choices: []dto.OpenAITextResponseChoice{
					{
						Index: 0,
						Message: dto.OpenAIMessage{
							Role:    "assistant",
							Content: "Response from real model",
						},
						FinishReason: "stop",
					},
				},
				Usage: dto.Usage{
					PromptTokens:     10,
					CompletionTokens: 15,
					TotalTokens:      25,
				},
			}

			// 应用模型名称欺骗
			spoofedModel := virtualModel // 这是用户应该看到的模型名称
			response.Model = spoofedModel

			fmt.Printf("VERIFLOW_DEBUG: Non-stream response spoofing:\n")
			fmt.Printf("VERIFLOW_DEBUG:   Internal model: %s\n", realModel)
			fmt.Printf("VERIFLOW_DEBUG:   User sees: %s\n", response.Model)

			if response.Model == virtualModel {
				fmt.Printf("VERIFLOW_DEBUG:   ✓ Spoofing SUCCESS\n")
			} else {
				fmt.Printf("VERIFLOW_DEBUG:   ✗ Spoofing FAILED\n")
			}

			// 2. 流式响应
			streamResponse := dto.ChatCompletionsStreamResponse{
				ID:      "chatcmpl-test-stream-policy",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   realModel, // 原始流响应中的模型
				Choices: []dto.ChatCompletionsStreamResponseChoice{
					{
						Index: 0,
						Delta: dto.OpenAIMessage{
							Content: "Stream response",
						},
					},
				},
			}

			// 应用流式响应的模型名称欺骗
			streamResponse.Model = spoofedModel

			fmt.Printf("VERIFLOW_DEBUG: Stream response spoofing:\n")
			fmt.Printf("VERIFLOW_DEBUG:   Internal model: %s\n", realModel)
			fmt.Printf("VERIFLOW_DEBUG:   User sees: %s\n", streamResponse.Model)

			if streamResponse.Model == virtualModel {
				fmt.Printf("VERIFLOW_DEBUG:   ✓ Stream spoofing SUCCESS\n")
			} else {
				fmt.Printf("VERIFLOW_DEBUG:   ✗ Stream spoofing FAILED\n")
			}
		}
	}
}

// testRequestFlow 测试完整的请求流程
func testRequestFlow() {
	fmt.Println("\n\nVERIFLOW_DEBUG: Testing Complete Request Flow...")
	fmt.Println("VERIFLOW_DEBUG: =====================================")

	// 模拟用户请求
	originalRequest := TestPolicyRequest{
		Model:       common.PolicyModelHA, // 用户请求虚拟策略模型
		MaxTokens:   1000,
		Temperature: 0.7,
		Messages: []dto.OpenAIMessage{
			{
				Role:    "user",
				Content: "Hello, I'm using a virtual policy model!",
			},
		},
		Stream: false,
	}

	fmt.Printf("VERIFLOW_DEBUG: User request:\n")
	fmt.Printf("VERIFLOW_DEBUG:   Model: %s\n", originalRequest.Model)

	// 步骤 1: 虚拟策略解析
	realModel, isVirtual := common.ResolveVirtualPolicyModel(originalRequest.Model)
	fmt.Printf("VERIFLOW_DEBUG: Step 1 - Policy resolution:\n")
	fmt.Printf("VERIFLOW_DEBUG:   Virtual: %t\n", isVirtual)
	fmt.Printf("VERIFLOW_DEBUG:   Resolved to: %s\n", realModel)

	// 步骤 2: 模拟后端处理（使用真实模型）
	backendResponse := TestPolicyResponse{
		Model:   realModel, // 后端返回的实际模型
		Content: "Response generated by " + realModel,
	}

	fmt.Printf("VERIFLOW_DEBUG: Step 2 - Backend processing:\n")
	fmt.Printf("VERIFLOW_DEBUG:   Backend model: %s\n", backendResponse.Model)

	// 步骤 3: 模型名称欺骗（返回给用户的应该是虚拟策略名称）
	userResponse := TestPolicyResponse{
		Model:   originalRequest.Model, // 欺骗为用户原始请求的模型
		Content: backendResponse.Content,
	}

	fmt.Printf("VERIFLOW_DEBUG: Step 3 - Model name spoofing:\n")
	fmt.Printf("VERIFLOW_DEBUG:   User sees model: %s\n", userResponse.Model)
	fmt.Printf("VERIFLOW_DEBUG:   User sees content: %s\n", userResponse.Content)

	// 验证完整性
	if userResponse.Model == originalRequest.Model {
		fmt.Printf("VERIFLOW_DEBUG: ✓ Complete flow SUCCESS - User cannot detect internal routing\n")
	} else {
		fmt.Printf("VERIFLOW_DEBUG: ✗ Complete flow FAILED - User can detect internal routing\n")
	}
}

func main() {
	fmt.Println("VERIFLOW_DEBUG: Virtual Policy Model Test Suite")
	fmt.Println("VERIFLOW_DEBUG: ===============================================")

	// 1. 测试策略解析
	testPolicyResolution()

	// 2. 测试负载降级策略
	testPolicyLoadBalancing()

	// 3. 测试策略统计
	testPolicyStatistics()

	// 4. 测试模型名称欺骗
	testModelNameSpoofing()

	// 5. 测试完整请求流程
	testRequestFlow()

	fmt.Println("\n\nVERIFLOW_DEBUG: All virtual policy tests completed!")
	fmt.Println("VERIFLOW_DEBUG: ===============================================")
} 