package controllers

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// MatchKeywords 判断 message 是否包含指定的关键字
func MatchKeywords(message any) bool {
	// 初始化 zap logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	// 获取配置
	keywords := viper.GetStringSlice("keywords.include")
	excludeKeywords := viper.GetStringSlice("keywords.exclude")

	// 将 message 转换为 map[string]interface{} 以便提取字段内容
	messageMap, ok := message.(map[string]interface{})
	if !ok {
		logger.Info("无法将 message 转换为 map")
		return false
	}

	// 遍历 message 中的所有字段并检查其值是否包含/排除关键字
	for _, value := range messageMap {
		valueStr := fmt.Sprintf("%v", value)
		lowerText := strings.ToLower(valueStr)

		// 先检查是否包含排除关键字
		for _, keyword := range excludeKeywords {
			if strings.Contains(lowerText, strings.ToLower(keyword)) {
				return false
			}
		}

		// 再检查是否包含包含关键字
		for _, keyword := range keywords {
			if strings.Contains(lowerText, strings.ToLower(keyword)) {
				return true
			}
		}
	}

	return false
}
