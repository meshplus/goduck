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
								Name:  "type",
								Value: types.TypeBinary,
								Usage: "configuration type, one of binary or docker",
							},
							&cli.StringFlag{
								Name:     "repo",
								Usage:    "Specify BitXHub node repo, e.g. $HOME/.goduck/bitxhub/.bitxhub/node1, only for binary",
								Required: false,
							},
							&cli.IntFlag{
								Name:     "num",
								Usage:    "Specify the sequence num of the node, only for docker",
								Value:    1,
								Required: false,
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
								Name:  "type",
								Value: types.TypeBinary,
								Usage: "configuration type, one of binary or docker",
							},
							&cli.StringFlag{
								Name:     "repo",
								Usage:    "Specify BitXHub node repo, e.g. $HOME/.goduck/bitxhub/.bitxhub/node1, only for binary",
								Required: false,
							},
							&cli.IntFlag{
								Name:     "num",
								Usage:    "Specify the sequence num of the node, only for docker",
								Value:    1,
								Required: false,
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
								Name:  "type",
								Value: types.TypeBinary,
								Usage: "configuration type, one of binary or docker",
							},
							&cli.StringFlag{
								Name:     "repo",
								Usage:    "Specify BitXHub node repo, e.g. $HOME/.goduck/bitxhub/.bitxhub/node1, only for binary",
								Required: false,
							},
							&cli.IntFlag{
								Name:     "num",
								Usage:    "Specify the sequence num of the node, only for docker",
								Value:    1,
								Required: false,
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
						Name:  "type",
						Value: types.TypeBinary,
						Usage: "configuration type, one of binary or docker",
					},
					&cli.StringFlag{
						Name:     "repo",
						Usage:    "Specify Pier repo, e.g. $HOME/.goduck/pier/.pier_ethereum, only for binary",
						Required: false,
					},
					&cli.IntFlag{
						Name:     "num",
						Usage:    "Specify the sequence num of the pier, only for docker",
						Value:    1,
						Required: false,
					},
					&cli.StringFlag{
						Name:  "appchain",
						Usage: "Specify appchain type, one of ethereum or fabric, only for docker",
						Value: types.ChainTypeEther,
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
	typ := ctx.String("type")
	target := ctx.String("repo")
	num := ctx.Int("num")
	n := ctx.Int("n")

	if typ == types.TypeBinary && target == "" {
		return fmt.Errorf("repo can not be nil")
	}

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "bxhAll", typ, path, strconv.Itoa(n), strconv.Itoa(num)}
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
	typ := ctx.String("type")
	target := ctx.String("repo")
	num := ctx.Int("num")
	n := ctx.Int("n")

	if typ == types.TypeBinary && target == "" {
		return fmt.Errorf("repo can not be nil")
	}

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "bxhNet", typ, path, strconv.Itoa(n), strconv.Itoa(num)}
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
	typ := ctx.String("type")
	target := ctx.String("repo")
	num := ctx.Int("num")
	n := ctx.Int("n")

	if typ == types.TypeBinary && target == "" {
		return fmt.Errorf("repo can not be nil")
	}

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "bxhOrder", typ, path, strconv.Itoa(n), strconv.Itoa(num)}
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
	typ := ctx.String("type")
	target := ctx.String("repo")
	num := ctx.Int("num")
	appchain := ctx.String("appchain")
	n := ctx.Int("n")

	if typ == types.TypeBinary && target == "" {
		return fmt.Errorf("repo can not be nil")
	}

	path, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	args := []string{types.LogScript, "pierLog", typ, path, strconv.Itoa(n), strconv.Itoa(num), appchain}
	if err := utils.ExecuteShell(args, repoPath); err != nil {
		return err
	}

	return nil
}
