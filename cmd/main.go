package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	Revision = "chuangshi.1"
)

// Cli: https://cli.urfave.org/v2/getting-started/
func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("version=%s revision=%s\n", cCtx.App.Version, Revision)
	}
	app := &cli.App{
		Name:    "Fluxway",
		Version: "v2014.1.0",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Run as a proxy server, mode by config",
				Action: runAsAutoServer,
			},
			{
				Name:   "forward",
				Usage:  "Run as a forward server, mode forward",
				Action: runAsForwardServer,
			},
			{
				Name:   "proxy",
				Usage:  "Run as a proxy server, mode proxy",
				Action: runAsProxyServer,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
