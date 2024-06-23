package main

import (
	"fluxway"
	"fluxway/common"
	"fmt"
	"github.com/urfave/cli/v2"
)

func runCommand(ctx *cli.Context) error {
	inst := fluxway.NewInstance(ctx.Context)
	if err := inst.Start(); err != nil {
		return fmt.Errorf("instance start: %w", err)
	}
	defer func() {
		common.LogIf(inst.Stop(), "instance stop: %w")
	}()
	return common.ErrIf(inst.Serve(), "instance serve")
}
