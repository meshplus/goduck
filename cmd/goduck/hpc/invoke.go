package hpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/meshplus/goduck/internal/solidity"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/gosdk/hvm"
	"github.com/meshplus/gosdk/rpc"
	"github.com/meshplus/gosdk/utils/java"
)

const (
	prefix = "hyperchain.bitxhub.invoke."
)

func Invoke(configPath, abiPath, typ, address, function, args string) error {
	hpc, err := New(configPath)
	if err != nil {
		return err
	}

	abiData, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return fmt.Errorf("read abi: %w", err)
	}

	switch typ {
	case "jvm":
		rec, err := invokeJvm(hpc, address, function, args)
		if err != nil {
			return err
		}

		fmt.Printf("[hpc] invoke function \"%s\", receipt is %s\n", function, string(rec))
	case "hvm":
		rec, err := invokeJava(hpc, abiData, address, function, args)
		if err != nil {
			return err
		}

		fmt.Printf("[hpc] invoke function \"%s\", receipt is %s\n", function, string(rec))
	case "solc":
		rec, err := invokeSolidity(hpc, abiData, address, args, function)
		if err != nil {
			return err
		}

		ret, err := solidity.UnpackOutput(hpc.abi, function, string(rec), types.ChainTypeHpc)
		if err != nil {
			return err
		}

		if ret == nil {
			fmt.Printf("[hpc] invoke function \"%s\", no receipt\n", function)
			return nil
		}

		str := ""
		for _, r := range ret {
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

		fmt.Printf("[hpc] invoke function \"%s\", receipt is %s\n", function, str)
	default:

	}

	return nil
}

func invokeSolidity(hpc *Hyperchain, file []byte, address, args, function string) ([]byte, error) {
	ab, err := abi.JSON(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}

	hpc.abi = ab

	var argx []interface{}

	if len(args) != 0 {
		argSplits := strings.Split(args, ",")
		var argArr [][]byte
		for _, arg := range argSplits {
			if arg == "" {
				return nil, fmt.Errorf("contract parameter can't be empty")
			}
			argArr = append(argArr, []byte(arg))
		}

		argx, err = solidity.ABIUnmarshal(ab, argArr, function)
		if err != nil {
			return nil, err
		}
	}

	packed, err := ab.Pack(function, argx...)
	if err != nil {
		return nil, err
	}

	tranInvoke := rpc.NewTransaction(hpc.Key().GetAddress().String()).Invoke(address, packed)
	tranInvoke.Sign(hpc.Key())

	receipt, err := hpc.api.InvokeContract(tranInvoke)
	if err != nil {
		return nil, err
	}

	return []byte(receipt.Ret), nil
}

func invokeJava(hpc *Hyperchain, a []byte, address, function, args string) ([]byte, error) {
	javaAbi, err := hvm.GenAbi(string(a))
	if err != nil {
		return nil, err
	}

	easyBean := prefix + function
	beanAbi, err := javaAbi.GetBeanAbi(easyBean)
	if err != nil {
		return nil, err
	}

	var HArgs []interface{}
	if len(args) != 0 {
		cArgs := strings.Split(args, ",")
		HArgs = make([]interface{}, len(cArgs))
		for i, arg := range cArgs {
			HArgs[i] = arg
		}

	}

	pd, err := hvm.GenPayload(beanAbi, HArgs...)
	if err != nil {
		return nil, err
	}

	tranInvoke := rpc.NewTransaction(hpc.Key().GetAddress().String()).
		Invoke(address, pd).VMType(rpc.HVM)
	tranInvoke.Sign(hpc.Key())

	receipt, err := hpc.api.InvokeContract(tranInvoke)
	if err != nil {
		return nil, err
	}

	return []byte(java.DecodeJavaResult(receipt.Ret)), nil
}

func invokeJvm(hpc *Hyperchain, address, function, args string) ([]byte, error) {
	var cArgs []string
	if len(args) != 0 {
		cArgs = strings.Split(args, ",")
	}

	tranInvoke := rpc.NewTransaction(hpc.Key().GetAddress().String()).
		Invoke(address, java.EncodeJavaFunc(function, cArgs...)).VMType(rpc.JVM)
	tranInvoke.Sign(hpc.Key())

	receipt, err := hpc.api.InvokeContract(tranInvoke)
	if err != nil {
		return nil, err
	}

	return []byte(java.DecodeJavaResult(receipt.Ret)), nil
}
