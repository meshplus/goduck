package main

import (
	"github.com/meshplus/goduck/cmd/goduck/fabric"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

func fabricCMD() *cli.Command {
	return &cli.Command{
		Name:  "fabric",
		Usage: "Operation about fabric network",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a fabric network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "crypto-config",
						Usage:    "specify fabric network crypto-config directory path",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}

					cryptoConfigPath := ctx.String("crypto-config")

					return fabric.Start(repoRoot, cryptoConfigPath)
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a fabric network",
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}

					return fabric.Stop(repoRoot)
				},
			},
			{
				Name:  "clean",
				Usage: "Clean a fabric network",
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}

					return fabric.Clean(repoRoot)
				},
			},
			fabric.ContractCMD,
		},
	}
}
