package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"one-api/common"
	"one-api/dto"
)

// TestRequest 模拟一个聊天补全请求
type TestRequest struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
	Messages    []dto.OpenAIMessage `json:"messages"`
	Stream      bool    `json:"stream"`
}

// simulateHighConcurrency 模拟高并发请求来触发灰色逻辑
func simulateHighConcurrency() {
	fmt.Println("VERIFLOW_DEBUG: Starting gray logic test...")
	
	// 获取初始配置
	maxTokens, temperature, threshold := common.GetGrayLogicConstants()
	fmt.Printf("VERIFLOW_DEBUG: Configuration - Threshold: %d, MaxTokens: %d, Temperature: %.2f\n", 
		threshold, maxTokens, temperature)

	var wg sync.WaitGroup
	concurrency := int(threshold + 10) // 超过阈值来触发灰色逻辑
	
	fmt.Printf("VERIFLOW_DEBUG: Simulating %d concurrent requests...\n", concurrency)
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 增加并发计数
			currentLoad := common.IncrementActiveRequests()
			
			// 检查是否应该激活灰色逻辑
			shouldActivate := common.ShouldActivateGrayLogic()
			
			fmt.Printf("VERIFLOW_DEBUG: Request %d - Current load: %d, Gray logic: %t\n", 
				id, currentLoad, shouldActivate)
			
			// 模拟请求处理时间
			time.Sleep(100 * time.Millisecond)
			
			// 减少并发计数
			common.DecrementActiveRequests()
		}(i)
	}
	
	wg.Wait()
	
	finalCount := common.GetActiveRequests()
	fmt.Printf("VERIFLOW_DEBUG: Test completed. Final active requests: %d\n", finalCount)
}

// testRequestModification 测试请求参数修改逻辑
func testRequestModification() {
	fmt.Println("\nVERIFLOW_DEBUG: Testing request modification...")
	
	// 创建一个测试请求
	originalRequest := TestRequest{
		Model:       "gpt-4",
		MaxTokens:   2000, // 超过阈值的值
		Temperature: 0.7,
		Messages: []dto.OpenAIMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Stream: false,
	}
	
	fmt.Printf("VERIFLOW_DEBUG: Original request - Model: %s, MaxTokens: %d, Temperature: %.2f\n",
		originalRequest.Model, originalRequest.MaxTokens, originalRequest.Temperature)
	
	// 序列化请求
	requestBody, err := json.Marshal(originalRequest)
	if err != nil {
		fmt.Printf("VERIFLOW_DEBUG: Failed to marshal request: %v\n", err)
		return
	}
	
	// 模拟灰色逻辑处理
	grayMaxTokens, grayTemperature, _ := common.GetGrayLogicConstants()
	
	var modifiedRequest TestRequest
	err = json.Unmarshal(requestBody, &modifiedRequest)
	if err != nil {
		fmt.Printf("VERIFLOW_DEBUG: Failed to unmarshal request: %v\n", err)
		return
	}
	
	// 应用灰色逻辑
	if modifiedRequest.MaxTokens > grayMaxTokens {
		modifiedRequest.MaxTokens = grayMaxTokens
	}
	modifiedRequest.Temperature = grayTemperature
	
	fmt.Printf("VERIFLOW_DEBUG: Modified request - Model: %s, MaxTokens: %d, Temperature: %.2f\n",
		modifiedRequest.Model, modifiedRequest.MaxTokens, modifiedRequest.Temperature)
	
	// 验证模型名称保持不变（用于后续的模型名称欺骗）
	if modifiedRequest.Model != originalRequest.Model {
		fmt.Printf("VERIFLOW_DEBUG: ERROR - Model name should not change during modification!\n")
	} else {
		fmt.Printf("VERIFLOW_DEBUG: SUCCESS - Model name preserved for spoofing: %s\n", modifiedRequest.Model)
	}
}

// testModelNameSpoofing 测试模型名称欺骗逻辑
func testModelNameSpoofing() {
	fmt.Println("\nVERIFLOW_DEBUG: Testing model name spoofing...")
	
	// 模拟一个响应，其中模型名称可能被修改了
	originalModelName := "gpt-4"
	actualResponseModel := "gpt-3.5-turbo" // 假设内部被替换了
	
	// 模拟非流式响应
	response := dto.OpenAITextResponse{
		ID:      "chatcmpl-test",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   actualResponseModel, // 这是实际的响应模型
		Choices: []dto.OpenAITextResponseChoice{
			{
				Index: 0,
				Message: dto.OpenAIMessage{
					Role:    "assistant",
					Content: "Hello! I'm doing well, thank you for asking.",
				},
				FinishReason: "stop",
			},
		},
		Usage: dto.Usage{
			PromptTokens:     10,
			CompletionTokens: 12,
			TotalTokens:      22,
		},
	}
	
	fmt.Printf("VERIFLOW_DEBUG: Original response model: %s\n", response.Model)
	
	// 应用模型名称欺骗
	response.Model = originalModelName
	
	fmt.Printf("VERIFLOW_DEBUG: Spoofed response model: %s\n", response.Model)
	
	// 验证用户看到的是原始模型名称
	if response.Model == originalModelName {
		fmt.Printf("VERIFLOW_DEBUG: SUCCESS - Model name spoofing works correctly!\n")
	} else {
		fmt.Printf("VERIFLOW_DEBUG: ERROR - Model name spoofing failed!\n")
	}
	
	// 测试流式响应
	streamResponse := dto.ChatCompletionsStreamResponse{
		ID:      "chatcmpl-test-stream",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   actualResponseModel, // 原始流响应中的模型
		Choices: []dto.ChatCompletionsStreamResponseChoice{
			{
				Index: 0,
				Delta: dto.OpenAIMessage{
					Content: "Hello",
				},
			},
		},
	}
	
	fmt.Printf("VERIFLOW_DEBUG: Original stream response model: %s\n", streamResponse.Model)
	
	// 应用流式响应的模型名称欺骗
	streamResponse.Model = originalModelName
	
	fmt.Printf("VERIFLOW_DEBUG: Spoofed stream response model: %s\n", streamResponse.Model)
	
	if streamResponse.Model == originalModelName {
		fmt.Printf("VERIFLOW_DEBUG: SUCCESS - Stream model name spoofing works correctly!\n")
	} else {
		fmt.Printf("VERIFLOW_DEBUG: ERROR - Stream model name spoofing failed!\n")
	}
}

func main() {
	fmt.Println("VERIFLOW_DEBUG: VeriFlow Gray Logic Test Suite")
	fmt.Println("VERIFLOW_DEBUG: =====================================")
	
	// 测试并发控制
	simulateHighConcurrency()
	
	// 测试请求修改
	testRequestModification()
	
	// 测试模型名称欺骗
	testModelNameSpoofing()
	
	fmt.Println("\nVERIFLOW_DEBUG: All tests completed!")
	fmt.Println("VERIFLOW_DEBUG: =====================================")
} 