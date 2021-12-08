package main

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/mq"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

var mqCMD = &cli.Command{
	Name:  "mq",
	Usage: "Mq for hyperchain",
	Subcommands: []*cli.Command{
		{
			Name:  "register",
			Usage: "Register for mq",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "config-path",
					Usage:    "specify hyperchain config path, default: $repo/hyperchain/hpc.toml",
					Required: false,
				},
			},
			ArgsUsage: "command: goduck mq register [address] [queue]",
			Action: func(ctx *cli.Context) error {
				configPath := ctx.String("config-path")

				if configPath == "" {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}
					configPath = filepath.Join(repoRoot, "hyperchain")
					if !fileutil.Exist(configPath) {
						return fmt.Errorf("please `goduck init` first")
					}
				}

				if ctx.NArg() < 2 {
					return fmt.Errorf("missing address or queue")
				}

				return mq.Register(configPath, ctx.Args().Get(0), ctx.Args().Get(1))
			},
		},
		{
			Name:  "unregister",
			Usage: "unregister for mq",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "config-path",
					Usage:    "specify hyperchain config path, default: $repo/hyperchain/hpc.toml",
					Required: false,
				},
			},
			ArgsUsage: "command: goduck mq unregister [exchange] [queue]",
			Action: func(ctx *cli.Context) error {
				configPath := ctx.String("config-path")

				if configPath == "" {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}
					configPath = filepath.Join(repoRoot, "hyperchain")
					if !fileutil.Exist(configPath) {
						return fmt.Errorf("please `goduck init` first")
					}
				}

				if ctx.NArg() < 2 {
					return fmt.Errorf("missing exchange name or queue")
				}

				return mq.Unregister(configPath, ctx.Args().Get(0), ctx.Args().Get(1))
			},
		},
	},
}
