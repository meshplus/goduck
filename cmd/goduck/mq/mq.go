package mq

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/gosdk/account"
	"github.com/meshplus/gosdk/common"
	"github.com/meshplus/gosdk/rpc"
)

type dmallMQ struct {
	key *account.ECDSAKey
	mq  *rpc.MqClient
}

func Register(address, queue string) error {
	dmq, err := New()
	if err != nil {
		return fmt.Errorf("create mq client: %w", err)
	}

	if err = dmq.informNormal(); err != nil {
		return fmt.Errorf("mq inform normal: %w", err)
	}

	meta := rpc.NewRegisterMeta(dmq.key.GetAddress().String(), queue, rpc.MQLog)
	meta.AddAddress(common.HexToAddress(address))
	meta.Sign(dmq.key)
	register, err := dmq.mq.Register(1, meta)
	if err != nil {
		return err
	}

	fmt.Printf("queue : %s, exchanger: %s\n", register.QueueName, register.ExchangerName)
	return nil
}

func Unregister(exchange, queue string) error {
	dmq, err := New()
	if err != nil {
		return fmt.Errorf("create mq client: %w", err)
	}

	_, err = dmq.mq.InformNormal(1, "")
	if err != nil {
		return err
	}

	meta := rpc.NewUnRegisterMeta(dmq.key.GetAddress().String(), queue, exchange)
	meta.Sign(dmq.key)
	if _, err := dmq.mq.UnRegister(1, meta); err != nil {
		return err
	}

	fmt.Printf("unregister queue : %s from exchanger: %s\n", queue, exchange)
	return nil
}

func New() (*dmallMQ, error) {
	accountJson := `{"address":"0xefb945a2c6f4d2f8b7f3fcc28450c0ca34b98ae0","algo":"0x02","encrypted":"9ae36b184d9a00f07e93c829346282cfb25ffb05d67f14b710f1f3675321d254d5af0c356f906ad4","version":"1.0","privateKeyEncrypted":true}`

	key, err := account.NewAccountFromAccountJSON(accountJson, "dmall")
	if err != nil {
		return nil, err
	}

	repoRoot, err := repo.PathRoot()
	if err != nil {
		return nil, err
	}

	rpcCli := rpc.NewRPCWithPath(filepath.Join(repoRoot, "hyperchain"))
	mq := rpcCli.GetMqClient()

	return &dmallMQ{
		key: key,
		mq:  mq,
	}, nil
}

func (dmq *dmallMQ) informNormal() error {
	_, err := dmq.mq.InformNormal(1, "")
	if err != nil {
		return err
	}

	return nil
}

func (dmq *dmallMQ) Query() error {
	queueNames, err := dmq.mq.GetAllQueueNames(1)
	if err != nil {
		return err
	}

	for i, v := range queueNames {
		fmt.Printf("%d: %s\n", i, v)
	}
	return nil
}
