package main

import (
	"fluxway"
	"fluxway/helper"
	"fluxway/proxy"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Configuration
var k = koanf.NewWithConf(koanf.Conf{
	Delim:       ".",
	StrictMerge: true,
})

func runAsAutoServer(ctx *cli.Context) error {
	return runCommandAs(ctx, "")
}

func runAsForwardServer(ctx *cli.Context) error {
	return runCommandAs(ctx, "forward")
}

func runAsProxyServer(ctx *cli.Context) error {
	return runCommandAs(ctx, "proxy")
}

func runCommandAs(ctx *cli.Context, serverMode string) error {
	confpath := "config.yml"
	if ctx.NArg() > 0 {
		confpath = ctx.Args().Get(0)
	}
	logrus.Infof("main: load config file: %s", confpath)
	if err := k.Load(file.Provider(confpath), yaml.Parser()); err != nil {
		return fmt.Errorf("load config file %s: %w", confpath, err)
	}
	// Instance
	inst := fluxway.NewInstance(proxy.ContextWithConfig(ctx.Context, k))
	if err := inst.Start(serverMode); err != nil {
		return fmt.Errorf("main: instance start: %w", err)
	}
	defer func() {
		helper.LogIf(inst.Stop(), "main: instance stop: %w")
	}()
	return helper.ErrIf(inst.Serve(), "main: instance serve")
}
