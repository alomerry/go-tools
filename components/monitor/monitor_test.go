package monitor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

func TestInfo(t *testing.T) {
	// 1. 获取 CPU 逻辑核心数
	count, _ := cpu.Counts(true)
	physicalCount, _ := cpu.Counts(false)
	fmt.Printf("逻辑 CPU 核心数: %d, 物理 CPU 核心数: %d\n", count, physicalCount)

	// 3. 获取 CPU 使用率（总体）
	// false 表示所有CPU的平均值，true 表示每个CPU核心单独显示
	percentages, _ := cpu.Percent(2*time.Second, false)
	fmt.Printf("过去 2 秒 CPU 总使用率: %.2f%%\n", percentages[0])

	// 1. 获取虚拟内存（物理内存）信息
	v, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	fmt.Printf("总内存: %s\n", humanize.IBytes(v.Total))
	fmt.Printf("已使用: %s (%.2f%%)\n", humanize.IBytes(v.Used), v.UsedPercent)

	// 2. 获取网络连接信息
	connections, err := net.Connections("all")
	if err != nil {
		panic(err)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	fmt.Println("网络接口信息:")
	for _, iface := range interfaces {
		fmt.Printf("接口: %s (MTU: %d, MAC: %s)\n",
			iface.Name, iface.MTU, iface.HardwareAddr)

		for _, addr := range iface.Addrs {
			fmt.Printf("  IP地址: %s\n", addr.Addr)
		}
	}
	fmt.Printf("\n网络连接数: %d\n", len(connections))

	// 4. 获取协议统计信息
	protoStats, err := net.ProtoCounters([]string{"tcp", "udp", "icmp"})
	if err == nil {
		fmt.Println("\n协议统计:")
		for _, stat := range protoStats {
			fmt.Printf("协议: %s\n", stat.Protocol)
			for k, v := range stat.Stats {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}

	// 1. 获取所有进程列表
	/*processes, err := process.Processes()
	if err != nil {
		panic(err)
	}

	fmt.Printf("系统共有 %d 个进程\n", len(processes))
	for i, pp := range processes {
		if i%10 != 0 {
			continue
		}
		p, err := process.NewProcess(pp.Pid)
		if err != nil {
			panic(err)
		}
		// 获取进程名
		name, _ := p.Name()
		fmt.Printf("\t进程 %d 名称: %s\n", pp.Pid, name)

		// 获取进程状态
		status, _ := p.Status()
		fmt.Printf("\t进程状态: %s\n", status)

		// 3. 获取进程 CPU 使用率
		cpuPercent, _ := p.CPUPercent()
		fmt.Printf("\tCPU 使用率: %.2f%%\n", cpuPercent)

		// 4. 获取进程内存信息
		memInfo, _ := p.MemoryInfo()
		if memInfo != nil {
			fmt.Printf("\t内存使用: RSS=%s, VMS=%s\n",
				humanize.IBytes(memInfo.RSS), humanize.IBytes(memInfo.VMS))
		}

		// 5. 获取进程内存百分比
		memPercent, _ := p.MemoryPercent()
		fmt.Printf("\t内存使用百分比: %.2f%%\n", memPercent)

		// 6. 获取进程创建时间
		//createTime, _ := p.CreateTime()
		//fmt.Printf("\t进程创建时间: %v\n", createTime)

		// 7. 获取进程命令行参数
		//cmdline, _ := p.Cmdline()
		//fmt.Printf("\t命令行: %s\n", cmdline)

		// 8. 获取进程打开的文件
		//openFiles, _ := p.OpenFiles()
		//fmt.Printf("\t打开文件数: %d\n", len(openFiles))

		// 9. 获取进程网络连接
		connections1, _ := p.Connections()
		fmt.Printf("\t网络连接数: %d\n", len(connections1))

		// 10. 获取进程线程数
		threads, _ := p.NumThreads()
		fmt.Printf("\t线程数: %d\n", threads)
	}*/

	//info, err := host.Info()
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Printf("主机名: %s\n", info.Hostname)
	//fmt.Printf("操作系统: %s %s %s\n",
	//	info.OS, info.Platform, info.PlatformVersion)
	//fmt.Printf("内核版本: %s\n", info.KernelVersion)
	//fmt.Printf("运行时间: %v\n", info.Uptime)
	//fmt.Printf("启动时间: %v\n", info.BootTime)
	//fmt.Printf("进程数: %d\n", info.Procs)

	// 2. 获取温度信息（需要硬件支持）

	// 3. 获取当前登录用户
	//users, err := host.Users()
	//if err == nil {
	//	fmt.Println("\n登录用户:")
	//	for _, user := range users {
	//		fmt.Printf("用户: %s, 终端: %s, 主机: %s, 登录时间: %v\n",
	//			user.User, user.Terminal, user.Host, user.Started)
	//	}
	//}

	//// 4. 获取主机 ID
	//hostID, err := host.HostID()
	//if err == nil {
	//	fmt.Printf("主机ID: %s\n", hostID)
	//}

	avg, err := load.Avg()
	if err != nil {
		fmt.Println("获取负载失败:", err)
		return
	}

	fmt.Printf("系统负载 (1分钟): %.2f\n", avg.Load1)
	fmt.Printf("系统负载 (5分钟): %.2f\n", avg.Load5)
	fmt.Printf("系统负载 (15分钟): %.2f\n", avg.Load15)

	// 获取负载占CPU核心数的百分比
	cpuCount, _ := cpu.Counts(true)
	fmt.Printf("CPU核心数: %d\n", cpuCount)
	fmt.Printf("1分钟负载占比: %.2f%%\n", (avg.Load1/float64(cpuCount))*100)
}

func TestSystemMonitor(t *testing.T) {
	var (
		ctx, cancel = context.WithTimeout(context.TODO(), time.Second*30)
		monitor     = NewSystemMonitor(WithContext(ctx), WithInterval(5*time.Second))
	)

	defer cancel()

	monitor.Watch()

	<-ctx.Done()
}
