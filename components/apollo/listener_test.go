package apollo

import (
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestListener(t *testing.T) {
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
