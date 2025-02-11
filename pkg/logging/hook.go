package logging

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// HookExecuter 接口定义了日志钩子执行器需要实现的方法
type HookExecuter interface {
	// Exec 执行日志处理，extra为额外的参数信息，b为日志内容
	Exec(extra map[string]string, b []byte) error
	// Close 关闭执行器
	Close() error
}

// hookOptions 定义了日志钩子的配置选项
type hookOptions struct {
	maxJobs    int               // 最大任务队列长度
	maxWorkers int               // 最大工作协程数
	extra      map[string]string // 额外的参数信息
}

// SetHookMaxJobs 设置最大任务队列长度的选项函数
func SetHookMaxJobs(maxJobs int) HookOption {
	return func(o *hookOptions) {
		o.maxJobs = maxJobs
	}
}

// SetHookMaxWorkers 设置最大工作协程数的选项函数
func SetHookMaxWorkers(maxWorkers int) HookOption {
	return func(o *hookOptions) {
		o.maxWorkers = maxWorkers
	}
}

// SetHookExtra 设置额外参数的选项函数
func SetHookExtra(extra map[string]string) HookOption {
	return func(o *hookOptions) {
		o.extra = extra
	}
}

// HookOption 定义了配置钩子的函数类型
type HookOption func(*hookOptions)

// NewHook 创建一个新的日志钩子实例
// exec: 日志处理执行器
// opt: 可选的配置选项
func NewHook(exec HookExecuter, opt ...HookOption) *Hook {
	// 设置默认配置
	opts := &hookOptions{
		maxJobs:    1024, // 默认队列长度为1024
		maxWorkers: 2,    // 默认2个工作协程
	}

	// 应用自定义配置
	for _, o := range opt {
		o(opts)
	}

	// 创建等待组，用于同步工作协程
	wg := new(sync.WaitGroup)
	wg.Add(opts.maxWorkers)

	// 创建Hook实例
	h := &Hook{
		opts: opts,
		q:    make(chan []byte, opts.maxJobs), // 创建带缓冲的通道
		wg:   wg,
		e:    exec,
	}
	h.dispatch() // 启动工作协程
	return h
}

// Hook 结构体定义了日志钩子的核心数据结构
type Hook struct {
	opts   *hookOptions    // 配置选项
	q      chan []byte     // 日志消息队列
	wg     *sync.WaitGroup // 用于同步的等待组
	e      HookExecuter    // 日志处理执行器
	closed int32           // 钩子是否已关闭的标志（原子操作）
}

// dispatch 启动工作协程处理日志消息
func (h *Hook) dispatch() {
	// 启动指定数量的工作协程
	for i := 0; i < h.opts.maxWorkers; i++ {
		go func() {
			defer func() {
				h.wg.Done()
				// 捕获可能的panic
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic in logger hook:", r)
				}
			}()

			// 持续从队列中获取并处理日志消息
			for data := range h.q {
				err := h.e.Exec(h.opts.extra, data)
				if err != nil {
					fmt.Println("Failed to write entry:", err.Error())
				}
			}
		}()
	}
}

// Write 实现io.Writer接口，用于写入日志数据
func (h *Hook) Write(p []byte) (int, error) {
	// 如果钩子已关闭，直接返回
	if atomic.LoadInt32(&h.closed) == 1 {
		return len(p), nil
	}
	// 如果队列已满，丢弃消息
	if len(h.q) == h.opts.maxJobs {
		fmt.Println("Too many jobs, waiting for queue to be empty, discard")
		return len(p), nil
	}

	// 复制数据并发送到队列
	data := make([]byte, len(p))
	copy(data, p)
	h.q <- data

	return len(p), nil
}

// Flush 等待所有日志处理完成并关闭钩子
func (h *Hook) Flush() {
	// 如果已经关闭，直接返回
	if atomic.LoadInt32(&h.closed) == 1 {
		return
	}
	// 标记为已关闭
	atomic.StoreInt32(&h.closed, 1)
	// 关闭队列通道
	close(h.q)
	// 等待所有工作协程完成
	h.wg.Wait()
	// 关闭执行器
	err := h.e.Close()
	if err != nil {
		fmt.Println("Failed to close logger hook:", err.Error())
	}
}
