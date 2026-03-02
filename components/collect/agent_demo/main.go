package main

import (
  "context"
  "log"
  "os"
  "time"
)

func main() {
  f, err := os.CreateTemp(os.TempDir(), "*.txt")
  if err != nil {
    log.Fatal(err)
  }
  defer func(f *os.File) {
    err := f.Close()
    if err != nil {
      log.Fatal(err)
    }
  }(f)
  
  
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  
  for {
    select {
    case <-ctx.Done():
      return
      default:
        if _, err := f.Write(make([]byte, 1024)); err != nil {
          log.Fatal(err)
        }
    }
    
    time.Sleep(1 * time.Second)
  }
}