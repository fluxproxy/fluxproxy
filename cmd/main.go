package main

import (
	"context"
	"github.com/cristalhq/acmd"
	"github.com/rocket-proxy/rocket-proxy/app"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	BuildVersion = "2024.1.0"
)

// cli: https://github.com/cristalhq/acmd
func main() {
	cmds := []acmd.Command{
		// Server
		{
			Name:        "run",
			Description: "Run server, mode: all features",
			ExecFunc:    runAsAutoServer,
		},
		{
			Name:        "http",
			Description: "Run server, mode(only): http",
			ExecFunc:    runAsHttpServer,
		},
		{
			Name:        "socks",
			Description: "Run server, mode(only): socks",
			ExecFunc:    runAsSocksServer,
		},
	}
	cmdCtx, cmdCancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName: "rocket-proxy",
		Version: BuildVersion,
		Context: cmdCtx,
	})
	go func() {
		<-signals
		cmdCancel()
	}()
	logrus.Infof("main: %s", BuildVersion)
	if err := r.Run(); err != nil {
		logrus.Fatal(err)
	}
}

func runAsAutoServer(runCtx context.Context, args []string) error {
	return app.RunAsMode(runCtx, args, app.RunServerModeAuto)
}

func runAsHttpServer(runCtx context.Context, args []string) error {
	return app.RunAsMode(runCtx, args, app.RunServerModeHttp)
}

func runAsSocksServer(runCtx context.Context, args []string) error {
	return app.RunAsMode(runCtx, args, app.RunServerModeSocks)
}
