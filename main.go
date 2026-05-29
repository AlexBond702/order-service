package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/AlexBond702/order-service/cmd"
)

func main() {
	app := &cli.App{
		Name:  "Order-Service",
		Usage: "Order management service",
		Commands: []*cli.Command{
			cmd.WebServer(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "no-json"},
		},
		Version: "2.00.00",
	}
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
