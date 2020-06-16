package main

import (
	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/urfave/cli/v2"
)

var pierCMD = &cli.Command{
	Name:  "pier",
	Usage: "Operation about pier",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "Start pier with its appchain up",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
				&cli.StringFlag{
					Name:     "type",
					Usage:    "specify up mode, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
				},
			},
			Action: pierStart,
		},
		{
			Name:  "stop",
			Usage: "Stop pier with its appchain down",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
			},
			Action: pierStop,
		},
	},
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	upType := ctx.String("type")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StartAppchain(repoRoot, chainType, upType)
}

func pierStop(ctx *cli.Context) error {
	chainType := ctx.String("chain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StopAppchain(repoRoot, chainType)
}
