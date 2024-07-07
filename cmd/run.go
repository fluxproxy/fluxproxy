package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fluxproxy/fluxproxy/app"
	"github.com/fluxproxy/fluxproxy/helper"
)

func runAppServers(runCtx context.Context, args []string) error {
	return runAppWith(runCtx, args, app.RunServerModeAuto, false)
}

func runAppAsHttp(runCtx context.Context, args []string) error {
	return runAppWith(runCtx, args, app.RunServerModeHttp, false)
}

func runAppAsSocks(runCtx context.Context, args []string) error {
	return runAppWith(runCtx, args, app.RunServerModeSocks, false)
}

func runAppWith(runCtx context.Context, args []string, mode string, dryRun bool) error {
	var confpath string
	fs := flag.NewFlagSet("run-app", flag.ContinueOnError)
	fs.StringVar(&confpath, "config", "./config.toml", "config file path")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("main: invalid flags. %w", err)
	}
	if err := loadConfig(runCtx, confpath); err != nil {
		return fmt.Errorf("main: load config: %s. %w", confpath, err)
	}
	inst := app.NewApp()
	if err := inst.Init(runCtx, mode); err != nil {
		return fmt.Errorf("main: init app. %w", err)
	}
	if !dryRun {
		return helper.ErrIf(inst.Serve(runCtx), "main: serve app, %s")
	}
	return nil
}
