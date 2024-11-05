package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"skywalking-filter-webhook/controllers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 初始化zap logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	// 加载配置，获取不到就退出
	viper.SetConfigName("config") // 配置文件名 (不带扩展名)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 配置文件路径

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("读取配置文件失败", zap.Error(err))
	}

	// 创建Gin路由
	r := gin.Default()

	// 定义 POST 路由
	r.POST("/skywalking", func(c *gin.Context) {

		// 读取原始请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("读取原始请求体失败", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
			return
		}

		// 打印原始报文
		logger.Info("原始报文，准备发送到飞书", zap.String("报文内容", string(bodyBytes)))

		// 将请求体重新放回 c.Request.Body，以便后续解析
		c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyBytes)))

		// 使用 map 来解析任意 JSON 报文
		var msg map[string]any
		if err := c.BindJSON(&msg); err != nil {
			logger.Error("原始报文解析失败", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// 飞书的 Webhook URL
		url := viper.GetString("webhook.skywalking.url")
		if url == "" {
			logger.Error("配置文件未设置 webhook.skywalking.url")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "配置文件未设置 webhook.skywalking.url"})
			return
		}

		// 判断报文是否匹配关键字（不区分大小写）
		if controllers.MatchKeywords(msg) {
			logger.Info("消息匹配，发送到飞书")
			err := controllers.SendToFeishu(msg, url)
			if err != nil {
				logger.Error("发送到飞书失败", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"message": "消息匹配，发送到飞书失败", "error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": "0", "message": "消息匹配，发送到飞书"})
		} else {
			logger.Info("消息不匹配，不发送到飞书")
			c.JSON(http.StatusOK, gin.H{"code": "9001", "message": "消息不匹配，不发送到飞书"})
		}
	})

	// 运行服务器
	port := viper.GetInt("server.port")
	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
