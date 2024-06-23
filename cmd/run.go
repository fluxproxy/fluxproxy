package main

import (
	"fluxway"
	"fluxway/common"
	"fluxway/proxy"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func runCommand(ctx *cli.Context) error {
	confpath := "conf.yml"
	if ctx.NArg() > 0 {
		confpath = ctx.Args().Get(0)
	}
	// Configuration
	var k = koanf.NewWithConf(koanf.Conf{
		Delim:       ".",
		StrictMerge: true,
	})
	logrus.Infof("load config file: %s", confpath)
	if err := k.Load(file.Provider(confpath), yaml.Parser()); err != nil {
		return fmt.Errorf("load config file %s: %w", confpath, err)
	}
	// Instance
	inst := fluxway.NewInstance(proxy.ContextWithConfig(ctx.Context, k))
	if err := inst.Start(); err != nil {
		return fmt.Errorf("instance start: %w", err)
	}
	defer func() {
		common.LogIf(inst.Stop(), "instance stop: %w")
	}()
	return common.ErrIf(inst.Serve(), "instance serve")
}
