package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func prometheusCMD() *cli.Command {
	return &cli.Command{
		Name:  "prometheus",
		Usage: "Start or stop prometheus",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start prometheus to monitoring BitXHub",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "addrs",
						Usage:    "address of BitXHub nodes",
						Value:    "host.docker.internal:40011 host.docker.internal:40012 host.docker.internal:40013 host.docker.internal:40014",
						Required: false,
					},
				},
				Action: startProm,
			},
			{
				Name:   "stop",
				Usage:  "Stop prometheus",
				Action: stopProm,
			},
			{
				Name:  "restart",
				Usage: "Restart prometheus",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "addrs",
						Usage:    "address of BitXHub nodes",
						Value:    "host.docker.internal:40011 host.docker.internal:40012 host.docker.internal:40013 host.docker.internal:40014",
						Required: false,
					},
				},
				Action: restartProm,
			},
		},
	}
}

func startProm(ctx *cli.Context) error {
	addrs := strings.Fields(ctx.String("addrs"))
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.Prometheus), "up")

	for i := 0; i < len(addrs); i++ {
		args = append(args, addrs[i])
	}

	return utils.ExecuteShell(args, repoRoot)
}

func stopProm(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.Prometheus), "down")
	return utils.ExecuteShell(args, repoRoot)
}

func restartProm(ctx *cli.Context) error {
	addrs := strings.Fields(ctx.String("addrs"))
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.Prometheus), "restart")

	for i := 0; i < len(addrs); i++ {
		args = append(args, addrs[i])
	}

	return utils.ExecuteShell(args, repoRoot)
}
