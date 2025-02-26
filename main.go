// Package main 是应用程序的入口包
package main

import (
	"os"

	// 导入项目自身的命令行工具包
	"github.com/LyricTian/gin-admin/v10/cmd"
	// 导入命令行应用框架
	"github.c
)om/urfave/cli/v2"

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

// 注解对应解释
// title: 应用程序名称（必填）
// version: 应用程序 API 的版本（必填）
// description: 应用程序的简短描述
// securityDefinitions.apikey: 指定调用 API 的安全方式，也即需要 API 密钥（需要与 HTTP/SSL 一起使用），ApiKeyAuth 指定了一个名称方便标识，可以任取
// in: apikey 的参数之一，指定API 密钥的传递方式（请求头或查询参数）
// name: apikey 的参数之一，标头或参数的名称
// schemes: 指明请求的传输协议（用空格分隔）
// basePath: 运行 API 的基本路径

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
	// os.Args 是一个字符串切片，存储程序运行时传入的所有命令行参数
	// 例如：命令 ./ginadmin start --port 8080
	// os.Args 的值是 
	// os.Args = []string{
	// 	"./ginadmin",  // os.Args[0] - 程序名
	// 	"start",       // os.Args[1] - 子命令
	// 	"--port",      // os.Args[2] - 参数名
	// 	"8080"        // os.Args[3] - 参数值
	// }
	err := app.Run(os.Args)
	// 如果出现错误，立即终止程序
	if err != nil {
		panic(err)
	}
}
