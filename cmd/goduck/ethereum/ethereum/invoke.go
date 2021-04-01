package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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
	result, err := UnpackOutput(es.ab, function, string(output))
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
