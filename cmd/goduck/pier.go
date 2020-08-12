package main

import (
	"fmt"
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
					Name:     "chain-type",
					Usage:    "specify appchain up type, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
				},
				&cli.StringFlag{
					Name:     "pier-type",
					Usage:    "specify pier up type, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
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
				&cli.BoolFlag{
					Name:     "pier-only",
					Usage:    "stop pier only or stop pier with its appchain",
					Required: false,
					Value:    true,
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
	chainUpType := ctx.String("chain-type")
	pierUpType := ctx.String("pier-type")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if pierUpType == types.TypeBinary {
		if !fileutil.Exist(filepath.Join(repoRoot, "bin/pier")) {
			if err := downloadPierBinary(repoRoot); err != nil {
				return fmt.Errorf("download pier binary error:%w", err)
			}
		}
	}

	// start pier with specific appchain
	return pier.StartPier(repoRoot, chainType, chainUpType, pierUpType)
}

func pierStop(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	isPierOnly := ctx.Bool("pier-only")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StopPier(repoRoot, chainType, isPierOnly)
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

func downloadPierBinary(repoPath string) error {
	root := filepath.Join(repoPath, "bin")
	if !fileutil.Exist(root) {
		err := os.Mkdir(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "pier")) {
			err := download.Download(root, types.PierUrlLinux)
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
			err := download.Download(root, types.PierUrlMacOS)
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
