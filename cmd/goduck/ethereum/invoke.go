package ethereum

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	types1 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/solidity"
)

type EtherSession struct {
	etherCli   *ethclient.Client
	privateKey *ecdsa.PrivateKey
	ctx        context.Context
	ab         abi.ABI
}

func (es *EtherSession) ethCall(invokerAddr, to *common.Address, function string, packed []byte) ([]interface{}, error) {
	//for read only call
	var (
		msg    = ethereum.CallMsg{From: *invokerAddr, To: to, Data: packed}
		output []byte
	)
	output, err := es.etherCli.CallContract(es.ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		if code, err := es.etherCli.CodeAt(es.ctx, *to, nil); err != nil {
			return nil, err
		} else if len(code) == 0 {
			return nil, fmt.Errorf("no code at your contract addresss")
		}
		return nil, fmt.Errorf("output is empty")
	}

	// unpack result for display
	result, err := solidity.UnpackOutput(es.ab, function, string(output))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (es *EtherSession) ethTx(invokerAddr, to *common.Address, packed []byte) (*types.Transaction, error) {
	// for write only transaction
	signedTx, err := es.buildTx(invokerAddr, to, packed)
	if err != nil {
		return nil, err
	}
	if err := es.etherCli.SendTransaction(es.ctx, signedTx); err != nil {
		return nil, err
	}

	return signedTx, nil
}

func (es *EtherSession) buildTx(from, dstAddr *common.Address, input []byte) (*types.Transaction, error) {
	nonce, err := es.etherCli.PendingNonceAt(es.ctx, *from)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	gasPrice, err := es.etherCli.SuggestGasPrice(es.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethereum.CallMsg{From: *from, To: dstAddr, GasPrice: gasPrice, Value: new(big.Int), Data: input}
	gasLimit, err := es.etherCli.EstimateGas(es.ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	// Create the transaction, sign it and schedule it for execution
	rawTx := types.NewTransaction(nonce, *dstAddr, new(big.Int), gasLimit, gasPrice, input)
	signer := types.HomesteadSigner{}
	signature, err := crypto.Sign(signer.Hash(rawTx).Bytes(), es.privateKey)
	if err != nil {
		return nil, err
	}

	return rawTx.WithSignature(signer, signature)
}

func Invoke(config Config, abiPath, dstAddr, function, argAbi string) error {
	repoRoot, err := repo.PathRoot()
	file, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return err
	}

	ether, err := New(config, repoRoot)
	if err != nil {
		return err
	}

	ab, err := abi.JSON(bytes.NewReader(file))
	if err != nil {
		return err
	}
	etherSession := &EtherSession{
		privateKey: ether.privateKey,
		etherCli:   ether.etherCli,
		ctx:        context.Background(),
		ab:         ab,
	}

	// prepare for invoke parameters
	var argx []interface{}
	if len(argAbi) != 0 {
		argSplits := strings.Split(argAbi, ",")
		var argArr [][]byte
		for _, arg := range argSplits {
			argArr = append(argArr, []byte(arg))
		}
		argx, err = solidity.ABIUnmarshal(ab, argArr, function)
		if err != nil {
			return err
		}
	}
	packed, err := ab.Pack(function, argx...)
	if err != nil {
		return err
	}
	invokerAddr := crypto.PubkeyToAddress(ether.privateKey.PublicKey)
	to := common.HexToAddress(dstAddr)

	fmt.Printf("\n======= invoke function %s =======\n", function)
	if ab.Methods[function].IsConstant() {
		// for read only eth calls
		result, err := etherSession.ethCall(&invokerAddr, &to, function, packed)
		if err != nil {
			return err
		}

		if result == nil {
			fmt.Println("no result")
			return nil
		}

		str := ""
		for _, r := range result {
			if r != nil {
				if reflect.TypeOf(r).String() == "[32]uint8" {
					v, ok := r.([32]byte)
					if ok {
						r = string(v[:])
					}
				}
			}
			str = fmt.Sprintf("%s,%v", str, r)
		}

		str = strings.Trim(str, ",")
		fmt.Printf("call result: %s\n", str)
		return nil
	}

	// for write only eth transaction
	signedTx, err := etherSession.ethTx(&invokerAddr, &to, packed)
	if err != nil {
		return err
	}

	var r *types1.Receipt
	if err := retry.Retry(func(attempt uint) error {
		r, err = ether.etherCli.TransactionReceipt(context.Background(), signedTx.Hash())
		if err != nil {
			return err
		}

		return nil
	}, strategy.Wait(1*time.Second)); err != nil {
		return err
	}

	if r.Status == types1.ReceiptStatusFailed {
		return fmt.Errorf("invoke contract failed, tx hash is: %s", r.TxHash.Hex())
	}
	fmt.Printf("invoke contract success, tx hash is: %s\n", signedTx.Hash().Hex())
	return nil
}
