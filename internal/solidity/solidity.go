package solidity

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	//"github.com/meshplus/gosdk/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	Bool    = "bool"
	Address = "address"
	Int64   = "int64"
	Uint64  = "uint64"
	Int32   = "int32"
	Uint32  = "uint32"
	Int8    = "int8"
	Uint8   = "uint8"
	Bytes32 = "bytes32"
	String  = "string"
)

func ABIUnmarshal(ab abi.ABI, args [][]byte, funcName string) ([]interface{}, error) {
	argx := make([]interface{}, len(args))

	var m abi.Method
	var err error
	if funcName == "" {
		m = ab.Constructor
	} else {
		m, err = getMethod(ab, funcName)
		if err != nil {
			return nil, err
		}
	}

	if len(m.Inputs) != len(args) {
		return nil, errors.New("args'length is not equal")
	} else {
		for idx, arg := range args {
			argx[idx], err = transform(string(arg), m.Inputs[idx].Type.String())
			if err != nil {
				return nil, err
			}
		}
		return argx, nil
	}
}

func getMethod(ab abi.ABI, method string) (abi.Method, error) {
	for k, v := range ab.Methods {
		if k == method {
			return v, nil
		}
	}

	return abi.Method{}, fmt.Errorf("method %s is not existed", method)
}

func UnpackOutput(abi abi.ABI, method string, receipt string) ([]interface{}, error) {
	m, err := getMethod(abi, method)
	if err != nil {
		return nil, fmt.Errorf("get method %w", err)
	}

	if len(m.Outputs) == 0 {
		return nil, nil
	}

	res, err := abi.Unpack(method, []byte(receipt))
	if err != nil {
		return nil, fmt.Errorf("unpack result %w", err)
	}

	return res, nil
}

func transform(input string, t string) (interface{}, error) {
	switch t {
	// todo:
	case Address:
		var address common.Address
		copy(address[:], common.Hex2Bytes(input[2:]))
		return address, nil
	case String:
		return input, nil
	case Bytes32:
		var r [32]byte
		copy(r[:], []byte(input)[:32])
		return r, nil
	case "int256":
		return nil, errors.New("type overFitted")
	case Int64:
		r, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Uint64:
		r, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Int32:
		r, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Int8:
		r, err := strconv.ParseInt(input, 10, 8)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Uint8:
		r, err := strconv.ParseUint(input, 10, 8)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Uint32:
		r, err := strconv.ParseUint(input, 10, 32)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Bool:
		r, err := strconv.ParseBool(input)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.New("type overFitted")
	}
}
