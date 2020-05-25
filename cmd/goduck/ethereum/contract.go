package ethereum

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

var contractCMD = &cli.Command{
	Name:  "contract",
	Usage: "operation about solidity contract",
	Subcommands: []*cli.Command{
		{
			Name:  "deploy",
			Usage: "deploy solidity contract to ethereum chain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "ether_addr",
					Usage:    "the address of ethereum chain",
					Value:    "localhost:8545",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "key_path",
					Usage:    "the ethereum account private key path",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "code_path",
					Usage:    "the path of solidity contract",
					Required: true,
				},
			},
			Action: deploy,
		},
	},
}

func deploy(ctx *cli.Context) error {
	etherAddr := ctx.String("ether_addr")
	keyPath := ctx.String("key_path")
	codePath := ctx.String("code_path")

	etherCli, privateKey, err := help(etherAddr, keyPath)
	if err != nil {
		return err
	}

	// compile solidity first
	compileResult, err := compileSolidityCode(codePath)
	if err != nil {
		return err
	}

	if len(compileResult.Abi) == 0 || len(compileResult.Bins) == 0 || len(compileResult.Types) == 0 {
		return fmt.Errorf("empty contract")
	}
	// deploy a contract
	auth := bind.NewKeyedTransactor(privateKey)
	parsed, err := abi.JSON(strings.NewReader(compileResult.Abi[0]))
	if err != nil {
		return err
	}

	for i, bin := range compileResult.Bins {
		addr, _, _, err := bind.DeployContract(auth, parsed, common.FromHex(bin), etherCli)
		if err != nil {
			return err
		}
		fmt.Printf("\n======= %s =======\n", compileResult.Types[i])
		fmt.Printf("Deployed contract address is %s\n", addr.Hex())
		fmt.Printf("Contract JSON ABI\n%s\n", compileResult.Abi[i])
	}

	return nil
}

func help(etherAddr, keyPath string) (*ethclient.Client, *ecdsa.PrivateKey, error) {
	etherCli, err := ethclient.Dial(etherAddr)
	if err != nil {
		return nil, nil, err
	}

	keyByte, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	type EtherAccount struct {
		PrivateKeys map[string]string `json:"private_keys"`
	}
	accounts := &EtherAccount{}
	if err := json.Unmarshal(keyByte, accounts); err != nil {
		return nil, nil, err
	}

	var privateKey *ecdsa.PrivateKey
	for _, account := range accounts.PrivateKeys {
		privateKey, err = crypto.HexToECDSA(account)
		if err != nil {
			return nil, nil, err
		}
		break
	}

	return etherCli, privateKey, nil
}
