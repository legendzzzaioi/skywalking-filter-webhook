package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// FeishuResponse 飞书返回的响应结构体
type FeishuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SendToFeishu 发送消息到飞书
func SendToFeishu(message any, url string) error {
	// 初始化zap logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	// 构造 JSON 报文
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %v", err)
	}

	// 发送 POST 请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送到飞书失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("飞书返回了错误 HTTP 状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取飞书响应体失败: %v", err)
	}

	// 解析响应 JSON
	var feishuResp FeishuResponse
	if err := json.Unmarshal(body, &feishuResp); err != nil {
		return fmt.Errorf("解析飞书响应 JSON 失败: %v", err)
	}

	// 检查返回的 code 字段
	if feishuResp.Code != 0 {
		return fmt.Errorf("飞书返回错误: code=%d, msg=%s, 响应内容=%s", feishuResp.Code, feishuResp.Msg, string(body))
	}

	// 打印成功信息
	logger.Info("发送到飞书成功",
		zap.String("响应内容", string(body)))
	return nil

}
