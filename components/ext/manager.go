package ext

import (
	"context"
	"log"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/alomerry/go-tools/utils/array"
)

var (
	extManager = make(map[string]func() Ext)

	ignoreExtensionForLocal = []string{
		cons.ExtCat,
		cons.ExtMetric,
		cons.ExtSysMonitor,
	}
)

type Ext interface {
	Init(context.Context) error
}

// Register registers a default extension implementation.
// Built-in extensions in go-tools call this in their init(). A caller may
// replace a default implementation by calling Override with the same name
// (Go runs the caller's init() after imported packages, so the caller wins).
func Register(name string, extFunc func() Ext) {
	extManager[name] = extFunc
}

// Override replaces an existing extension implementation. It is the explicit
// counterpart of Register for callers that want to customize a built-in
// extension. It panics if no implementation was registered for the name yet,
// so typos surface early instead of silently registering a new entry.
func Override(name string, extFunc func() Ext) {
	if _, exists := extManager[name]; !exists {
		log.Panicf("override extension [%v] failed: no implementation registered yet", name)
	}
	extManager[name] = extFunc
}

func LoadExt(extensions ...string) {
	var (
		ctx = context.TODO()
	)
	for _, extName := range extensions {
		if env.Local() && array.Contains(ignoreExtensionForLocal, extName) {
			continue
		}

		extFunc, exists := extManager[extName]
		if !exists {
			log.Fatalf("extension [%v] not exists", extName)
		}

		ext := extFunc()
		err := ext.Init(ctx)
		if err != nil {
			log.Panicf("init extension [%v] failed, err: %v", extName, err)
		}
	}
}
