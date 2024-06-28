package main

import (
	"context"
	"github.com/cristalhq/acmd"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// cli: https://github.com/cristalhq/acmd
func main() {
	cmds := []acmd.Command{
		{
			Name:        "run",
			Description: "Run as a proxy server, full features",
			ExecFunc:    runAsFullServer,
		},
		{
			Name:        "proxy",
			Description: "Run as a proxy server, as server mode: proxy",
			ExecFunc:    runAsProxyServer,
		},
		{
			Name:        "forward",
			Description: "Run as a forward server, as server mode: forward",
			ExecFunc:    runAsForwardServer,
		},
	}
	cmdCtx, cmdCancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName: "fluxway",
		Version: "2024.1",
		Context: cmdCtx,
	})
	go func() {
		<-signals
		cmdCancel()
	}()
	if err := r.Run(); err != nil {
		logrus.Fatal(err)
	}
}
