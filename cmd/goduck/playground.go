package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/crypto/asym"
	"github.com/meshplus/goduck/cmd/goduck/bitxhub"
	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func playgroundCMD() *cli.Command {
	return &cli.Command{
		Name:  "playground",
		Usage: "Set up and experience interchain system smoothly",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a demo interchain system",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "prometheus",
						Usage: "Whether to enable Prometheus",
						Value: "false",
					},
					&cli.StringFlag{
						Aliases:  []string{"version", "v"},
						Usage:    "version of the demo interchain system",
						Value:    "v1.6.1",
						Required: false,
					},
				},
				Action: dockerUp,
			},
			{
				Name:   "stop",
				Usage:  "Stop demo interchain system(container remained)",
				Action: dockerStop,
			},
			{
				Name:   "clean",
				Usage:  "Clean up the demo interchain system",
				Action: dockerDown,
			},
			{
				Name:   "transfer",
				Usage:  "Conduct cross-chain transactions",
				Action: transfer,
			},
		},
	}
}

func dockerUp(ctx *cli.Context) error {
	prometheus := ctx.String("prometheus")
	version := ctx.String("version")

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

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport verison")
	}

	//get bitxhub addr
	privKey, err := asym.RestorePrivateKey(filepath.Join(repoRoot, "docker/quick_start/adminKey.json"), "bitxhub")
	if err != nil {
		return err
	}

	address, err := privKey.PublicKey().Address()
	if err != nil {
		return fmt.Errorf("get address error:%w", err)
	}

	// download
	if err := pier.DownloadPierPlugin(repoRoot, types.ChainTypeEther, version, types.LinuxSystem); err != nil {
		return fmt.Errorf("download pier plugin binary error:%w", err)
	}

	if err := bitxhub.DownloadBitxhubConfig(filepath.Join(repoRoot, fmt.Sprintf(types.QuickStartBitxhubCofigPath, version)), version); err != nil {
		return fmt.Errorf("download bitxhub.toml error:%w", err)
	}

	ethVersion := EthConfigMap[version]
	args := []string{types.QuickStartScript, "up"}
	args = append(args, address.String(), prometheus, version, ethVersion)
	return utils.ExecuteShell(args, repoRoot)
}

func dockerDown(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "down"}
	return utils.ExecuteShell(args, repoRoot)
}

func dockerStop(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "stop"}
	return utils.ExecuteShell(args, repoRoot)
}

func transfer(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	args := []string{types.QuickStartScript, "transfer"}
	return utils.ExecuteShell(args, repoRoot)
}
