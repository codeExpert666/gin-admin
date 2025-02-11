// Package cmd 提供了应用程序的命令行接口实现
package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// VersionCmd 创建并返回一个用于显示应用程序版本号的命令
// 参数：
//   - v: 版本号字符串
// 返回：
//   - *cli.Command: 一个配置好的 CLI 命令对象
func VersionCmd(v string) *cli.Command {
	return &cli.Command{
		// 命令名称
		Name: "version",
		// 命令用途说明
		Usage: "Show version",
		// 命令执行函数
		Action: func(_ *cli.Context) error {
			// 打印版本号并返回
			fmt.Println(v)
			return nil
		},
	}
}
