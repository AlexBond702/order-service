package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/AlexBond702/order-service/cmd"
	"github.com/AlexBond702/order-service/internal/app/constant"
	msentry "github.com/AlexBond702/order-service/internal/app/processor/monitor/sentry"
)

func main() {
	defer msentry.Flush()
	app := &cli.App{
		Name:  constant.AppName,
		Usage: "Order management service",
		Commands: []*cli.Command{
			cmd.WebServer(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "no-json"},
		},
		Version: constant.Version,
	}
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
