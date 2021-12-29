package ethereum

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types1 "github.com/ethereum/go-ethereum/core/types"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/solidity"
)

func Deploy(config Config, codePath, argContract string) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}
	ether, err := New(config, repoRoot)
	if err != nil {
		return err
	}

	// compile solidity first
	compileResult, err := compileSolidityCode(codePath)
	if err != nil {
		return err
	}

	if len(compileResult.Abis) == 0 || len(compileResult.Bins) == 0 || len(compileResult.Types) == 0 {
		return fmt.Errorf("empty contract")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(ether.privateKey, ether.cid)
	if err != nil {
		return err
	}

	for i, bin := range compileResult.Bins {
		if bin == "0x" {
			continue
		}
		parsed, err := abi.JSON(strings.NewReader(compileResult.Abis[i]))
		if err != nil {
			return err
		}
		code := strings.TrimPrefix(strings.TrimSpace(bin), "0x")

		// prepare for constructor parameters
		var argx []interface{}
		if len(argContract) != 0 {
			rex := regexp.MustCompile(`(\[([^]]+)\])|([\w]+)`)
			out := rex.FindAllStringSubmatch(argContract, -1)
			params := make([]interface{}, len(out))
			for k, v := range out {
				var param []interface{}
				v[0] = strings.TrimSpace(v[0])
				v[0] = strings.ReplaceAll(v[0], "[", "")
				v[0] = strings.ReplaceAll(v[0], "]", "")
				str := strings.Split(v[0], ",")
				if len(str) == 1 {
					params[k] = v[0]
				} else {
					for _, v := range str {
						param = append(param, v)
					}
					params[k] = param
				}
			}
			argx, err = solidity.Encode(parsed, "", argx)
			if err != nil {
				return err
			}
		}

		addr, tx, _, err := bind.DeployContract(auth, parsed, common.FromHex(code), ether.etherCli, argx...)
		if err != nil {
			return err
		}
		var r *types1.Receipt
		if err := retry.Retry(func(attempt uint) error {
			r, err = ether.etherCli.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				return err
			}

			return nil
		}, strategy.Wait(1*time.Second)); err != nil {
			return err
		}

		if r.Status == types1.ReceiptStatusFailed {
			return fmt.Errorf("deploy contract failed, tx hash is: %s", r.TxHash.Hex())
		}

		fmt.Printf("\n======= %s =======\n", compileResult.Types[i])
		fmt.Printf("Deployed contract address is %s\n", addr.Hex())
		fmt.Printf("Contract JSON ABI\n%s\n", compileResult.Abis[i])

		//write abi file
		dir := filepath.Dir(compileResult.Types[i])
		base := filepath.Base(compileResult.Types[i])
		ext := filepath.Ext(compileResult.Types[i])
		f := strings.TrimSuffix(base, ext)
		filename := fmt.Sprintf("%s.abi", f)
		p := filepath.Join(dir, filename)
		err = ioutil.WriteFile(p, []byte(compileResult.Abis[i]), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
