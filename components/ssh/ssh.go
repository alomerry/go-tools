package ssh

import (
  "context"
  "fmt"
  "io/ioutil"
  "net"
  "os"
  "time"
  
  "github.com/alomerry/go-tools/components/log"
  "golang.org/x/crypto/ssh"
)

type client struct {
	config config
	client *ssh.Client
}

func NewClient(ctx context.Context, options ...Option) (Client, error) {
	var (
		cfg = config{
			user:    "root",
			port:    22,
			timeout: time.Duration(10) * time.Second,
		}
	)

	for _, opt := range options {
		opt(&cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &client{config: cfg}, nil
}

// parsePrivateKey 解析私钥，支持文件路径或直接内容
func (c *client) parsePrivateKey() (ssh.Signer, error) {
	var keyBytes []byte
	var err error

	// 尝试作为文件路径读取
	if _, err = os.Stat(c.config.privateKeyPath); err == nil {
		keyBytes, err = ioutil.ReadFile(c.config.privateKeyPath)
		if err != nil {
			return nil, err
		}
	} else {
		// 假设是私钥内容
		keyBytes = []byte(c.config.privateKey)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func (c *client) Connect() error {
  
  sshConfig := &ssh.ClientConfig{
    User:    c.config.user,
    Timeout: c.config.timeout,
    HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
      // 在生产环境中应该验证 HostKey，这里为了简化操作暂时跳过
      return nil
    },
  }
  
	if c.config.AuthByPrivateKey() {
		signer, err := c.parsePrivateKey()
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
    sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
    sshConfig.Auth = []ssh.AuthMethod{ssh.Password(c.config.password)}
  }

	addr := fmt.Sprintf("%s:%d", c.config.host, c.config.port)
	cc, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial ssh: %w", err)
	}

	c.client = cc
	log.Infof(nil, "Connected to %s via SSH", addr)
	return nil
}

func (c *client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *client) Config() Config {
  return c.config
}

func (c *client) Session(ctx context.Context) (Session,error) {
  if c.client == nil {
    return nil, fmt.Errorf("client not connected")
  }
  
  ss, err := c.client.NewSession()
  if err != nil {
    return nil, fmt.Errorf("failed to create session: %w", err)
  }
  return session{ss}, nil
}
