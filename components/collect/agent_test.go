package collect

import (
	"fmt"
  "testing"
  "time"

	monitor2 "github.com/alomerry/go-tools/components/monitor"
	"github.com/sirupsen/logrus"

	"github.com/alomerry/go-tools/model/monitor"
)

func TestAgentUsage(t*testing.T) {
	// 创建一个简单的上报函数，将数据打印到控制台
	reporter := func(stats *monitor.SystemStats) error {
		fmt.Printf("系统监控数据:\n")
		fmt.Printf("  时间: %s\n", stats.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  CPU使用率: %.2f%%\n", stats.CPUUsage)
		fmt.Printf("  内存使用率: %.2f%% (%d/%d MB)\n", stats.MemoryUsage,
			stats.UsedMemory/1024/1024, stats.TotalMemory/1024/1024)
		fmt.Printf("  磁盘使用率:\n")
		for mountpoint, usage := range stats.DiskUsage {
			fmt.Printf("    %s: %.2f%%\n", mountpoint, usage)
		}
		fmt.Printf("  负载平均: %.2f, %.2f, %.2f\n",
			stats.LoadAvg[0], stats.LoadAvg[1], stats.LoadAvg[2])
		fmt.Printf("  IP地址: %s\n", stats.Ip)
		return nil
	}

	// 创建错误处理函数
	errorHandler := func(err error) {
		logrus.Errorf("Agent 错误: %v", err)
	}

	// 创建 Agent 实例，设置 10 秒间隔（演示用）
	agent, err := NewAgent(
		WithInterval(3*time.Second),
		WithReporter(reporter),
		WithErrorHandler(errorHandler),
	)
	if err != nil {
		logrus.Fatalf("创建 Agent 失败: %v", err)
	}

	// 启动Agent
	agent.Start()
	fmt.Println("Agent 已启动，将每 10 秒收集一次系统信息...")

	// 运行一段时间后停止
	time.Sleep(15 * time.Second)
	agent.Stop()
	fmt.Println("Agent 已停止")

	// 获取 Agent 统计信息
	stats, err := agent.GetStats()
	if err != nil {
		logrus.Errorf("获取 Agent 统计信息失败: %v", err)
		return
	}
	fmt.Printf("Agent 统计信息: %+v\n", stats)
}

func TestCustomCollector(t*testing.T) {
	// 自定义收集器，可以收集任何你想要的数据
	customCollector := func() (*monitor.SystemStats, error) {
		// 调用默认收集器获取基础系统信息
		stats, err := monitor2.CollectStats()
		if err != nil {
			return nil, err
		}

		// 可以在这里添加自定义数据，例如：
		// stats.CustomField = "自定义数据"
		// 由于SystemStats结构体没有导出字段，我们可以创建一个包装结构体

		return stats, nil
	}

	// 自定义上报函数
	customReporter := func(stats *monitor.SystemStats) error {
		// 自定义上报逻辑，例如发送到HTTP API
		logrus.Infof("自定义上报: CPU=%.2f%%, Memory=%.2f%%",
			stats.CPUUsage, stats.MemoryUsage)
		return nil
	}

	// 创建Agent，使用自定义收集器
	agent, err := NewAgent(
		WithInterval(1*time.Second),
		WithCollector(customCollector),
		WithReporter(customReporter),
	)
	if err != nil {
		logrus.Fatalf("创建 Agent 失败: %v", err)
	}

	agent.Start()
	defer agent.Stop()

	// 运行一段时间
	time.Sleep(5 * time.Second)
}

func TestAgentWithContext(t*testing.T) {
	// 创建上报函数
	reporter := func(stats *monitor.SystemStats) error {
		// 上报逻辑
		logrus.Infof("触发上报。")
		return nil
	}

	// 创建 Agent
	agent, err := NewAgent(
		WithInterval(1*time.Second),
		WithReporter(reporter),
	)
	if err != nil {
		logrus.Fatalf("创建 Agent 失败: %v", err)
	}

	// 启动 Agent
	agent.Start()

	// 模拟外部信号，停止Agent
	go func() {
		time.Sleep(3 * time.Second)
		logrus.Info("收到停止信号，停止 Agent")
		agent.Stop()
	}()

	// 等待 Agent 停止
	time.Sleep(5 * time.Second)
	logrus.Info("示例结束")
}

func TestAgentWithRemote(t*testing.T) {
  // 创建上报函数
  reporter := func(stats *monitor.SystemStats) error {
    // 上报逻辑
    logrus.Infof("触发上报。")
    return nil
  }
  
  // 创建 Agent
  agent, err := NewAgent(
    WithInterval(1*time.Second),
    WithReporter(reporter),
  )
  if err != nil {
    logrus.Fatalf("创建 Agent 失败: %v", err)
  }
  
  // 启动 Agent
  agent.Start()
  
  // 模拟外部信号，停止Agent
  go func() {
    time.Sleep(3 * time.Second)
    logrus.Info("收到停止信号，停止 Agent")
    agent.Stop()
  }()
  
  // 等待 Agent 停止
  time.Sleep(5 * time.Second)
  logrus.Info("示例结束")
}