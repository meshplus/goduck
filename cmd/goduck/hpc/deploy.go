package hpc

import (
	"fmt"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/gosdk/common"
	"github.com/meshplus/gosdk/hvm"
	"github.com/meshplus/gosdk/rpc"
	"github.com/meshplus/gosdk/utils/java"
	"github.com/pkg/errors"
	"github.com/ttacon/chalk"
)

func Deploy(configPath, codePath, typ string, local bool) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	hpc, err := New(configPath, repoRoot)
	if err != nil {
		return err
	}

	addr, ret, err := hpc.deployContract(codePath, typ, local)
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

func (h *Hyperchain) deployContract(path string, typ string, local bool) (string, *rpc.CompileResult, error) {
	var code string
	var err error

	switch typ {
	case "jvm":
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

		return h.deployContractWithCode([]byte(code), local)
	default:
		return "", nil, fmt.Errorf("no this type %s", typ)
	}
}

func (h *Hyperchain) deployContractWithCode(code []byte, local bool) (string, *rpc.CompileResult, error) {
	compileResult, err := h.compileContract(string(code), local)
	if err != nil {
		return "", nil, err
	}

	bin := ""
	// filter non-abstract contract
	for _, i := range compileResult.Bin {
		if len(common.HexToString(i)) > 1 {
			bin = i
			break
		}
	}
	if bin == "" {
		return "", nil, errors.New("cannot found non-abstract contract")
	}
	tranDeploy := rpc.NewTransaction(h.key.GetAddress().String()).Deploy(bin)
	tranDeploy.Sign(h.key)
	txDeploy, err := h.api.DeployContract(tranDeploy)
	if err != nil {
		return "", nil, err
	}

	return txDeploy.ContractAddress, compileResult, nil
}
