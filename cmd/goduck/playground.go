package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

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
				Name:  "start",
				Usage: "Start a demo interchain system",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Aliases:  []string{"version", "v"},
						Usage:    "version of the demo interchain system",
						Value:    "v1.6.0",
						Required: false,
					},
				},
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
	version := ctx.String("version")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filepath.Join(repoRoot, "release.json"))
	if err != nil {
		return err
	}

	var release *Release
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport verison")
	}

	args := []string{types.QuickStartScript, "up"}
	args = append(args, version)
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
