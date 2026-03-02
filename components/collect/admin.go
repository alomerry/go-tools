package collect

import (
	"context"
	"fmt"

	"github.com/alomerry/go-tools/components/log"
	"github.com/alomerry/go-tools/components/ssh"
	"github.com/alomerry/go-tools/utils/shell"
)

type agentAdmin struct {
	ssh ssh.Client
}

func NewAgentAdmin(ctx context.Context, options ...ssh.Option) (agentAdmin, error) {
	client, err := ssh.NewClient(ctx, options...)
	if err != nil {
		return agentAdmin{}, err
	}

	if err := client.Connect(); err != nil {
		return agentAdmin{}, err
	}
	return agentAdmin{
		ssh: client,
	}, nil
}

func (a *agentAdmin) RegisterAgent(ctx context.Context) (string, error) {
	const (
		scp = `scp /tmp/agent_demo root@%s:/tmp/`
	)
	command, err := shell.ExecuteCommand(fmt.Sprintf(scp, a.ssh.Config().Host()))
	if err != nil {
		return "", err
	}

	log.Infof(ctx, "Command: %s", command)

	session, err := a.ssh.Session(ctx)
	if err != nil {
		return "", err
	}
	defer session.Close()

	res, err := session.Run(ctx, "ls /tmp")
	if err != nil {
		return "", err
	}
	log.Infof(ctx, "Result: %s", res)
	return "", nil
}

func (a *agentAdmin) Close(ctx context.Context) error {
	return a.ssh.Close()
}

// // GetUptime 获取系统运行时间
// func (c *SSHClient) GetUptime(ctx context.Context) (string, error) {
//   return c.RunCommand(ctx, "uptime -p")
// }
//
// // GetDiskUsage 获取磁盘使用情况
// func (c *SSHClient) GetDiskUsage(ctx context.Context) (string, error) {
//   return c.RunCommand(ctx, "df -h")
// }
//
// // GetMemUsage 获取内存使用情况
// func (c *SSHClient) GetMemUsage(ctx context.Context) (string, error) {
//   return c.RunCommand(ctx, "free -h")
// }
//
// // GetCPUUsage 获取 CPU 使用概况 (top batch mode)
// func (c *SSHClient) GetCPUUsage(ctx context.Context) (string, error) {
//   // top -bn1 | grep "Cpu(s)"
//   return c.RunCommand(ctx, "top -bn1 | grep 'Cpu(s)'")
// }
