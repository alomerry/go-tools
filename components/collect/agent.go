package collect

import (
  "context"
  "sync"
  "time"
  
  "github.com/alomerry/go-tools/components/log"
  monitor2 "github.com/alomerry/go-tools/components/monitor"
  "github.com/alomerry/go-tools/utils/trace"
  "github.com/sirupsen/logrus"
  
  "github.com/alomerry/go-tools/model/monitor"
)

// ClientInfo 客户端信息结构，用于Redis存储
type ClientInfo struct {
	Hostname     string    `redis:"hostname"`
	LastReportAt time.Time `redis:"last_report_at"`
	CpuPercent   float64   `redis:"cpu_percent"`
	MemPercent   float64   `redis:"mem_percent"`
	DiskPercent  float64   `redis:"disk_percent"`
	NetBytesSent float64   `redis:"net_bytes_sent"`
	NetBytesRec  float64   `redis:"net_bytes_rec"`
	Version      string    `redis:"version"`
}

// Agent 通用采集器，定期采集本机信息并执行上报流程
type Agent struct {
  ctx    context.Context
  cancel context.CancelFunc
  
  mu sync.RWMutex
  
  // 配置选项
  interval     time.Duration
  collector    func() (*monitor.SystemStats, error)
  reporter     func(*monitor.SystemStats) error
  onError      func(error)
  startTime    time.Time
  running      bool
  lastCollect  time.Time
  collectCount int
}

// Option 配置选项函数类型
type Option func(*Agent)

// WithInterval 设置采集间隔
func WithInterval(interval time.Duration) Option {
  return func(a *Agent) {
    if interval > 0 {
      a.interval = interval
    }
  }
}

// WithCollector 设置自定义数据收集函数
// 如果不设置，将使用默认的 monitor.CollectStats
func WithCollector(collector func() (*monitor.SystemStats, error)) Option {
  return func(a *Agent) {
    if collector != nil {
      a.collector = collector
    }
  }
}

// WithReporter 设置数据上报函数
// 这是必需的选项，如果不设置，Agent 将无法上报数据
func WithReporter(reporter func(*monitor.SystemStats) error) Option {
  return func(a *Agent) {
    if reporter != nil {
      a.reporter = reporter
    }
  }
}

// WithErrorHandler 设置错误处理函数
func WithErrorHandler(onError func(error)) Option {
  return func(a *Agent) {
    if onError != nil {
      a.onError = onError
    }
  }
}

// NewAgent 创建新的Agent实例
// 如果未提供 Collector，将使用默认的 monitor.CollectStats
// Reporter 是必需的，如果未提供，将返回错误
func NewAgent(opts ...Option) (*Agent, error) {
  agent := &Agent{
    interval:  1 * time.Minute, // 默认间隔1分钟
    startTime: time.Now(),
  }
  
  // 应用所有选项
  for _, opt := range opts {
    opt(agent)
  }
  
  // 设置默认收集器
  if agent.collector == nil {
    agent.collector = monitor2.CollectStats
  }
  
  // 验证必需选项
  if agent.reporter == nil {
    return nil, ErrReporterRequired
  }
  
  // 创建上下文
  agent.ctx, agent.cancel = context.WithCancel(trace.NewContext(nil))
  
  return agent, nil
}

// Start 启动采集器
func (a *Agent) Start() {
  a.mu.Lock()
  defer a.mu.Unlock()
  
  if a.running {
    log.Warn(a.ctx, "Agent already running")
    return
  }
  
  a.running = true
  go a.run()
  log.Info(a.ctx, "Agent started")
}

// Stop 停止采集器
func (a *Agent) Stop() {
  a.mu.Lock()
  defer a.mu.Unlock()
  
  if !a.running {
    return
  }
  
  a.running = false
  if a.cancel != nil {
    a.cancel()
  }
  logrus.Info("Agent stopped")
}

// IsRunning 检查采集器是否正在运行
func (a *Agent) IsRunning() bool {
  a.mu.RLock()
  defer a.mu.RUnlock()
  return a.running
}

// GetStats 获取采集器统计信息
func (a *Agent) GetStats() (map[string]interface{}, error) {
  a.mu.RLock()
  defer a.mu.RUnlock()
  
  return map[string]interface{}{
    "start_time":    a.startTime,
    "running":       a.running,
    "interval":      a.interval,
    "collect_count": a.collectCount,
    "last_collect":  a.lastCollect,
  }, nil
}

// run 采集循环
func (a *Agent) run() {
  ticker := time.NewTicker(a.interval)
  defer func() {
    ticker.Stop()
    a.mu.Lock()
    a.running = false
    a.mu.Unlock()
  }()
  
  for {
    select {
    case <-a.ctx.Done():
      return
    case <-ticker.C:
      a.collectAndReport()
    }
  }
}

// collectAndReport 收集并上报数据
func (a *Agent) collectAndReport() {
  startTime := time.Now()
  stats, err := a.collector()
  if err != nil {
    a.handleError(err)
    return
  }
  
  a.mu.Lock()
  a.collectCount++
  a.lastCollect = time.Now()
  a.mu.Unlock()
  
  err = a.reporter(stats)
  if err != nil {
    a.handleError(err)
    return
  }
  
  log.WithFields(logrus.Fields{
    "duration": time.Since(startTime),
    "cpu":      stats.CpuUsage,
    "memory":   stats.MemoryUsage,
    "disk":     stats.DiskUsage,
  }).Infof(a.ctx, "Data collection and reporting completed")
}

// handleError 处理错误
func (a *Agent) handleError(err error) {
  if a.onError != nil {
    a.onError(err)
    return
  }
  log.Errorf(a.ctx, "Agent error: %v", err)
}

// ErrReporterRequired 错误：未提供上报函数
var ErrReporterRequired = &AgentError{message: "reporter function is required"}

// AgentError Agent相关错误
type AgentError struct {
  message string
}

func (e *AgentError) Error() string {
  return e.message
}
