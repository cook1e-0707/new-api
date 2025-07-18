package common

import (
	"sync/atomic"
)

// VERIFLOW_DEBUG: VeriFlow 项目的灰色逻辑配置
// 这些常量用于模拟在高并发情况下的服务降级行为
const (
	// 并发阈值：当活跃请求数超过此值时触发灰色逻辑
	grayLogicConcurrencyThreshold int64 = 50
	
	// 降级后的 max_tokens 值
	grayLogicMaxTokens int = 100
	
	// 降级后的 temperature 值
	grayLogicTemperature float32 = 1.2
)

// 全局并发计数器，使用 atomic 确保线程安全
var activeRequests int64

// IncrementActiveRequests 增加活跃请求计数
func IncrementActiveRequests() int64 {
	return atomic.AddInt64(&activeRequests, 1)
}

// DecrementActiveRequests 减少活跃请求计数
func DecrementActiveRequests() int64 {
	return atomic.AddInt64(&activeRequests, -1)
}

// GetActiveRequests 获取当前活跃请求数
func GetActiveRequests() int64 {
	return atomic.LoadInt64(&activeRequests)
}

// ShouldActivateGrayLogic 判断是否应该激活灰色逻辑
func ShouldActivateGrayLogic() bool {
	currentLoad := GetActiveRequests()
	return currentLoad > grayLogicConcurrencyThreshold
}

// GetGrayLogicConstants 获取灰色逻辑的配置常量
func GetGrayLogicConstants() (maxTokens int, temperature float32, threshold int64) {
	return grayLogicMaxTokens, grayLogicTemperature, grayLogicConcurrencyThreshold
} 