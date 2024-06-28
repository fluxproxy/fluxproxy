package main

import (
	"context"
	"fluxway"
	"fluxway/helper"
	"fluxway/proxy"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/sirupsen/logrus"
)

// Configuration
var k = koanf.NewWithConf(koanf.Conf{
	Delim:       ".",
	StrictMerge: true,
})

func runAsFullServer(runCtx context.Context, args []string) error {
	return runCommandAs(runCtx, args, "" /*full by config*/)
}

func runAsForwardServer(runCtx context.Context, args []string) error {
	return runCommandAs(runCtx, args, fluxway.ServerModeForward)
}

func runAsProxyServer(runCtx context.Context, args []string) error {
	return runCommandAs(runCtx, args, fluxway.ServerModeProxy)
}

func runCommandAs(runCtx context.Context, args []string, serverMode string) error {
	confpath := "config.yml"
	if len(args) > 0 {
		confpath = args[0]
	}
	logrus.Infof("main: load config file: %s", confpath)
	if err := k.Load(file.Provider(confpath), yaml.Parser()); err != nil {
		return fmt.Errorf("load config file %s: %w", confpath, err)
	}
	// Instance
	runCtx = context.WithValue(runCtx, proxy.CtxKeyConfiger, k)
	inst := fluxway.NewInstance()
	if err := inst.Init(runCtx, serverMode); err != nil {
		return fmt.Errorf("main: instance start: %w", err)
	}
	return helper.ErrIf(inst.Serve(runCtx), "main: instance serve")
}
