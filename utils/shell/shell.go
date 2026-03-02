package shell

import (
  "bytes"
  "fmt"
  "os/exec"
)

func ExecuteCommand(command string) (string, error) {
  cmd := exec.Command("sh", "-c", command)
  var out bytes.Buffer
  var stderr bytes.Buffer
  cmd.Stdout = &out
  cmd.Stderr = &stderr
  err := cmd.Run()
  
  if err != nil {
    return "", fmt.Errorf("run command error: %v, ERROR: %s", err, stderr.String())
  }
  if stderr.String() != "" {
    return "", fmt.Errorf("exec command error: %s", stderr.String())
  }
  return out.String(), nil
}