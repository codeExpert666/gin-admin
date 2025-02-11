// Package cmd 提供了应用程序的命令行接口实现
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2" // 使用 urfave/cli 包来实现命令行接口
)

// StopCmd 创建并返回一个用于停止服务器的命令行命令
// 该命令的执行流程:
// 1. 读取锁文件获取进程 PID
// 2. 使用 kill 命令终止对应的进程
// 3. 删除锁文件
// 4. 输出停止服务成功的消息
func StopCmd() *cli.Command {
	return &cli.Command{
		Name:  "stop",    // 命令名称
		Usage: "stop server", // 命令用途说明
		Action: func(c *cli.Context) error {
			// 获取应用程序名称
			appName := c.App.Name
			// 构造锁文件名称 (格式: 应用名称.lock)
			lockFile := fmt.Sprintf("%s.lock", appName)
			// 读取锁文件内容获取进程 PID
			pid, err := os.ReadFile(lockFile)
			if err != nil {
				return err // 如果读取失败则返回错误
			}

			// 创建 kill 命令用于终止进程
			command := exec.Command("kill", string(pid))
			// 执行 kill 命令
			err = command.Start()
			if err != nil {
				return err // 如果执行失败则返回错误
			}

			// 删除锁文件
			err = os.Remove(lockFile)
			if err != nil {
				// 如果删除失败则返回格式化的错误信息
				return fmt.Errorf("can't remove %s.lock. %s", appName, err.Error())
			}

			// 输出服务停止成功的消息
			fmt.Printf("service %s stopped \n", appName)
			return nil // 返回 nil 表示命令执行成功
		},
	}
}
