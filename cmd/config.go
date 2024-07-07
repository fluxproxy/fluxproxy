package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/fluxproxy/fluxproxy/app"
	"github.com/sirupsen/logrus"
	"os"
)

//go:embed config.toml
var configContent string

func runConfigGenerate(runCtx context.Context, args []string) error {
	var outfile string
	fs := flag.NewFlagSet("config-gen", flag.ContinueOnError)
	fs.StringVar(&outfile, "out", "./config-gen.toml", "output config file path")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("main: invalid flags. %w", err)
	}
	if _, err := os.Stat(outfile); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("main: config file is already exists: %s", outfile)
	}
	f, err := os.OpenFile(outfile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("main: open config file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(configContent); err != nil {
		return fmt.Errorf("main: write config content: %w", err)
	}
	logrus.Infof("main: generate config file: %s", outfile)
	return nil
}

func runConfigVerify(runCtx context.Context, args []string) error {
	if err := runAppWith(runCtx, args, app.RunServerModeAuto, true); err != nil {
		return err
	}
	logrus.Infof("main: verify config successfully.")
	return nil
}
