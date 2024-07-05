package app

import (
	"context"
	"fmt"
	"github.com/bytepowered/goes"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

// Configuration
var k = koanf.NewWithConf(koanf.Conf{
	Delim:       ".",
	StrictMerge: true,
})

func init() {
	goes.SetPanicHandler(func(ctx context.Context, r interface{}) {
		logrus.Errorf("goroutine panic %v: %s", r, debug.Stack())
	})
	logrus.SetReportCaller(false)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: false,
		FullTimestamp:    true,
		SortingFunc: func(fields []string) {
			for i, f := range fields {
				if f == "id" {
					fields[i], fields[0] = fields[0], fields[i]
					break
				}
			}
		},
	})
}

func RunAsMode(runCtx context.Context, args []string, cmdMode string) error {
	confpath := "config.toml"
	if len(args) > 0 {
		confpath = args[0]
	}
	if err := k.Load(file.Provider(confpath), toml.Parser()); err != nil {
		return fmt.Errorf("main: load config: %s. %w", confpath, err)
	}

	logrus.Infof("main: load: %s", confpath)
	// App
	runCtx = context.WithValue(runCtx, rocket.CtxKeyConfiger, k)
	inst := NewApp()
	if err := inst.Init(runCtx, cmdMode); err != nil {
		return fmt.Errorf("main: instance start. %w", err)
	}
	return helper.ErrIf(inst.Serve(runCtx), "main: instance serve, %s")
}
