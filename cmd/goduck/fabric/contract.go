package fabric

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

var ContractCMD = &cli.Command{
	Name:  "contract",
	Usage: "Interact with fabric contract about invoke and querying upgrading",
	Subcommands: cli.Commands{
		{
			Name:  "chaincode",
			Usage: "Deploy default chaincode on fabric network",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "config-path",
					Usage:    "specify fabric network config.yaml file path, default(our fabric config)",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "code-path",
					Usage:    "specify chain code path, default(our interchain chaincode)",
					Required: false,
				},
			},
			Action: installChaincode,
		},
		{
			Name:  "deploy",
			Usage: "Deploy fabric chaincode",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "config-path",
					Usage:    "specify fabric network config.yaml file path, default(our fabric config)",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "gopath",
					Usage:    "specify GOPATH for chaincode install command. If not set, GOPATH is taken from the environment",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "ccp",
					Usage:    "specify chaincode path",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "ccid",
					Usage:    "specify chaincode id",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "mspid",
					Usage:    "specify msp id",
					Required: false,
					Value:    "Org2MSP",
				},
				&cli.StringFlag{
					Name:     "version",
					Usage:    "specify chaincode version. This version is a customized contract version",
					Required: true,
				},
			},
			Action: deployChaincode,
		},
		{
			Name:      "invoke",
			Usage:     "Invoke fabric chaincode",
			ArgsUsage: "command: goduck fabric contract invoke [chaincode_id] [function] [args(optional)]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "config-path",
					Usage:    "specify fabric network config.yaml file path, default(our fabric config)",
					Required: false,
				},
			},
			Action: invokeChaincode,
		},
		{
			Name:      "query",
			Usage:     "Query fabric chaincode",
			ArgsUsage: "command: goduck fabric contract query [chaincode_id] [function] [args(optional)]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "config-path",
					Usage:    "specify fabric network config.yaml file path, default(our fabric config)",
					Required: false,
				},
			},
			Action: queryChaincode,
		},
		{
			Name:  "download",
			Usage: "download the default cross-chain contract on fabric chain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "version",
					Usage:    "specify bitxhub version",
					Value:    "v1.6.2",
					Required: false,
				},
			},
			Action: downloadContract,
		},
	},
}

func installChaincode(ctx *cli.Context) error {
	configPath := ctx.String("config-path")
	if configPath == "" {
		repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
		if err != nil {
			return err
		}
		configPath = filepath.Join(repoRoot, types.ChainTypeFabric, "config.yaml")
	}

	codePath := ctx.String("code-path")
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

	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.ChaincodeScript), "install", "-c", configPath, "-g", absPath)

	return utils.ExecuteShell(args, repoRoot)
}

func deployChaincode(ctx *cli.Context) error {
	configPath := ctx.String("config-path")
	if configPath == "" {
		repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
		if err != nil {
			return err
		}
		configPath = filepath.Join(repoRoot, types.ChainTypeFabric, "config.yaml")
	}

	gopath := ctx.String("gopath")
	ccp := ctx.String("ccp")
	ccid := ctx.String("ccid")
	mspid := ctx.String("mspid")
	version := ctx.String("version")
	return Deploy(configPath, gopath, ccp, ccid, mspid, version)
}

func invokeChaincode(ctx *cli.Context) error {
	configPath := ctx.String("config-path")
	if configPath == "" {
		repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
		if err != nil {
			return err
		}
		configPath = filepath.Join(repoRoot, types.ChainTypeFabric, "config.yaml")
	}

	args := ctx.Args().Slice()
	if len(args) < 2 {
		return fmt.Errorf("args must be (chaincode_id function args[optional])")
	}

	return Invoke(configPath, args[0], args[1], args[2], true)
}

func queryChaincode(ctx *cli.Context) error {
	configPath := ctx.String("config-path")
	if configPath == "" {
		repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
		if err != nil {
			return err
		}
		configPath = filepath.Join(repoRoot, types.ChainTypeFabric, "config.yaml")
	}

	args := ctx.Args().Slice()
	if len(args) < 2 {
		return fmt.Errorf("args must be (chaincode_id function args[optional])")
	}

	return Invoke(configPath, args[0], args[1], args[2], false)
}

func downloadContract(ctx *cli.Context) error {
	version := ctx.String("version")

	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(repoPath) {
		return fmt.Errorf("please `goduck init` first")
	}

	path := filepath.Join(repoPath, types.ChainTypeFabric, "contract", version, "contracts.zip")
	if !fileutil.Exist(path) {
		url := fmt.Sprintf(types.FabricContractUrl, version)
		err := download.Download(path, url)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Download fabric contract successfully in %s\n", path)

	return nil
}
