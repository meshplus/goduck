package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func bitxhubCMD() *cli.Command {
	return &cli.Command{
		Name:  "bitxhub",
		Usage: "start or stop BitXHub nodes",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start bitxhub nodes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Value: types.TypeBinary,
						Usage: "configuration type, one of binary or docker",
					},
					&cli.StringFlag{
						Name:  "mode",
						Value: types.SoloMode,
						Usage: "configuration mode, one of solo or cluster",
					},
					&cli.Uint64Flag{
						Name:  "num",
						Value: 4,
						Usage: "node number, only useful in cluster mode, ignored in solo mode",
					},
				},
				Action: startBitXHub,
			},
			{
				Name:   "stop",
				Usage:  "Stop bitxhub nodes",
				Action: stopBitXHub,
			},
			{
				Name:  "config",
				Usage: "Generate configuration for BitXHub nodes",
				Flags: []cli.Flag{
					&cli.Uint64Flag{
						Name:  "num",
						Value: 4,
						Usage: "node number, only useful in cluster mode, ignored in solo mode",
					},
					&cli.StringFlag{
						Name:  "type",
						Value: types.TypeBinary,
						Usage: "configuration type, one of binary or docker",
					},
					&cli.StringFlag{
						Name:  "mode",
						Value: types.ClusterMode,
						Usage: "configuration mode, one of solo or cluster",
					},
					&cli.StringSliceFlag{
						Name:  "ips",
						Usage: "nodes' IP, use 127.0.0.1 for all nodes by default",
					},
					&cli.StringFlag{
						Name:  "target",
						Value: ".",
						Usage: "where to put the generated configuration files",
					},
				},
				Action: generateBitXHubConfig,
			},
		},
	}
}

func stopBitXHub(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoPath, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.PlaygroundScript), "down")

	return utils.ExecCmd(args, repoPath)
}

func startBitXHub(ctx *cli.Context) error {
	num := ctx.Int("num")
	typ := ctx.String("type")
	mode := ctx.String("mode")

	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoPath, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	ips := make([]string, 0)
	err = InitBitXHubConfig(typ, mode, repoPath, num, ips)
	if err != nil {
		return fmt.Errorf("init config error:%w", err)
	}

	if typ == types.TypeBinary {
		err := downloadBinary(repoPath)
		if err != nil {
			return fmt.Errorf("download binary error:%w", err)
		}
	}

	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.PlaygroundScript), "up")
	args = append(args, mode, typ, strconv.Itoa(num))
	return utils.ExecCmd(args, repoPath)
}

func downloadBinary(repoPath string) error {
	root := filepath.Join(repoPath, "bin")
	if !fileutil.Exist(root) {
		err := os.Mkdir(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "bitxhub")) {
			err := download.Download(root, types.BitxhubUrlLinux)
			if err != nil {
				return err
			}
		}
		if !fileutil.Exist(filepath.Join(root, "libwasmer.so")) {
			err := download.Download(root, types.LinuxWasmLibUrl)
			if err != nil {
				return err
			}
		}
	}
	if runtime.GOOS == "darwin" {
		if !fileutil.Exist(filepath.Join(root, "bitxhub")) {
			err := download.Download(root, types.BitxhubUrlMacOS)
			if err != nil {
				return err
			}
		}
		if !fileutil.Exist(filepath.Join(root, "libwasmer.dylib")) {
			err := download.Download(root, types.MacOSWasmLibUrl)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
