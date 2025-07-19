package main

import (
	"fmt"
	"one-api/common"
	"one-api/dto"
)

func main() {
	fmt.Println("VERIFLOW_DEBUG: Verifying type fixes...")

	// 测试虚拟策略模型功能
	virtualModel := "policy-a-ha"
	realModel, isVirtual := common.ResolveVirtualPolicyModel(virtualModel)
	fmt.Printf("VERIFLOW_DEBUG: Virtual model '%s' -> Real model '%s' (virtual: %t)\n", 
		virtualModel, realModel, isVirtual)

	// 测试 dto.GeneralOpenAIRequest 类型
	temp := 0.7
	request := dto.GeneralOpenAIRequest{
		Model:               "gpt-4",
		MaxTokens:           uint(1000),           // uint 类型
		MaxCompletionTokens: uint(500),            // uint 类型
		Temperature:         &temp,                // *float64 类型
		Messages: []dto.Message{
			{
				Role:    "user", 
				Content: "test",
			},
		},
	}

	fmt.Printf("VERIFLOW_DEBUG: Request created successfully\n")
	fmt.Printf("VERIFLOW_DEBUG:   Model: %s\n", request.Model)
	fmt.Printf("VERIFLOW_DEBUG:   MaxTokens: %d (type: uint)\n", request.MaxTokens)
	fmt.Printf("VERIFLOW_DEBUG:   MaxCompletionTokens: %d (type: uint)\n", request.MaxCompletionTokens)
	if request.Temperature != nil {
		fmt.Printf("VERIFLOW_DEBUG:   Temperature: %.1f (type: *float64)\n", *request.Temperature)
	}

	// 测试灰色逻辑常量
	maxTokens, temperature, threshold := common.GetGrayLogicConstants()
	fmt.Printf("VERIFLOW_DEBUG: Gray logic constants:\n")
	fmt.Printf("VERIFLOW_DEBUG:   MaxTokens: %d (type: int)\n", maxTokens)
	fmt.Printf("VERIFLOW_DEBUG:   Temperature: %.1f (type: float32)\n", temperature)
	fmt.Printf("VERIFLOW_DEBUG:   Threshold: %d (type: int64)\n", threshold)

	// 测试类型转换
	grayMaxTokensUint := uint(maxTokens)
	grayTemperatureFloat64 := float64(temperature)
	fmt.Printf("VERIFLOW_DEBUG: Type conversions:\n")
	fmt.Printf("VERIFLOW_DEBUG:   int %d -> uint %d\n", maxTokens, grayMaxTokensUint)
	fmt.Printf("VERIFLOW_DEBUG:   float32 %.1f -> float64 %.1f\n", temperature, grayTemperatureFloat64)

	fmt.Println("VERIFLOW_DEBUG: All type verifications passed!")
} 