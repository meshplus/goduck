package hpc

import (
	"fmt"
	"strings"

	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/meshplus/gosdk/common"
	"github.com/meshplus/gosdk/hvm"
	"github.com/meshplus/gosdk/rpc"
	"github.com/meshplus/gosdk/utils/java"
	"github.com/pkg/errors"
	"github.com/ttacon/chalk"
)

func Deploy(configPath, codePath, typ string, local bool, args string) error {
	hpc, err := New(configPath)
	if err != nil {
		return err
	}

	addr, ret, err := hpc.deployContract(codePath, typ, local, args)
	if err != nil {
		return err
	}

	fmt.Printf("%sDeploy contract address: %s%s\n", chalk.Cyan, chalk.Reset, addr)
	fmt.Printf("%sContract abi:%s\n", chalk.Cyan, chalk.Reset)
	if ret != nil {
		for _, a := range ret.Abi {
			fmt.Println(a)
		}
	}

	return nil
}

func (h *Hyperchain) deployContract(path string, typ string, local bool, args string) (string, *rpc.CompileResult, error) {
	var code string
	var err error

	switch typ {
	case "jvm":
		// todo args
		payload, err := java.ReadJavaContract(path)
		if err != nil {
			return "", nil, err
		}

		tx := rpc.NewTransaction(h.key.GetAddress().String()).Deploy(payload).VMType(rpc.JVM)
		tx.Sign(h.key)

		txReceipt, err := h.api.DeployContract(tx)
		if err != nil {
			return "", nil, err
		}

		return txReceipt.ContractAddress, nil, nil
	case "java":
		// todo args
		bin, err := hvm.ReadJar(path)
		if err != nil {
			return "", nil, err
		}

		tx := rpc.NewTransaction(h.key.GetAddress().String()).Deploy(bin).VMType(rpc.HVM)
		tx.Sign(h.key)

		txReceipt, err := h.api.DeployContract(tx)
		if err != nil {
			return "", nil, err
		}

		return txReceipt.ContractAddress, nil, nil
	case "solc":
		code, err = common.ReadFileAsString(path)
		if err != nil {
			return "", nil, err
		}

		return h.deployContractWithCode([]byte(code), local, args)
	default:
		return "", nil, fmt.Errorf("no this type %s", typ)
	}
}

func (h *Hyperchain) deployContractWithCode(code []byte, local bool, args string) (string, *rpc.CompileResult, error) {
	compileResult, err := h.compileContract(string(code), local)
	if err != nil {
		return "", nil, err
	}

	bin := ""
	abi := ""
	// filter non-abstract contract
	for k, i := range compileResult.Bin {
		if len(common.HexToString(i)) > 1 {
			bin = i
			abi = compileResult.Abi[k]
			break
		}
	}
	if bin == "" {
		return "", nil, errors.New("cannot found non-abstract contract")
	}
	var tranDeploy *rpc.Transaction
	if "" == args {
		tranDeploy = rpc.NewTransaction(h.key.GetAddress().String()).Deploy(bin)
	} else {
		argSplits := strings.Split(args, "^")
		var argArr []interface{}
		for _, arg := range argSplits {
			if arg == "" {
				return "", nil, fmt.Errorf("contract parameter can't be empty")
			}
			if strings.Index(arg, "[") == 0 && strings.LastIndex(arg, "]") == len(arg)-1 {
				// deal with slice
				argSp := strings.Split(arg[1:len(arg)-1], ",")
				argArr = append(argArr, argSp)
				continue
			}
			argArr = append(argArr, arg)
		}
		tranDeploy = rpc.NewTransaction(h.key.GetAddress().String()).Deploy(bin).DeployStringArgs(abi, argArr...)
	}
	tranDeploy.Sign(h.key)
	txDeploy, err := h.api.DeployContract(tranDeploy)
	if err != nil {
		return "", nil, err
	}

	return eth_common.HexToAddress(txDeploy.ContractAddress).Hex(), compileResult, nil
}
