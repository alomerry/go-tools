package main

import (
	"fmt"
	"github.com/alomerry/go-pusher/modes"
	"github.com/alomerry/go-pusher/share"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
)

var (
	configPath = pflag.String("configPath", "", "path of configuration")

	cpuProfile = "cpu.profile" // write cpu profile
	memProfile = "mem.profile" // write memory profile
)

func main() {
	pprofProfile()
	viper.BindPFlag("configPath", pflag.Lookup("configPath"))
	pflag.Parse()

	ctx := context.Background()
	modes.IClient.Init(ctx)
	modes.IClient.Run()
}

func pprofProfile() func() {
	return func() {
		if strings.HasSuffix(cpuProfile, ".profile") {
			f, err := os.Create(fmt.Sprintf("%s/pprof/%s", share.ExPath, cpuProfile))
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()
		}

		if strings.HasSuffix(memProfile, ".profile") {
			defer func() {
				f, err := os.Create(fmt.Sprintf("%s/pprof/%s", share.ExPath, memProfile))
				if err != nil {
					panic(err)
				}
				defer f.Close()
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					panic(err)
				}
			}()
		}
	}
}
