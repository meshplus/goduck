package main

import (
	"github.com/meshplus/goduck/internal/repo"

	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/urfave/cli/v2"
)

var pierCMD = &cli.Command{
	Name:  "pier",
	Usage: "operation about pier",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "start pier with its appchain up",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type",
					Required: false,
					Value:    "ethereum",
				},
				&cli.StringFlag{
					Name:     "type",
					Usage:    "specify up mode",
					Required: false,
					Value:    "binary",
				},
			},
			Action: pierStart,
		},
		{
			Name:  "stop",
			Usage: "stop pier with its appchain down",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type",
					Required: false,
					Value:    "ethereum",
				},
			},
			Action: pierStop,
		},
	},
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	mode := ctx.String("type")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StartAppchain(repoRoot, chainType, mode)
}

func pierStop(ctx *cli.Context) error {
	chainType := ctx.String("chain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StopAppchain(repoRoot, chainType)
}
