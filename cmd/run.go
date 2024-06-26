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

func runAsFullServer(ctx context.Context, args []string) error {
	return runCommandAs(ctx, args, "" /*full by config*/)
}

func runAsForwardServer(ctx context.Context, args []string) error {
	return runCommandAs(ctx, args, fluxway.ServerModeForward)
}

func runAsProxyServer(ctx context.Context, args []string) error {
	return runCommandAs(ctx, args, fluxway.ServerModeProxy)
}

func runCommandAs(ctx context.Context, args []string, serverMode string) error {
	confpath := "config.yml"
	if len(args) > 0 {
		confpath = args[0]
	}
	logrus.Infof("main: load config file: %s", confpath)
	if err := k.Load(file.Provider(confpath), yaml.Parser()); err != nil {
		return fmt.Errorf("load config file %s: %w", confpath, err)
	}
	// Instance
	inst := fluxway.NewInstance(proxy.ContextWithConfig(ctx, k))
	if err := inst.Start(serverMode); err != nil {
		return fmt.Errorf("main: instance start: %w", err)
	}
	defer func() {
		helper.LogIf(inst.Stop(), "main: instance stop: %w")
	}()
	return helper.ErrIf(inst.Serve(), "main: instance serve")
}
