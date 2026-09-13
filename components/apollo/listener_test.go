package apollo

import (
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/alomerry/go-tools/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestListener 依赖真实 Apollo 集群（Init 连接失败即 panic），集成开关控制
func TestListener(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	Init("colona", "colona")

	d, err := GetJson[value]("colona.meta,dynamic")
	assert.Nil(t, err)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	tick := time.NewTicker(time.Second * 3)

	go func() {
		for {
			select {
			case <-tick.C:
				if v := d.Load(); v != nil {
					logrus.Info(v.Test)
				}
			}
		}
	}()

	for {
		select {
		case <-sigChan:
			tick.Stop()
			break
		}
	}
}

type value struct {
	Port int    `json:"port"`
	Test string `json:"test"`
}
