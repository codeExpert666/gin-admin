// mail 包提供了发送电子邮件的功能实现
package mail

import (
	"context"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	// globalSender 全局的SMTP发送器实例
	globalSender *SmtpSender
	// once 确保全局发送器只被初始化一次的同步锁
	once sync.Once
)

// SetSender 设置全局SMTP发送器
// 使用 sync.Once 确保该函数只执行一次，保证线程安全
func SetSender(sender *SmtpSender) {
	once.Do(func() {
		globalSender = sender
	})
}

// Send 使用SMTP客户端发送邮件，支持完整的收件人设置（收件人、抄送、密送）
// 参数说明:
// - ctx: 上下文对象，用于控制请求的生命周期
// - to: 主要收件人列表
// - cc: 抄送人列表
// - bcc: 密送人列表
// - subject: 邮件主题
// - body: 邮件正文
// - file: 可变参数，表示要附加的文件路径列表
func Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	return globalSender.Send(ctx, to, cc, bcc, subject, body, file...)
}

// SendTo 使用SMTP客户端发送邮件的简化版本，只需指定收件人
// 这是一个便捷方法，适用于只需要设置收件人的简单场景
func SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	return globalSender.SendTo(ctx, to, subject, body, file...)
}

// SmtpSender SMTP邮件客户端结构体
// 包含发送邮件所需的所有配置信息
type SmtpSender struct {
	SmtpHost string // SMTP服务器地址
	Port     int    // SMTP服务器端口
	FromName string // 发件人显示名称
	FromMail string // 发件人邮箱地址
	UserName string // SMTP认证用户名
	AuthCode string // SMTP认证密码或授权码
}

// Send 发送邮件的具体实现方法
// 支持设置收件人、抄送、密送，以及添加附件
func (s *SmtpSender) Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	// 创建新的邮件消息，设置Base64编码
	msg := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	// 设置发件人信息
	msg.SetHeader("From", msg.FormatAddress(s.FromMail, s.FromName))
	// 设置收件人、抄送和密送
	msg.SetHeader("To", to...)
	msg.SetHeader("Cc", cc...)
	msg.SetHeader("Bcc", bcc...)
	msg.SetHeader("Subject", subject)
	// 设置邮件内容为HTML格式
	msg.SetBody("text/html;charset=utf-8", body)

	// 添加附件
	for _, v := range file {
		msg.Attach(v)
	}

	// 创建SMTP拨号器并发送邮件
	d := gomail.NewDialer(s.SmtpHost, s.Port, s.UserName, s.AuthCode)
	return d.DialAndSend(msg)
}

// SendTo 带有重试机制的邮件发送方法
// 最多尝试3次发送，每次失败后等待500毫秒
func (s *SmtpSender) SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	var err error
	// 最多重试3次
	for i := 0; i < 3; i++ {
		err = s.Send(ctx, to, nil, nil, subject, body, file...)
		if err != nil {
			// 发送失败后等待500毫秒再重试
			time.Sleep(time.Millisecond * 500)
			continue
		}
		err = nil
		break
	}
	return err
}
