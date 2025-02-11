// Package promx 提供了一个 Prometheus 监控指标的包装器
// 用于在 Web 应用中方便地集成 Prometheus 监控,支持请求数、延迟、字节数等多种指标的收集
package promx

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config 定义了 Prometheus 监控的配置选项
type Config struct {
	Enable         bool                // 是否启用 Prometheus 监控
	App            string              // 应用名称,用于指标标签
	ListenPort     int                 // Prometheus 指标暴露端口
	BasicUserName  string              // 基础认证用户名
	BasicPassword  string              // 基础认证密码
	LogApi         map[string]struct{} // 需要记录的 API 路径白名单
	LogMethod      map[string]struct{} // 需要记录的 HTTP 方法白名单
	Buckets        []float64           // 直方图的桶值配置
	Objectives     map[float64]float64 // 摘要的分位数配置
	DefaultCollect bool                // 是否收集默认的 Go 运行时指标
}

// PrometheusWrapper 封装了所有 Prometheus 相关的监控指标
type PrometheusWrapper struct {
	c                                  Config                   // 配置信息
	reg                                *prometheus.Registry     // Prometheus 注册器
	gaugeState                         *prometheus.GaugeVec     // 状态度量指标
	histogramLatency                   *prometheus.HistogramVec // 延迟时间直方图
	summaryLatency                     *prometheus.SummaryVec   // 延迟时间摘要
	counterRequests, counterSendBytes  *prometheus.CounterVec   // 请求计数器和发送字节计数器
	counterRcvdBytes, counterException *prometheus.CounterVec   // 接收字节计数器和异常计数器
	counterEvent, counterSiteEvent     *prometheus.CounterVec   // 事件计数器和站点事件计数器
}

// init 初始化所有 Prometheus 监控指标
func (p *PrometheusWrapper) init() {
	// 初始化请求计数器
	p.counterRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_requests",
			Help: "number of module requests",
		},
		[]string{"app", "module", "api", "method", "code"},
	)
	p.reg.MustRegister(p.counterRequests)

	// 初始化发送字节计数器
	p.counterSendBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_send_bytes",
			Help: "number of module send bytes",
		},
		[]string{"app", "module", "api", "method", "code"},
	)
	p.reg.MustRegister(p.counterSendBytes)

	// 初始化接收字节计数器
	p.counterRcvdBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_rcvd_bytes",
			Help: "number of module receive bytes",
		},
		[]string{"app", "module", "api", "method", "code"},
	)
	p.reg.MustRegister(p.counterRcvdBytes)

	// 初始化延迟时间直方图
	p.histogramLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "histogram_latency",
			Help:    "histogram of module latency",
			Buckets: p.c.Buckets,
		},
		[]string{"app", "module", "api", "method"},
	)
	p.reg.MustRegister(p.histogramLatency)

	// 初始化延迟时间摘要
	p.summaryLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "summary_latency",
			Help:       "summary of module latency",
			Objectives: p.c.Objectives,
		},
		[]string{"app", "module", "api", "method"},
	)
	p.reg.MustRegister(p.summaryLatency)

	// 初始化状态度量指标
	p.gaugeState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gauge_state",
			Help: "gauge of app state",
		},
		[]string{"app", "module", "state"},
	)
	p.reg.MustRegister(p.gaugeState)

	// 初始化异常计数器
	p.counterException = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_exception",
			Help: "number of module exception",
		},
		[]string{"app", "module", "exception"},
	)
	p.reg.MustRegister(p.counterException)

	// 初始化事件计数器
	p.counterEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_event",
			Help: "number of module event",
		},
		[]string{"app", "module", "event"},
	)
	p.reg.MustRegister(p.counterEvent)

	// 初始化站点事件计数器
	p.counterSiteEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_site_event",
			Help: "number of module site event",
		},
		[]string{"app", "module", "event", "site"},
	)
	p.reg.MustRegister(p.counterSiteEvent)

	// 如果启用默认收集,注册 Go 运行时指标收集器
	if p.c.DefaultCollect {
		p.reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		p.reg.MustRegister(collectors.NewGoCollector())
	}
}

// run 启动 HTTP 服务器暴露 Prometheus 指标
func (p *PrometheusWrapper) run() {
	// 如果未配置端口则不启动服务
	if p.c.ListenPort == 0 {
		return
	}

	go func() {
		// 创建 Prometheus 指标处理器
		handle := promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})
		// 注册 /metrics 路径,并添加基础认证
		http.Handle("/metrics", promhttp.InstrumentMetricHandler(
			p.reg,
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// 进行基础认证验证
				username, pwd, ok := req.BasicAuth()
				if !ok || !(username == p.c.BasicUserName && pwd == p.c.BasicPassword) {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte("401 Unauthorized"))
					return
				}
				handle.ServeHTTP(w, req)
			})),
		)
		log.Printf("Prometheus listening on: %d", p.c.ListenPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.c.ListenPort), nil))
	}()
}

// Log 记录完整的请求信息,包括 API、方法、状态码、发送接收字节数和延迟时间
func (p *PrometheusWrapper) Log(api, method, code string, sendBytes, rcvdBytes, latency float64) {
	if !p.c.Enable {
		return
	}
	if len(p.c.LogMethod) > 0 {
		if _, ok := p.c.LogMethod[method]; !ok {
			return
		}
	}
	if len(p.c.LogApi) > 0 {
		if _, ok := p.c.LogApi[api]; !ok {
			return
		}
	}

	p.counterRequests.WithLabelValues(p.c.App, "self", api, method, code).Inc()
	if sendBytes > 0 {
		p.counterSendBytes.WithLabelValues(p.c.App, "self", api, method, code).Add(sendBytes)
	}
	if rcvdBytes > 0 {
		p.counterRcvdBytes.WithLabelValues(p.c.App, "self", api, method, code).Add(rcvdBytes)
	}
	if len(p.c.Buckets) > 0 {
		p.histogramLatency.WithLabelValues(p.c.App, "self", api, method).Observe(latency)
	}
	if len(p.c.Objectives) > 0 {
		p.summaryLatency.WithLabelValues(p.c.App, "self", api, method).Observe(latency)
	}
}

// RequestLog 记录请求次数
func (p *PrometheusWrapper) RequestLog(module, api, method, code string) {
	if !p.c.Enable {
		return
	}
	p.counterRequests.WithLabelValues(p.c.App, module, api, method, code).Inc()
}

// SendBytesLog 记录发送的字节数
func (p *PrometheusWrapper) SendBytesLog(module, api, method, code string, byte float64) {
	if !p.c.Enable {
		return
	}
	p.counterSendBytes.WithLabelValues(p.c.App, module, api, method, code).Add(byte)
}

// RcvdBytesLog 记录接收的字节数
func (p *PrometheusWrapper) RcvdBytesLog(module, api, method, code string, byte float64) {
	if !p.c.Enable {
		return
	}
	p.counterRcvdBytes.WithLabelValues(p.c.App, module, api, method, code).Add(byte)
}

// HistogramLatencyLog 记录延迟时间的直方图数据
func (p *PrometheusWrapper) HistogramLatencyLog(module, api, method string, latency float64) {
	if !p.c.Enable {
		return
	}
	p.histogramLatency.WithLabelValues(p.c.App, module, api, method).Observe(latency)
}

// SummaryLatencyLog 记录延迟时间的摘要数据
func (p *PrometheusWrapper) SummaryLatencyLog(module, api, method string, latency float64) {
	if !p.c.Enable {
		return
	}
	p.summaryLatency.WithLabelValues(p.c.App, module, api, method).Observe(latency)
}

// ExceptionLog 记录异常事件
func (p *PrometheusWrapper) ExceptionLog(module, exception string) {
	if !p.c.Enable {
		return
	}
	p.counterException.WithLabelValues(p.c.App, module, exception).Inc()
}

// EventLog 记录普通事件
func (p *PrometheusWrapper) EventLog(module, event string) {
	if !p.c.Enable {
		return
	}
	p.counterEvent.WithLabelValues(p.c.App, module, event).Inc()
}

// SiteEventLog 记录站点相关事件
func (p *PrometheusWrapper) SiteEventLog(module, event, site string) {
	if !p.c.Enable {
		return
	}
	p.counterSiteEvent.WithLabelValues(p.c.App, module, event, site).Inc()
}

// StateLog 记录状态指标
func (p *PrometheusWrapper) StateLog(module, state string, value float64) {
	if !p.c.Enable {
		return
	}
	p.gaugeState.WithLabelValues(p.c.App, module, state).Set(value)
}

// ResetCounter 重置所有计数器类型的指标,包括站点事件、普通事件、异常事件、
// 接收字节数和发送字节数等计数器
func (p *PrometheusWrapper) ResetCounter() {
	if !p.c.Enable {
		return
	}
	p.counterSiteEvent.Reset()
	p.counterEvent.Reset()
	p.counterException.Reset()
	p.counterRcvdBytes.Reset()
	p.counterSendBytes.Reset()
}

// RegCustomCollector 注册自定义的 Prometheus 收集器到注册器中
// 参数 c 为要注册的自定义收集器实例
func (p *PrometheusWrapper) RegCustomCollector(c prometheus.Collector) {
	p.reg.MustRegister(c)
}

// NewPrometheusWrapper 创建一个新的 Prometheus 包装器实例
// 参数 conf 为监控配置信息,包含是否启用监控、应用名称、监听端口等配置
// 如果未指定应用名称,默认使用"app"
// 如果启用监控但未指定端口,默认使用9100端口
// 返回初始化好的 Prometheus 包装器实例
func NewPrometheusWrapper(conf *Config) *PrometheusWrapper {
	if conf.App == "" {
		conf.App = "app"
	}
	if conf.Enable && conf.ListenPort == 0 {
		conf.ListenPort = 9100
	}

	w := &PrometheusWrapper{
		c:   *conf,
		reg: prometheus.NewRegistry(),
	}

	if conf.Enable {
		w.init()
		w.run()
	}

	return w
}
