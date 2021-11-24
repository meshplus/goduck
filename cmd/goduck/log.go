package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func logCMD() *cli.Command {
	return &cli.Command{
		Name:  "log",
		Usage: "Print log of BitXHub or Pier",
		Subcommands: []*cli.Command{
			{
				Name:  "bitxhub",
				Usage: "Print log of BitXHub",
				Subcommands: []*cli.Command{
					{
						Name:  "all",
						Usage: "Print all log of BitXHub node",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "repo",
								Usage:    "Specify BitXHub node repo, e.g. $HOME/.goduck/bitxhub/.bitxhub/node1",
								Required: true,
							},
							&cli.IntFlag{
								Name:     "n",
								Usage:    "Specify the number of lines to print the latest log, -1 indicates that all logs are printed",
								Value:    -1,
								Required: false,
							},
						},
						Action: bxhAll,
					},
					{
						Name:  "net",
						Usage: "Print all log of BitXHub node",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "repo",
								Usage:    "Specify BitXHub node repo, e.g. $HOME/.goduck/bitxhub/.bitxhub/node1",
								Required: true,
							},
							&cli.IntFlag{
								Name:     "n",
								Usage:    "Specify how many lines of the latest log to filter from, -1 indicates all",
								Value:    -1,
								Required: false,
							},
						},
						Action: bxhNet,
					},
					{
						Name:  "order",
						Usage: "Print all log of BitXHub node",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "repo",
								Usage:    "Specify BitXHub node repo, e.g. $HOME/.goduck/bitxhub/.bitxhub/node1",
								Required: true,
							},
							&cli.IntFlag{
								Name:     "n",
								Usage:    "Specify how many lines of the latest log to filter from, -1 indicates all",
								Value:    -1,
								Required: false,
							},
						},
						Action: bxhOrder,
					},
				},
			},
			{
				Name:  "pier",
				Usage: "Print log of Pier",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "repo",
						Usage:    "Specify Pier repo, e.g. $HOME/.goduck/pier/.pier_ethereum",
						Required: true,
					},
					&cli.IntFlag{
						Name:     "n",
						Usage:    "Specify the number of lines to print the latest log, -1 indicates that all logs are printed",
						Value:    -1,
						Required: false,
					},
				},
				Action: pierLog,
			},
		},
	}
}

func bxhAll(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	target := ctx.String("repo")
	n := ctx.Int("n")

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "bxhAll", path, strconv.Itoa(n)}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}

func bxhNet(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	target := ctx.String("repo")
	n := ctx.Int("n")

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "bxhNet", path, strconv.Itoa(n)}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}

func bxhOrder(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	target := ctx.String("repo")
	n := ctx.Int("n")

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "bxhOrder", path, strconv.Itoa(n)}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}

func pierLog(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	target := ctx.String("repo")
	n := ctx.Int("n")

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "pierLog", path, strconv.Itoa(n)}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}
