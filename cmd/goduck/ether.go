package main

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/ethereum"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

var EthConfigMap = map[string]string{
	"v1.6.1":  "1.2.0",
	"v1.6.5":  "1.2.0",
	"v1.7.0":  "1.2.0",
	"v1.8.0":  "1.2.0",
	"v1.9.0":  "1.2.0",
	"v1.10.0": "1.2.0",
	"v1.11.0": "1.3.0",
	"v1.11.1": "1.3.0",
}

func etherCMD() *cli.Command {
	return &cli.Command{
		Name:  "ether",
		Usage: "Operation about ethereum chain",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a ethereum chain",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Usage:    "specify ethereum up type, docker(default) or binary",
						Required: false,
						Value:    types.TypeDocker,
					},
					&cli.StringFlag{
						Name:     "bxh-version",
						Usage:    "specify bitxhub version (Only for docker. The launched ethereum private chain in docker mod has already deployed the cross-chain contract required for the corresponding version of BitXHub)",
						Required: false,
						Value:    "v1.6.5",
					},
				},
				Action: startEther,
			},
			{
				Name:   "stop",
				Usage:  "Stop ethereum chain",
				Action: stopEther,
			},
			{
				Name:   "clean",
				Usage:  "Clean ethereum chain",
				Action: cleanEther,
			},
			ethereum.ContractCMD,
		},
	}
}

func startEther(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	typ := ctx.String("type")
	version := EthConfigMap[ctx.String("bxh-version")]

	if typ == types.TypeBinary {
		err := ethereum.DownloadGethBinary(repoRoot)
		if err != nil {
			return fmt.Errorf("download binary error:%w", err)
		}
	}

	if err := StartEthereum(repoRoot, typ, version); err != nil {
		return err
	}

	fmt.Printf("start ethereum private chain with data directory in %s/ethereum/datadir.\n", repoRoot)
	return nil
}

func stopEther(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if err := StopEthereum(repoRoot); err != nil {
		return err
	}
	return nil
}

func cleanEther(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if err := CleanEthereum(repoRoot); err != nil {
		return err
	}
	return nil
}

func StopEthereum(repoPath string) error {
	if !fileutil.Exist(filepath.Join(repoPath, types.EthereumScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return utils.ExecuteShell([]string{types.EthereumScript, "down"}, repoPath)
}

func CleanEthereum(repoPath string) error {
	if !fileutil.Exist(filepath.Join(repoPath, types.EthereumScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return utils.ExecuteShell([]string{types.EthereumScript, "clean"}, repoPath)
}

func StartEthereum(repoPath, mod, version string) error {
	if !fileutil.Exist(filepath.Join(repoPath, types.EthereumScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return utils.ExecuteShell([]string{types.EthereumScript, mod, version}, repoPath)
}
