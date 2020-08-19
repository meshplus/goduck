package fabric

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func GetFabricCMD() *cli.Command {
	return &cli.Command{
		Name:  "fabric",
		Usage: "Operation about fabric network",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a fabric network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "crypto-config",
						Usage:    "specify fabric network crypto-config directory path",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}

					cryptoConfigPath := ctx.String("crypto-config")

					return Start(repoRoot, cryptoConfigPath)
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a fabric network",
				Action: func(ctx *cli.Context) error {
					repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
					if err != nil {
						return err
					}

					return Stop(repoRoot)
				},
			},
			{
				Name:  "chaincode",
				Usage: "Deploy chaincode on your network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Usage: "specify fabric network config.yaml file path",
					},
					&cli.StringFlag{
						Name:     "code",
						Usage:    "specify chain code path, default(our interchain chaincode)",
						Required: false,
					},
				},
				Action: installChaincode,
			},
			{
				Name:      "invoke",
				Usage:     "Invoke fabric chaincode",
				ArgsUsage: "command: goduck fabric invoke [chaincode_id] [function] [args(optional)]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config_path",
						Usage:    "the path of fabric config, default(our fabric config)",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					return invokeChaincode(ctx)
				},
			},
		},
	}
}

func Start(repoRoot, cryptoConfigPath string) error {
	var (
		cryptoPath string
		err        error
	)
	if cryptoConfigPath != "" {
		cryptoPath, err = filepath.Abs(cryptoConfigPath)
		if err != nil {
			return err
		}
	}

	args := []string{filepath.Join(repoRoot, types.FabricScript), "up", cryptoPath}

	return utils.ExecuteShell(args, repoRoot)
}

func Stop(repoRoot string) error {
	args := []string{filepath.Join(repoRoot, types.FabricScript), "down"}

	return utils.ExecuteShell(args, repoRoot)
}

func installChaincode(ctx *cli.Context) error {
	codePath := ctx.String("code")
	absPath := ""
	if codePath != "" {
		absP, err := filepath.Abs(codePath)
		if err != nil {
			return err
		}
		absPath = absP
	}

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	fabricConfig := ctx.String("config")
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.ChaincodeScript), "install", "-c", fabricConfig, "-g", absPath)

	return utils.ExecuteShell(args, repoRoot)
}

func invokeChaincode(ctx *cli.Context) error {
	configPath := ctx.String("config_path")

	if configPath == "" {
		repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
		if err != nil {
			return err
		}
		configPath = filepath.Join(repoRoot, "config.yaml")
	}

	args := ctx.Args().Slice()
	if len(args) < 2 {
		return fmt.Errorf("args must be (chaincode_id function args[optional])")
	}

	if len(args) == 2 {
		args = append(args, "")
	}

	return Invoke(configPath, args[0], args[1], args[2])
}
