package fabric

import (
	"path/filepath"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func GetFabricCMD() *cli.Command {
	return &cli.Command{
		Name:  "fabric",
		Usage: "Operation about fabric network",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a fabric network",
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}

					return Start(repoRoot)
				},
			},
			{
				Name:  "chaincode",
				Usage: "Deploy chaincode on your network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "specify fabric network config.yaml file path",
						Required: true,
					},
				},
				Action: installChaincode,
			},
		},
	}
}

func Start(repoRoot string) error {
	args := []string{filepath.Join(repoRoot, types.FabricScript), "up"}

	return utils.ExecCmd(args, repoRoot)
}

func Stop(repoRoot string) error {
	args := []string{filepath.Join(repoRoot, types.FabricScript), "down"}

	return utils.ExecCmd(args, repoRoot)
}

func installChaincode(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	fabricConfig := ctx.String("config")
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.ChaincodeScript), "install", "-c", fabricConfig)

	return utils.ExecCmd(args, repoRoot)
}
