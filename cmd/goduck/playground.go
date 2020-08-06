package main

import (
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func playgroundCMD() *cli.Command {
	return &cli.Command{
		Name:  "playground",
		Usage: "Set up and experience interchain system smoothly",
		Subcommands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "Start a demo interchain system",
				Action: dockerUp,
			},
			{
				Name:   "stop",
				Usage:  "Stop demo interchain system(container remained)",
				Action: dockerStop,
			},
			{
				Name:   "clean",
				Usage:  "Clean up the demo interchain system",
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

func dockerStop(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "stop"}
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
