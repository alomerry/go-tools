package ssh

import (
  "errors"
  "time"
)

type Option func(*config)

func WithHost(host string) Option {
  return func(o *config) {
    o.host = host
  }
}

func WithPort(port int) Option {
  return func(config *config) {
    config.port = port
  }
}

func WithUser(user string) Option {
  return func(config *config) {
    config.user = user
  }
}

func WithPassword(password string) Option {
  return func(config *config) {
    config.password = password
  }
}

func WithPrivateKey(privateKey string) Option {
  return func(config *config) {
    config.privateKey = privateKey
  }
}

func WithPrivateKeyPath(privateKeyPath string) Option {
  return func(config *config) {
    config.privateKeyPath = privateKeyPath
  }
}

func WithTimeout(timeout time.Duration) Option {
  return func(config *config) {
    config.timeout = timeout
  }
}

type config struct {
  host           string        // 主机地址
  port           int           // SSH 端口，默认 22
  user           string        // 用户名
  password       string        // 密码（可选）
  privateKey     string        // 私钥内容（可选）
  privateKeyPath string        // 私钥路径（可选）
  timeout        time.Duration // 连接超时时间
}

func (c config) Host() string {
  return c.host
}

func (c config) Port() int {
  return c.port
}

func (c config) Validate() error {
  if len(c.host) == 0 {
    return errors.New("host is required")
  }
  
  if len(c.user) == 0 {
    return errors.New("user is required")
  }
  
  if len(c.password) == 0 && len(c.privateKeyPath) == 0 && len(c.privateKey) == 0 {
    return errors.New("password or private key is required")
  }
  
  return nil
}

func (c config) AuthByPrivateKey() bool {
  return c.privateKey != "" || c.privateKeyPath != ""
}