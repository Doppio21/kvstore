package storeservice

import (
	"kvstore/internal/common"
	"kvstore/internal/common/grpcserver"
	"kvstore/internal/storeservice/manager"
	"kvstore/internal/storeservice/store/badgerkv"

	"github.com/urfave/cli/v2"
)

var CLICommand = &cli.Command{
	Name:  "store",
	Usage: "Key-Value storage service",
	Subcommands: []*cli.Command{
		{
			Name: "run",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "address",
					Value: "localhost:20001",
				},
			},
			Action: runStore,
		},
	},
}

func runStore(ctx *cli.Context) error {
	store := New(
		Config{
			Server: grpcserver.Config{
				Address: ctx.String("address"),
			},
			Manager: manager.Config{
				UseCompression: false,
			},
			Store: badgerkv.Config{
				InMem: true,
				Root:  "/tmp/store-temp",
			},
		},
		Dependencies{
			Registry: common.NewPrometheusRegistry(),
			Log:      common.NewLogger(),
		},
	)

	return store.Run(ctx.Context)
}
