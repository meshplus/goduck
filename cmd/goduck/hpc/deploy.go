package hpc

import (
	"fmt"
	"regexp"
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
	if len(args) != 0 {
		tranDeploy = rpc.NewTransaction(h.key.GetAddress().String()).Deploy(bin)
	} else {
		rex := regexp.MustCompile(`(\[([^]]+)\])|([\w]+)`)
		out := rex.FindAllStringSubmatch(args, -1)
		var params []interface{}
		for _, v := range out {
			var param []interface{}
			isArray := false
			if strings.Contains(v[0], "[") {
				isArray = true
			}
			v[0] = strings.TrimSpace(v[0])
			v[0] = strings.ReplaceAll(v[0], "[", "")
			v[0] = strings.ReplaceAll(v[0], "]", "")
			str := strings.Split(v[0], ",")
			if !isArray {
				params = append(params, v[0])
			} else {
				for _, v := range str {
					param = append(param, v)
				}
				params = append(params, param)
			}
		}
		tranDeploy = rpc.NewTransaction(h.key.GetAddress().String()).Deploy(bin).DeployStringArgs(abi, params...)
	}
	tranDeploy.Sign(h.key)
	txDeploy, err := h.api.DeployContract(tranDeploy)
	if err != nil {
		return "", nil, err
	}

	return eth_common.HexToAddress(txDeploy.ContractAddress).Hex(), compileResult, nil
}
