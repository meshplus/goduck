package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type Release struct {
	Bitxhub []string `json:"bitxhub"`
	Pier    []string `json:"pier"`
}

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
					&cli.StringFlag{
						Name:  "version,v",
						Value: "v1.1.0-rc1",
						Usage: "bitxhub version",
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
				Name:   "clean",
				Usage:  "Clean bitxhub nodes",
				Action: cleanBitXHub,
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
					&cli.StringFlag{
						Name:  "version,v",
						Value: "v1.1.0-rc1",
						Usage: "bitxhub version",
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

	return utils.ExecuteShell(args, repoPath)
}

func startBitXHub(ctx *cli.Context) error {
	num := ctx.Int("num")
	typ := ctx.String("type")
	mode := ctx.String("mode")
	version := ctx.String("version")

	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoPath, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	data, err := ioutil.ReadFile(filepath.Join(repoPath, "release.json"))
	if err != nil {
		return err
	}

	var release *Release
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport bitxhub verison")
	}

	bxhConfig := filepath.Join(repoPath, "bitxhub")
	ips := make([]string, 0)
	err = InitBitXHubConfig(typ, mode, bxhConfig, num, ips, version)
	if err != nil {
		return fmt.Errorf("init config error:%w", err)
	}

	if typ == types.TypeBinary {
		err := downloadBinary(repoPath, version)
		if err != nil {
			return fmt.Errorf("download binary error:%w", err)
		}
	}

	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.PlaygroundScript), "up")
	args = append(args, version, mode, typ, strconv.Itoa(num))
	return utils.ExecuteShell(args, repoPath)
}

func AdjustVersion(version string, release []string) bool {
	for _, bxhRelease := range release {
		if version ==  bxhRelease {
			return true
		}
	}
	return false
}

func cleanBitXHub(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoPath, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, types.PlaygroundScript), "clean")
	return utils.ExecuteShell(args, repoPath)
}

func downloadBinary(repoPath string, version string) error {
	path := fmt.Sprintf("bitxhub_%s", version)
	root := filepath.Join(repoPath, "bin", path)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "bitxhub")) {
			url := fmt.Sprintf(types.BitxhubUrlLinux, version, version)
			err := download.Download(root, url)
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
			url := fmt.Sprintf(types.BitxhubUrlMacOS, version, version)
			err := download.Download(root, url)
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
