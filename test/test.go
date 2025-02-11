// package test 定义测试包,用于存放测试相关的代码
package test

import (
	"context"
	"net/http"
	"os"
	"testing"

	// 导入项目内部的配置包
	"github.com/LyricTian/gin-admin/v10/internal/config"
	// 导入依赖注入相关的包
	"github.com/LyricTian/gin-admin/v10/internal/wirex"
	// 导入HTTP测试框架
	"github.com/gavv/httpexpect/v2"
	// 导入Gin Web框架
	"github.com/gin-gonic/gin"
)

const (
	// baseAPI 定义API的基础路径
	baseAPI = "/api/v1"
)

var (
	// app 全局变量,存储Gin的引擎实例
	app *gin.Engine
)

// init 函数在包被导入时自动执行,用于初始化测试环境
func init() {
	// 加载配置文件,空字符串表示使用默认配置文件路径
	config.MustLoad("")

	// 删除已存在的数据库文件,确保测试环境干净
	_ = os.RemoveAll(config.C.Storage.DB.DSN)

	// 创建一个新的上下文
	ctx := context.Background()

	// 使用wire框架构建依赖注入器
	injector, _, err := wirex.BuildInjector(ctx)
	if err != nil {
		panic(err)
	}

	// 初始化注入器中的所有组件
	if err := injector.M.Init(ctx); err != nil {
		panic(err)
	}

	// 创建新的Gin引擎实例
	app = gin.New()
	// 注册所有路由
	err = injector.M.RegisterRouters(ctx, app)
	if err != nil {
		panic(err)
	}
}

// tester 函数用于创建HTTP测试客户端
// 参数 t 是测试对象,用于报告测试结果
// 返回一个配置好的httpexpect测试实例
func tester(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		// 配置HTTP客户端
		Client: &http.Client{
			// 设置传输层,使用自定义的Binder将请求直接发送到Gin应用
			Transport: httpexpect.NewBinder(app),
			// 创建新的Cookie管理器
			Jar: httpexpect.NewCookieJar(),
		},
		// 设置测试报告器
		Reporter: httpexpect.NewAssertReporter(t),
		// 配置打印器,用于调试输出
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}
