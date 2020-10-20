package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/urfave/cli/v2"
)

var pierCMD = &cli.Command{
	Name:  "pier",
	Usage: "Operation about pier",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "Start pier with its appchain up",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
				&cli.StringFlag{
					Name:     "cryptoPath",
					Usage:    "path of crypto-config, only useful for fabric chain, e.g $HOME/crypto-config",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "pier-type",
					Usage:    "specify pier up type, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
				},
				&cli.StringFlag{
					Name:  "version,v",
					Value: "v1.1.0-rc1",
					Usage: "pier version",
				},
				&cli.StringFlag{
					Name:  "pprof-port",
					Value: "44550",
					Usage: "pier pprof port, only useful for binary",
				},
			},
			Action: pierStart,
		},
		{
			Name:  "stop",
			Usage: "Stop pier with its appchain down",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
			},
			Action: pierStop,
		},
		{
			Name:  "clean",
			Usage: "Clean pier with its appchain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
				&cli.BoolFlag{
					Name:     "pier-only",
					Usage:    "clean pier only or clean pier with its appchain",
					Required: false,
					Value:    true,
				},
			},
			Action: pierClean,
		},
		{
			Name:  "config",
			Usage: "Generate configuration for Pier",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "mode",
					Value: types.PierModeDirect,
					Usage: "configuration mode, one of direct or relay",
				},
				&cli.StringFlag{
					Name:  "type",
					Value: types.TypeBinary,
					Usage: "configuration type, one of binary or docker",
				},
				&cli.StringFlag{
					Name:  "bitxhub",
					Usage: "BitXHub's address, only useful in relay mode",
				},
				&cli.StringSliceFlag{
					Name:  "validators",
					Usage: "BitXHub's validators, only useful in relay mode",
				},
				&cli.IntFlag{
					Name:  "port",
					Value: 5001,
					Usage: "pier's port, only useful in direct mode",
				},
				&cli.StringSliceFlag{
					Name:  "peers",
					Usage: "peers' address, only useful in direct mode",
				},
				&cli.StringFlag{
					Name:  "appchain-type",
					Value: "ethereum",
					Usage: "appchain type, one of ethereum or fabric",
				},
				&cli.StringFlag{
					Name:  "appchain-IP",
					Value: "127.0.0.1",
					Usage: "appchain IP address",
				},
				&cli.StringFlag{
					Name:  "target",
					Value: ".",
					Usage: "where to put the generated configuration files",
				},
				&cli.IntFlag{
					Name:  "ID",
					Value: 0,
					Usage: "specify Pier's ID which is in [0,9], cannot exist 2 or more Piers with same ID in one OS",
				},
			},
			Action: generatePierConfig,
		},
	},
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	cryptoPath := ctx.String("cryptoPath")
	pierUpType := ctx.String("pier-type")
	version := ctx.String("version")
	port := ctx.String("pprof-port")

	if chainType == "fabric" && cryptoPath == "" {
		return fmt.Errorf("start fabric pier need crypto-config path")
	}

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

	if !AdjustVersion(version, release.Pier) {
		return fmt.Errorf("unsupport pier verison")
	}

	if pierUpType == types.TypeBinary {
		if !fileutil.Exist(filepath.Join(repoRoot, fmt.Sprintf("bin/pier_%s_%s/pier", runtime.GOOS, version))) {
			if err := downloadPierBinary(repoRoot, version); err != nil {
				return fmt.Errorf("download pier binary error:%w", err)
			}
		}
	}

	return pier.StartPier(repoRoot, chainType, cryptoPath, pierUpType, version, port)
}

func pierStop(ctx *cli.Context) error {
	chainType := ctx.String("chain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StopPier(repoRoot, chainType)
}

func pierClean(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	isPierOnly := ctx.Bool("pier-only")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.CleanPier(repoRoot, chainType, isPierOnly)
}

func downloadPierBinary(repoPath string, version string) error {
	path := fmt.Sprintf("pier_%s_%s", runtime.GOOS, version)
	root := filepath.Join(repoPath, "bin", path)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "pier")) {
			url := fmt.Sprintf(types.PierUrlLinux, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			if !fileutil.Exist(filepath.Join(root, "libwasmer.so")) {
				err := download.Download(root, types.LinuxWasmLibUrl)
				if err != nil {
					return err
				}
			}
		}
	}
	if runtime.GOOS == "darwin" {
		if !fileutil.Exist(filepath.Join(root, "pier")) {
			url := fmt.Sprintf(types.PierUrlMacOS, version, version)
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
