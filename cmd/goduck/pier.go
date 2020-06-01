package main

import (
	"fmt"

	"github.com/meshplus/goduck/cmd/goduck/ethereum"
	"github.com/meshplus/goduck/cmd/goduck/fabric"
	"github.com/meshplus/goduck/internal/repo"
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
	},
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	mode := ctx.String("type")

	switch mode {
	case "binary":
		switch chainType {
		case "fabric":
			return fmt.Errorf("fabric is not supported to start up with binary")
		case "ethereum":
			return startEthereumWithBinary(ctx)
		default:
			return fmt.Errorf("chain type %s is not supported", chainType)
		}
	case "docker":
		switch chainType {
		case "fabric":
			return startFabricWithDocker(ctx)
		case "ethereum":
			return startEthereumWithDocker(ctx)
		default:
			return fmt.Errorf("chain type %s is not supported", chainType)
		}
	default:
		return fmt.Errorf("start up mode %s is not supported", chainType)
	}
}

func startEthereumWithBinary(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return ethereum.StartEthereum(repoRoot, "binary")
}

func startEthereumWithDocker(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return ethereum.StartEthereum(repoRoot, "docker")
}

func startFabricWithDocker(ctx *cli.Context) error {
	return fabric.Start(ctx)
}
