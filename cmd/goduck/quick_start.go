package main

import (
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func quickStartCMD() *cli.Command {
	return &cli.Command{
		Name:  "quick-start",
		Usage: "Set up and experience interchain system smoothly",
		Subcommands: []*cli.Command{
			{
				Name:   "up",
				Usage:  "Start a demo interchain system",
				Action: dockerUp,
			},
			{
				Name:   "transfer",
				Usage:  "Invoke bidirectional transfer in demo interchain system",
				Action: transfer,
			},
		},
	}
}

func dockerUp(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "up"}
	return utils.ExecCmd(args, repoRoot)
}

func transfer(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "transfer"}
	return utils.ExecCmd(args, repoRoot)
}
