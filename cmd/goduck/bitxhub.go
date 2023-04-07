package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/meshplus/goduck/cmd/goduck/bitxhub"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

type Release struct {
	Bitxhub []string `json:"bitxhub"`
	Pier    []string `json:"pier"`
}

var bxhConfigMap = map[string]string{
	"v1.6.1":  "v1.6.1",
	"v1.6.5":  "v1.6.5", // same to 1.6.1
	"v1.7.0":  "v1.7.0",
	"v1.8.0":  "v1.8.0",
	"v1.9.0":  "v1.9.0",
	"v1.11.0": "v1.11.0",
	"v1.11.1": "v1.11.0",
	"v1.11.3": "v1.11.3",
	"v1.23.0": "v1.23.0",
	"v2.8.0":  "v2.8.0",
	"v2.10.0": "v2.10.0",
}

func bitxhubCMD() *cli.Command {
	return &cli.Command{
		Name:  "bitxhub",
		Usage: "Start or stop BitXHub nodes",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start BitXHub nodes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Value: types.TypeBinary,
						Usage: "configuration type, one of binary or docker",
					},
					&cli.StringFlag{
						Name:  "target",
						Usage: "Specify the directory to where to put the generated configuration files, default: $repo/bitxhub/.bitxhub/",
					},
					&cli.StringFlag{
						Name:  "configPath",
						Usage: "Specify the configuration file path for the configuration to be modified, default: $repo/bxh_config/$version/bxh_modify_config.toml",
					},
					&cli.StringFlag{
						Aliases: []string{"version", "v"},
						Value:   "v2.10.0",
						Usage:   "BitXHub version",
					},
				},
				Action: startBitXHub,
			},
			{
				Name:   "stop",
				Usage:  "Stop BitXHub nodes",
				Action: stopBitXHub,
			},
			{
				Name:   "clean",
				Usage:  "Clean BitXHub nodes",
				Action: cleanBitXHub,
			},
			{
				Name:  "config",
				Usage: "Generate configuration for BitXHub nodes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "target",
						Usage: "Specify the directory to where to put the generated configuration files, default: $repo/bitxhub/.bitxhub/",
					},
					&cli.StringFlag{
						Name:  "configPath",
						Usage: "Specify the configuration file path for the configuration to be modified, default: $repo/bxh_config/$version/bxh_modify_config.toml",
					},
					&cli.StringFlag{
						Aliases: []string{"version", "v"},
						Value:   "v2.10.0",
						Usage:   "BitXHub version",
					},
				},
				Action: generateBitXHubConfig,
			},
		},
	}
}

func stopBitXHub(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.PlaygroundScript), "down")

	return utils.ExecuteShell(args, repoRoot)
}

func startBitXHub(ctx *cli.Context) error {
	typ := ctx.String("type")
	configPath := ctx.String("configPath")
	target := ctx.String("target")
	version := ctx.String("version")

	// 1. check goduck init
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	// 2. check version
	data, err := ioutil.ReadFile(filepath.Join(repoRoot, "release.json"))
	if err != nil {
		return err
	}

	var release *Release
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport BitXHub verison")
	}

	// 3. get args and bin
	if configPath == "" {
		configPath = filepath.Join(repoRoot, fmt.Sprintf("bxh_config/%s/%s", bxhConfigMap[version], types.BxhModifyConfig))
	} else {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("get absolute config path: %w", err)
		}
	}

	if target == "" {
		target = filepath.Join(repoRoot, fmt.Sprintf("bitxhub/.bitxhub"))
	} else {
		target, err = filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("get absolute target path: %w", err)
		}
	}

	if typ == types.TypeBinary {
		err := bitxhub.DownloadBitxhubBinary(repoRoot, version, runtime.GOOS)
		if err != nil {
			return fmt.Errorf("download binary error:%w", err)
		}
	}

	// 4. execute
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.PlaygroundScript), "up")
	args = append(args, version, typ, configPath, target)
	return utils.ExecuteShell(args, repoRoot)
}

func AdjustVersion(version string, release []string) bool {
	for _, bxhRelease := range release {
		if version == bxhRelease {
			return true
		}
	}
	return false
}

func cleanBitXHub(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.PlaygroundScript), "clean")
	return utils.ExecuteShell(args, repoRoot)
}

func generateBitXHubConfig(ctx *cli.Context) error {
	target := ctx.String("target")
	configPath := ctx.String("configPath")
	version := ctx.String("version")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.PlaygroundScript)) {
		return fmt.Errorf("please `goduck init` first")
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
		return fmt.Errorf("unsupport BitXHub verison")
	}

	if target == "" {
		target = filepath.Join(repoRoot, fmt.Sprintf("bitxhub/.bitxhub"))
	} else {
		target, err = filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("get absolute target path: %w", err)
		}
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	if configPath == "" {
		configPath = filepath.Join(repoRoot, fmt.Sprintf("bxh_config/%s/%s", bxhConfigMap[version], types.BxhModifyConfig))
	} else {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("get absolute config path: %w", err)
		}
	}

	err = bitxhub.DownloadBitxhubBinary(repoRoot, version, runtime.GOOS)
	if err != nil {
		return fmt.Errorf("download binary error:%w", err)
	}
	err = bitxhub.ExtractBitxhubBinary(repoRoot, version, runtime.GOOS)
	if err != nil {
		return fmt.Errorf("extract binary error:%w", err)
	}
	binPath := filepath.Join(repoRoot, fmt.Sprintf("bin/%s", fmt.Sprintf("bitxhub_%s_%s", runtime.GOOS, version)))
	fmt.Println(binPath)

	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.BxhConfigRepo, bxhConfigMap[version], types.BitxhubConfigScript))
	args = append(args, "-t", target, "-b", binPath, "-p", configPath)
	return utils.ExecuteShell(args, repoRoot)
}
