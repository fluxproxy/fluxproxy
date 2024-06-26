package main

import (
	"github.com/cristalhq/acmd"
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
	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName: "fluxway",
		Version: "2024.1",
	})
	if err := r.Run(); err != nil {
		r.Exit(err)
	}
}
