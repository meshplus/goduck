package main

import (
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func quickStartCMD() *cli.Command {
	return &cli.Command{
		Name:   "quickstart",
		Usage:  "Set up and experience interchain system smoothly",
		Action: dockerUp,
		Subcommands: []*cli.Command{
			{
				Name:   "stop",
				Usage:  "Stop demo interchain system",
				Action: dockerDown,
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
	return utils.ExecuteShell(args, repoRoot)
}

func dockerDown(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "down"}
	return utils.ExecuteShell(args, repoRoot)
}

func transfer(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "transfer"}
	return utils.ExecuteShell(args, repoRoot)
}
