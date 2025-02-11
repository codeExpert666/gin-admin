package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"go.uber.org/zap"
)

// Run 函数设置信号处理程序并执行处理函数,直到接收到终止信号为止
// 参数:
//   - ctx: 上下文对象,用于传递取消信号
//   - handler: 处理函数,接收上下文对象,返回清理函数和错误
//
// 返回:
//   - error: 如果发生错误则返回,否则返回 nil
func Run(ctx context.Context, handler func(ctx context.Context) (func(), error)) error {
	// 初始化状态值为1,表示正常退出
	state := 1
	// 创建一个缓冲区大小为1的信号通道
	sc := make(chan os.Signal, 1)
	// 监听系统信号:SIGHUP(终端断开)、SIGINT(中断)、SIGTERM(终止)、SIGQUIT(退出)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 执行处理函数,获取清理函数和可能的错误
	cleanFn, err := handler(ctx)
	if err != nil {
		return err
	}

EXIT:
	// 无限循环等待信号
	for {
		// 从信号通道接收信号
		sig := <-sc
		// 记录接收到的信号
		logging.Context(ctx).Info("Received signal", zap.String("signal", sig.String()))

		// 根据不同的信号类型进行处理
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// 收到退出信号时,将状态设为0,表示异常退出
			state = 0
			break EXIT
		case syscall.SIGHUP:
			// 忽略SIGHUP信号
		default:
			// 其他信号直接退出循环
			break EXIT
		}
	}

	// 执行清理函数
	cleanFn()
	// 记录服务器退出日志
	logging.Context(ctx).Info("Server exit, bye...")
	// 等待100毫秒,确保日志写入
	time.Sleep(time.Millisecond * 100)
	// 使用状态码退出程序
	os.Exit(state)
	return nil
}
