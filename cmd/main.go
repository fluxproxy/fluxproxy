package main

import (
	"context"
	"github.com/cristalhq/acmd"
	proxy "github.com/fluxproxy/fluxproxy"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	BuildVersion = "2024.1.0"
)

func init() {
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

// cli: https://github.com/cristalhq/acmd
func main() {
	cmds := []acmd.Command{
		// Server
		{
			Name:        "run",
			Description: "Run server",
			Subcommands: []acmd.Command{
				{
					Name:        "auto",
					Description: "Run server, mode: all features",
					ExecFunc:    runAppServers,
				},
				{
					Name:        "http",
					Description: "Run server, mode(only): http",
					ExecFunc:    runAppAsHttp,
				},
				{
					Name:        "socks",
					Description: "Run server, mode(only): socks",
					ExecFunc:    runAppAsSocks,
				},
			},
		},
		// Config
		{
			Name:        "config",
			Description: "Generate or verify configuration",
			Subcommands: []acmd.Command{
				{
					Name:        "generate",
					Description: "Generate config file",
					ExecFunc:    runConfigGenerate,
				},
				{
					Name:        "verify",
					Description: "Verify config file",
					ExecFunc:    runConfigVerify,
				},
			},
		},
	}

	// Configuration
	var configer = koanf.NewWithConf(koanf.Conf{
		Delim:       ".",
		StrictMerge: true,
	})
	cmdCtx, cmdCancel := context.WithCancel(context.WithValue(context.Background(), proxy.CtxKeyConfiger, configer))

	// Signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	// Runner
	runner := acmd.RunnerOf(cmds, acmd.Config{
		AppName: "fluxproxy",
		Version: BuildVersion,
		Context: cmdCtx,
	})
	go func() {
		<-signals
		cmdCancel()
	}()
	logrus.Infof("main: %s", BuildVersion)
	if err := runner.Run(); err != nil {
		logrus.Fatal(err)
	}
}

func loadConfig(ctx context.Context, confpath string) error {
	if err := proxy.Configer(ctx).Load(file.Provider(confpath), toml.Parser()); err != nil {
		return err
	}
	logrus.Infof("main: load: %s", confpath)
	return nil
}
