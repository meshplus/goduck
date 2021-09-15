package main

import (
	"fmt"

	"github.com/meshplus/goduck/cmd/goduck/mq"
	"github.com/urfave/cli/v2"
)

var mqCMD = &cli.Command{
	Name:  "mq",
	Usage: "Mq for hyperchain",
	Subcommands: []*cli.Command{
		{
			Name:      "register",
			Usage:     "Register for mq",
			ArgsUsage: "command: goduck mq register [address] [queue]",
			Action: func(ctx *cli.Context) error {

				if ctx.NArg() < 2 {
					return fmt.Errorf("missing address or queue")
				}

				return mq.Register(ctx.Args().Get(0), ctx.Args().Get(1))
			},
		},
		{
			Name:      "unregister",
			Usage:     "unregister for mq",
			ArgsUsage: "command: goduck mq unregister [exchange] [queue]",
			Action: func(ctx *cli.Context) error {

				if ctx.NArg() < 2 {
					return fmt.Errorf("missing exchange name or queue")
				}

				return mq.Unregister(ctx.Args().Get(0), ctx.Args().Get(1))
			},
		},
	},
}
