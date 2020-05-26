package ethereum

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func ABIUnmarshal(abi abi.ABI, args [][]byte, funcName string) ([]interface{}, error) {
	argx := make([]interface{}, len(args))

	m, err := getMethod(abi, funcName)
	if err != nil {
		return nil, err
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

	if len(m.Outputs) == 1 {
		p, t := getSinglePacker(m.Outputs)
		if err := abi.Unpack(p, method, []byte(receipt)); err != nil {
			return nil, fmt.Errorf("unpack result %w", err)
		}

		return []interface{}{unpackResult(p, t)}, nil
	} else {
		p := getPacker(m.Outputs)
		if err := abi.Unpack(&p.packers, method, []byte(receipt)); err != nil {
			return nil, err
		}

		ts := strings.Split(p.types, ",")

		ret := make([]interface{}, len(p.packers))
		for i, r := range p.packers {
			ret[i] = unpackResult(r, ts[i])
		}

		return ret, nil

	}
}

type packer struct {
	types   string
	packers []interface{}
}

func getSinglePacker(args abi.Arguments) (interface{}, string) {
	switch args[0].Type.String() {
	case "int32":
		var num int32
		return &num, "int32"
	case "uint32":
		var num uint32
		return &num, "uint32"
	case "int8":
		var num int8
		return &num, "int8"
	case "uint8":
		var num uint8
		return &num, "uint8"
	case "int64":
		var num int64
		return &num, "int64"
	case "uint64":
		var num uint64
		return &num, "uint64"
	case "bool":
		var num bool
		return &num, "bool"
	case "string":
		var num string
		return &num, "string"
	case "bytes32":
		var num [32]byte
		return &num, "bytes32"
	case "address":
		var num common.Address
		return &num, "address"
	}

	return nil, ""
}

func getPacker(args abi.Arguments) *packer {
	packers := make([]interface{}, len(args))
	types := ""
	for i, arg := range args {
		switch arg.Type.String() {
		case "int32":
			var num int32
			packers[i] = &num
			types += "int32,"
		case "uint32":
			var num uint32
			packers[i] = &num
			types += "uint32,"
		case "int8":
			var num int8
			packers[i] = &num
			types += "int8,"
		case "uint8":
			var num uint8
			packers[i] = &num
			types += "uint8,"
		case "int64":
			var num int64
			packers[i] = &num
			types += "int64,"
		case "uint64":
			var num uint64
			packers[i] = &num
			types += "uint64,"
		case "uint64[]":
			var nums []uint64
			packers[i] = &nums
			types += "uint64[],"
		case "bool":
			var f bool
			packers[i] = &f
			types += "bool,"
		case "string":
			var s string
			packers[i] = &s
			types += "string,"
		case "bytes32":
			var b [32]byte
			packers[i] = &b
			types += "bytes32,"
		case "address":
			var addr common.Address
			packers[i] = &addr
			types += "address,"
		case "address[]":
			var addrs []common.Address
			packers[i] = &addrs
			types += "address[],"
		}
	}

	types = strings.TrimRight(types, ",")

	return &packer{
		types:   types,
		packers: packers,
	}
}

func unpackResult(ret interface{}, types string) interface{} {
	switch types {
	case "int32":
		return *ret.(*int32)
	case "uint32":
		return *ret.(*uint32)
	case "int8":
		return *ret.(*int8)
	case "uint8":
		return *ret.(*uint8)
	case "int64":
		return *ret.(*int64)
	case "uint64":
		return *ret.(*uint64)
	case "uint64[]":
		return reflect.ValueOf(ret).Interface()
	case "bool":
		return *ret.(*bool)
	case "string":
		return *ret.(*string)
	case "bytes32":
		return *ret.(*[32]byte)
	case "address":
		return *ret.(*common.Address)
	case "address[]":
		return reflect.ValueOf(ret).Interface()
	}

	return nil
}

func transform(input string, t string) (interface{}, error) {
	switch t {
	// todo:
	case "address":
		var address common.Address
		copy(address[:], common.Hex2Bytes(input[2:]))
		return address, nil
	case "string":
		return input, nil
	case "bytes32":
		var r [32]byte
		copy(r[:], []byte(input)[:32])
		return r, nil
	case "int256":
		return nil, errors.New("type overFitted")
	case "int64":
		r, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, err
		}
		return r, nil
	case "uint64":
		r, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
		return r, nil
	case "int32":
		r, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return nil, err
		}
		return r, nil
	case "uint8":
		r, err := strconv.ParseUint(input, 10, 8)
		if err != nil {
			return nil, err
		}
		return r, nil
	case "uint32":
		r, err := strconv.ParseUint(input, 10, 32)
		if err != nil {
			return nil, err
		}
		return r, nil
	case "bool":
		r, err := strconv.ParseBool(input)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.New("type overFitted")
	}
}
