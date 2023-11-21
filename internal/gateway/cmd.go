package gateway

import (
	"kvstore/internal/common"
	"kvstore/internal/common/grpcclient"
	"kvstore/internal/gateway/server"

	"github.com/urfave/cli/v2"
)

var CLICommand = &cli.Command{
	Name:  "gw",
	Usage: "Key-Value storage gateway",
	Subcommands: []*cli.Command{
		{
			Name: "run",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "address",
					Value: "localhost:10001",
				},
			},
			Action: runGW,
		},
	},
}

func runGW(ctx *cli.Context) error {
	gw := New(
		Config{
			Client: grpcclient.Config{
				Address: "localhost:20001",
			},
			Server: server.Config{
				Address: ctx.String("address"),
			},
		},
		Dependencies{
			Registry: common.NewPrometheusRegistry(),
			Log:      common.NewLogger(),
		},
	)

	return gw.Run(ctx.Context)
}
