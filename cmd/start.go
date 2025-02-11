// Package cmd 提供命令行工具相关的功能实现
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LyricTian/gin-admin/v10/internal/bootstrap"
	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/urfave/cli/v2"
)

// StartCmd 定义了启动服务器的命令行指令
// 该命令支持以下功能：
// 1. 指定工作目录
// 2. 指定配置文件
// 3. 指定静态文件目录
// 4. 支持以守护进程方式运行
func StartCmd() *cli.Command {
	return &cli.Command{
		Name:  "start",    // 命令名称
		Usage: "Start server",  // 命令用途说明
		Flags: []cli.Flag{  // 命令行参数定义
			&cli.StringFlag{
				Name:        "workdir",      // 参数名
				Aliases:     []string{"d"},  // 参数别名
				Usage:       "Working directory",  // 参数说明
				DefaultText: "configs",      // 默认值显示文本
				Value:       "configs",      // 实际默认值
			},
			&cli.StringFlag{
				Name:        "config",       // 配置参数名
				Aliases:     []string{"c"},  // 配置参数别名
				Usage:       "Runtime configuration files or directory (relative to workdir, multiple separated by commas)",  // 配置文件说明
				DefaultText: "dev",          // 默认配置环境显示
				Value:       "dev",          // 默认配置环境值
			},
			&cli.StringFlag{
				Name:    "static",           // 静态文件目录参数名
				Aliases: []string{"s"},      // 静态文件目录参数别名
				Usage:   "Static files directory",  // 静态文件目录说明
			},
			&cli.BoolFlag{
				Name:  "daemon",             // 守护进程参数名
				Usage: "Run as a daemon",    // 守护进程运行说明
			},
		},
		// Action 定义了命令的具体执行逻辑
		Action: func(c *cli.Context) error {
			// 获取命令行参数
			workDir := c.String("workdir")    // 获取工作目录
			staticDir := c.String("static")   // 获取静态文件目录
			configs := c.String("config")     // 获取配置文件

			// 如果指定了守护进程模式
			if c.Bool("daemon") {
				// 获取当前可执行文件的绝对路径
				bin, err := filepath.Abs(os.Args[0])
				if err != nil {
					fmt.Printf("failed to get absolute path for command: %s \n", err.Error())
					return err
				}

				// 构建守护进程的启动参数
				args := []string{"start"}
				args = append(args, "-d", workDir)      // 添加工作目录参数
				args = append(args, "-c", configs)      // 添加配置文件参数
				args = append(args, "-s", staticDir)    // 添加静态文件目录参数
				fmt.Printf("execute command: %s %s \n", bin, strings.Join(args, " "))
				command := exec.Command(bin, args...)     // 创建守护进程命令

				// 将标准输出和错误输出重定向到日志文件
				stdLogFile := fmt.Sprintf("%s.log", c.App.Name)  // 生成日志文件名
				// 打开日志文件，如果不存在则创建，追加写入模式
				file, err := os.OpenFile(stdLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
				if err != nil {
					fmt.Printf("failed to open log file: %s \n", err.Error())
					return err
				}
				defer file.Close()  // 确保文件最终被关闭

				// 设置守护进程的输出重定向
				command.Stdout = file
				command.Stderr = file

				// 启动守护进程
				err = command.Start()
				if err != nil {
					fmt.Printf("failed to start daemon thread: %s \n", err.Error())
					return err
				}

				// 不等待命令完成
				// 主进程将退出，允许守护进程独立运行
				fmt.Printf("Service %s daemon thread started successfully\n", config.C.General.AppName)

				// 记录守护进程的PID
				pid := command.Process.Pid
				// 将PID写入锁文件
				_ = os.WriteFile(fmt.Sprintf("%s.lock", c.App.Name), []byte(fmt.Sprintf("%d", pid)), 0666)
				fmt.Printf("service %s daemon thread started with pid %d \n", config.C.General.AppName, pid)
				os.Exit(0)  // 主进程退出
			}

			// 非守护进程模式：直接启动服务
			err := bootstrap.Run(context.Background(), bootstrap.RunConfig{
				WorkDir:   workDir,    // 工作目录
				Configs:   configs,    // 配置文件
				StaticDir: staticDir,  // 静态文件目录
			})
			if err != nil {
				panic(err)  // 如果启动失败，直接panic
			}
			return nil
		},
	}
}
