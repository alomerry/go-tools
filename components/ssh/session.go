package ssh

import (
  "bytes"
  "context"
  "fmt"
  
  "github.com/alomerry/go-tools/components/log"
  "golang.org/x/crypto/ssh"
)

type session struct {
	*ssh.Session
}

func (session session) Run(ctx context.Context, cmd string) (string,error) {
  var (
  	stdout, stderr bytes.Buffer
    err            error
  )
  session.Stdout = &stdout
  session.Stderr = &stderr
  log.Infof(ctx, "Executing remote command: %s", cmd)
  err = session.Session.Run(cmd)
  if err != nil {
    return stdout.String(), fmt.Errorf("command execution failed: %v, stderr: %s", err, stderr.String())
  }
  
  return stdout.String(), nil
}