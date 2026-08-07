package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/AlexBond702/order-service/internal/app/builder"
)

const (
	cmdWebServerUsage = "Starts the web (REST) server"

	cmdWebServerDescription = `
Initializes and starts web-server, that listens specified port
for incoming REST requests.
`
)

func WebServer() *cli.Command {
	return &cli.Command{
		Name:            "web-server",
		Aliases:         []string{"web", "http"},
		Usage:           cmdWebServerUsage,
		Description:     strings.TrimSpace(cmdWebServerDescription),
		Action:          cmdWebServer,
		HideHelpCommand: true,
	}
}

func cmdWebServer(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()
	app.BuildMonitorOpenTelemetry()
	app.BuildRepoConnPostgres()
	app.BuildRepoOrder()
	app.BuildCatalogClient()
	app.BuildBrokerKafka()
	app.BuildBusOrderCreated()
	app.BuildModuleOrder()
	app.BuildHandlerHttpOrder()

	app.BuildMonitorPrometheus()

	app.BuildProcHttp()
	app.Run()
	return nil
}
