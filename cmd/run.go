package main

import (
	"fluxway"
	"fluxway/common"
	"fluxway/proxy"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/urfave/cli/v2"
)

func runCommand(ctx *cli.Context) error {
	var k = koanf.NewWithConf(koanf.Conf{
		Delim:       ".",
		StrictMerge: true,
	})
	inst := fluxway.NewInstance(proxy.ContextWithConfig(ctx.Context, k))
	if err := inst.Start(); err != nil {
		return fmt.Errorf("instance start: %w", err)
	}
	defer func() {
		common.LogIf(inst.Stop(), "instance stop: %w")
	}()
	return common.ErrIf(inst.Serve(), "instance serve")
}
