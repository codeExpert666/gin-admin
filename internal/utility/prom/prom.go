// Package prom 提供了 Prometheus 监控指标的集成功能
package prom

import (
	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/promx"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
)

// 全局变量定义
var (
	// Ins 是 Prometheus 包装器的实例
	Ins *promx.PrometheusWrapper
	// GinMiddleware 是用于 Gin 框架的 Prometheus 中间件
	GinMiddleware gin.HandlerFunc
)

// Init 初始化 Prometheus 监控
// 该函数配置并启动 Prometheus 监控服务
func Init() {
	// 创建用于存储需要监控的 HTTP 方法的集合
	logMethod := make(map[string]struct{})
	// 创建用于存储需要监控的 API 路径的集合
	logAPI := make(map[string]struct{})

	// 从配置文件中读取需要监控的 HTTP 方法，并存入集合
	for _, m := range config.C.Util.Prometheus.LogMethods {
		logMethod[m] = struct{}{}
	}
	// 从配置文件中读取需要监控的 API 路径，并存入集合
	for _, a := range config.C.Util.Prometheus.LogApis {
		logAPI[a] = struct{}{}
	}

	// 创建并配置 Prometheus 包装器实例
	Ins = promx.NewPrometheusWrapper(&promx.Config{
		Enable:         config.C.Util.Prometheus.Enable,                          // 是否启用 Prometheus
		App:            config.C.General.AppName,                                 // 应用名称
		ListenPort:     config.C.Util.Prometheus.Port,                            // Prometheus 监听端口
		BasicUserName:  config.C.Util.Prometheus.BasicUsername,                   // 基础认证用户名
		BasicPassword:  config.C.Util.Prometheus.BasicPassword,                   // 基础认证密码
		LogApi:         logAPI,                                                   // 需要监控的 API 路径集合
		LogMethod:      logMethod,                                                // 需要监控的 HTTP 方法集合
		Objectives:     map[float64]float64{0.9: 0.01, 0.95: 0.005, 0.99: 0.001}, // 监控指标的分位数配置
		DefaultCollect: config.C.Util.Prometheus.DefaultCollect,                  // 是否收集默认指标
	})

	// 创建 Gin 中间件
	// 该中间件会收集 HTTP 请求的相关指标
	GinMiddleware = promx.NewAdapterGin(Ins).Middleware(config.C.Util.Prometheus.Enable, util.ReqBodyKey)
}
