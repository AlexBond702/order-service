package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/AlexBond702/order-service/internal/app/builder"
)

const (
	cmdSubscribeDeliveryCalculatedUsage = "Starts the delivery calculation consumer"

	cmdSubscribeDeliveryCalculatedDescription = `
Initializes and starts the delivery calculation consumer.

The process subscribes to the order.delivery.calculated topic,
updates orders in the database and exposes only service HTTP
endpoints (health, metrics and pprof).
`
)

func SubscribeDeliveryCalculated() *cli.Command {
	return &cli.Command{
		Name:            "subscribe-delivery-calculated",
		Aliases:         []string{"delivery"},
		Usage:           cmdSubscribeDeliveryCalculatedUsage,
		Description:     strings.TrimSpace(cmdSubscribeDeliveryCalculatedDescription),
		Action:          cmdSubscribeDeliveryCalculated,
		HideHelpCommand: true,
	}
}

func cmdSubscribeDeliveryCalculated(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()
	app.BuildMonitorOpenTelemetry()
	app.BuildRepoConnPostgres()
	app.BuildRepoOrder()
	app.BuildOrderCalculatedService()
	app.BuildBrokerKafka()
	app.BuildConsumerOrderDeliveryCalculated()

	app.BuildMonitorPrometheus()

	app.BuildProcHttpAdmin()
	app.Run()
	return nil
}
