package ethereum

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/urfave/cli/v2"
)

var ethContract = []string{
	"broker",
	"transfer",
	"data_swapper",
}

var ContractCMD = &cli.Command{
	Name:  "contract",
	Usage: "Operation about solidity contract",
	Subcommands: []*cli.Command{
		{
			Name:  "deploy",
			Usage: "Deploy solidity contract to ethereum chain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "address",
					Usage:    "specify the address of ethereum chain",
					Value:    "http://localhost:8545",
					Required: false,
				},
				&cli.StringFlag{
					Name:  "key-path",
					Usage: "specify the ethereum account private key path",
				},
				&cli.StringFlag{
					Name:  "psd-path",
					Usage: "specify ethereum account password path",
				},
				&cli.StringFlag{
					Name:     "code-path",
					Usage:    "specify the path of solidity contract (If there are multiple contracts, separate their paths with \",\")",
					Required: true,
				},
			},
			ArgsUsage: "\n\t command: goduck ether contract deploy [args(optional)]",
			Action: func(ctx *cli.Context) error {
				config := Config{
					EtherAddr:    ctx.String("address"),
					KeyPath:      ctx.String("key-path"),
					PasswordPath: ctx.String("psd-path"),
				}

				codePath := ctx.String("code-path")
				args := ctx.Args()

				return Deploy(config, codePath, args.First())
			},
		},
		{
			Name:  "invoke",
			Usage: "Invoke solidity contract on ethereum chain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "address",
					Usage:    "specify the address of ethereum chain",
					Value:    "http://localhost:8545",
					Required: false,
				},
				&cli.StringFlag{
					Name:  "key-path",
					Usage: "specify the ethereum account private key path",
				},
				&cli.StringFlag{
					Name:  "psd-path",
					Usage: "specify ethereum account password path",
				},
				&cli.StringFlag{
					Name:     "abi-path",
					Usage:    "specify the path of solidity contract abi file",
					Required: true,
				},
			},
			ArgsUsage: "\n\t command: goduck ether contract invoke [contract_address] [function] [args(optional)]",
			Action: func(ctx *cli.Context) error {
				config := Config{
					EtherAddr:    ctx.String("address"),
					KeyPath:      ctx.String("key-path"),
					PasswordPath: ctx.String("psd-path"),
				}

				abiPath := ctx.String("abi-path")

				if ctx.NArg() < 2 {
					return fmt.Errorf("args must be (dst_addr function args[optional])")
				}

				args := ctx.Args().Slice()
				if ctx.NArg() == 2 {
					args = append(args, "")
				}
				dstAddr := args[0]
				function := args[1]
				argAbi := args[2]

				return Invoke(config, abiPath, dstAddr, function, argAbi)
			},
		},
		{
			Name:  "trust",
			Usage: "get trust meta",
			Flags: []cli.Flag{
				&cli.Int64Flag{
					Name:     "height",
					Usage:    "block height",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "address",
					Usage:    "the address of ethereum chain",
					Value:    "http://localhost:8545",
					Required: false,
				},
			},
			Action: getTrustMeta,
		},
		{
			Name:  "download",
			Usage: "download the default cross-chain contract on ethereum chain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "version",
					Usage:    "specify bitxhub version",
					Value:    "v1.6.5",
					Required: false,
				},
			},
			Action: downloadContract,
		},
	},
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

	for _, contract := range ethContract {
		path := filepath.Join(repoPath, types.ChainTypeEther, "contract", version, fmt.Sprintf("%s.sol", contract))
		if !fileutil.Exist(path) {
			url := fmt.Sprintf(types.EthereumContractUrl, version, contract)
			err := download.Download(path, url)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Download ethereum contract successfully in %s\n", path)
	}

	return nil
}
func getTrustMeta(ctx *cli.Context) error {
	height := ctx.Int64("height")
	etherAddr := ctx.String("address")

	etherCli, err := ethclient.Dial(etherAddr)
	if err != nil {
		return err
	}

	etherSession := &EtherSession{
		etherCli: etherCli,
		ctx:      context.Background(),
	}
	trustMeta, err := etherSession.getTrustMeta(height)
	if err != nil {
		return err
	}
	fmt.Println(trustMeta)
	return nil
}
