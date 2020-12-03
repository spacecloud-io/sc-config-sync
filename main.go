package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/spaceuptech/helpers"
	"github.com/urfave/cli"

	"github.com/spaceuptech/sc-config-sync/model"
	"github.com/spaceuptech/sc-config-sync/server"
)

const version = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Name = "sc-config-sync"
	app.Usage = "core binary to run space cloud sync config"

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "runs the space cloud instance",
			Action: actionRun,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "log-level",
					EnvVar: "LOG_LEVEL",
					Usage:  "Set the log level [debug | info | error]",
					Value:  helpers.LogLevelInfo,
				},
				cli.StringFlag{
					Name:   "log-format",
					EnvVar: "LOG_FORMAT",
					Usage:  "Set the log format [json | text]",
					Value:  helpers.LogFormatJSON,
				},
				cli.StringFlag{
					Name:   "admin-secret",
					EnvVar: "ADMIN_SECRET",
					Usage:  "secret used for token validation",
				},
				cli.StringFlag{
					Name:   "gateway-addr",
					EnvVar: "GATEWAY_ADDR",
					Value:  "http://localhost:4122",
					Usage:  "address of gateway",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func actionRun(c *cli.Context) error {
	// Load cli flags
	logLevel := c.String("log-level")
	logFormat := c.String("log-format")
	if err := helpers.InitLogger(logLevel, logFormat, false); err != nil {
		return helpers.Logger.LogError(helpers.GetRequestID(context.TODO()), "Unable to initialize loggers", err, nil)
	}

	secret := c.String("admin-secret")
	if secret == "" {
		return helpers.Logger.LogError(helpers.GetRequestID(context.TODO()), "Cannot start sc config sync, admin secret not provided", nil, nil)
	}

	gatewayAddr := c.String("gateway-addr")
	if !strings.HasPrefix(gatewayAddr, "http") {
		return helpers.Logger.LogError(helpers.GetRequestID(context.TODO()), "Gateway address should start from (http) scheme, e.g->  http://gateway.space-cloud.svc.cluster.local:4122", nil, nil)
	}
	model.GatewayAddr = gatewayAddr

	return server.New(secret).Start()
}
