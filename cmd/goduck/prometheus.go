package main

import (
	"fmt"
	"path/filepath"
	"strings"

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
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.Prometheus), "up")

	for i := 0; i < len(addrs); i++ {
		args = append(args, addrs[i])
	}

	return utils.ExecuteShell(args, repoPath)
}

func stopProm(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.Prometheus), "down")
	return utils.ExecuteShell(args, repoPath)
}

func restartProm(ctx *cli.Context) error {
	addrs := strings.Fields(ctx.String("addrs"))
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.Prometheus), "restart")

	for i := 0; i < len(addrs); i++ {
		args = append(args, addrs[i])
	}

	return utils.ExecuteShell(args, repoPath)
}
