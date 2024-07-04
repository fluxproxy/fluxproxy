package app

import (
	"context"
	"fmt"
	"github.com/bytepowered/goes"
	"github.com/knadh/koanf/parsers/yaml"
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
}

func RunAsMode(runCtx context.Context, args []string, cmdMode string) error {
	confpath := "config.yml"
	if len(args) > 0 {
		confpath = args[0]
	}
	if err := k.Load(file.Provider(confpath), yaml.Parser()); err != nil {
		return fmt.Errorf("main: load config: %s. %w", confpath, err)
	}
	switch k.String("log.format") {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors:    false,
			DisableTimestamp: false,
			FullTimestamp:    true,
		})
	}
	logrus.SetReportCaller(false)
	logrus.Infof("main: load: %s", confpath)
	// Instance
	runCtx = context.WithValue(runCtx, rocket.CtxKeyConfiger, k)
	inst := NewInstance()
	if err := inst.Init(runCtx, cmdMode); err != nil {
		return fmt.Errorf("main: instance start. %w", err)
	}
	return helper.ErrIf(inst.Serve(runCtx), "main: instance serve, %s")
}
