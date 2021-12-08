package main

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func infoCMD() *cli.Command {
	return &cli.Command{
		Name:  "info",
		Usage: "Show basic info about interchain system",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all basic info about interchain system",
				Action: showInfo,
			},
			{
				Name:  "bitxhub",
				Usage: "List all basic info about BitXHub",
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}
					if !fileutil.Exist(repoRoot) {
						return fmt.Errorf("please `goduck init` first")
					}
					return showBxhInfo(repoRoot)
				},
			},
			{
				Name:  "pier",
				Usage: "List all basic info about piers",
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}
					if !fileutil.Exist(repoRoot) {
						return fmt.Errorf("please `goduck init` first")
					}
					return showPierInfo(repoRoot)
				},
			},
		},
	}
}

func showInfo(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	if err := showBxhInfo(repoRoot); err != nil {
		return err
	}

	if err := showPierInfo(repoRoot); err != nil {
		return err
	}
	return nil
}

func showPierInfo(repoPath string) error {
	args := []string{types.InfoScript, "pier"}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}

func showBxhInfo(repoPath string) error {
	args := []string{types.InfoScript, "bitxhub"}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}
