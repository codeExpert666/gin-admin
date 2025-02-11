// Package main 是应用程序的入口包
package main

import (
	"os"

	// 导入项目自身的命令行工具包
	"github.com/LyricTian/gin-admin/v10/cmd"
	// 导入命令行应用框架
	"github.com/urfave/cli/v2"
)

// VERSION 定义应用程序版本号
// 可以在编译时通过 -ldflags "-X main.VERSION=x.x.x" 指定版本号
var VERSION = "v10.1.0"

// Swagger API 文档相关注解
// @title ginadmin
// @version v10.1.0
// @description A lightweight, flexible, elegant and full-featured RBAC scaffolding based on GIN + GORM 2.0 + Casbin 2.0 + Wire DI.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @schemes http https
// @basePath /

// main 函数是应用程序的入口点
func main() {
	// 创建一个新的命令行应用实例
	app := cli.NewApp()
	
	// 设置应用基本信息
	app.Name = "ginadmin"
	app.Version = VERSION
	app.Usage = "A lightweight, flexible, elegant and full-featured RBAC scaffolding based on GIN + GORM 2.0 + Casbin 2.0 + Wire DI."
	
	// 注册子命令
	app.Commands = []*cli.Command{
		cmd.StartCmd(), // 启动服务命令
		cmd.StopCmd(),  // 停止服务命令
		cmd.VersionCmd(VERSION), // 版本信息命令
	}

	// 运行应用，传入命令行参数
	err := app.Run(os.Args)
	// 如果出现错误，立即终止程序
	if err != nil {
		panic(err)
	}
}
