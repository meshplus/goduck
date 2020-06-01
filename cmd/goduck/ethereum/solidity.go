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
	case Int32:
		var num int32
		return &num, Int32
	case Uint32:
		var num uint32
		return &num, Uint32
	case Int8:
		var num int8
		return &num, Int8
	case Uint8:
		var num uint8
		return &num, Uint8
	case Int64:
		var num int64
		return &num, Int64
	case Uint64:
		var num uint64
		return &num, Uint64
	case Bool:
		var num bool
		return &num, Bool
	case String:
		var num string
		return &num, String
	case Bytes32:
		var num [32]byte
		return &num, Bytes32
	case Address:
		var num common.Address
		return &num, Address
	}

	return nil, ""
}

func getPacker(args abi.Arguments) *packer {
	packers := make([]interface{}, len(args))
	types := ""
	for i, arg := range args {
		switch arg.Type.String() {
		case Int32:
			var num int32
			packers[i] = &num
			types += "int32,"
		case Uint32:
			var num uint32
			packers[i] = &num
			types += "uint32,"
		case Int8:
			var num int8
			packers[i] = &num
			types += "int8,"
		case Uint8:
			var num uint8
			packers[i] = &num
			types += "uint8,"
		case Int64:
			var num int64
			packers[i] = &num
			types += "int64,"
		case Uint64:
			var num uint64
			packers[i] = &num
			types += "uint64,"
		case "uint64[]":
			var nums []uint64
			packers[i] = &nums
			types += "uint64[],"
		case Bool:
			var f bool
			packers[i] = &f
			types += "bool,"
		case String:
			var s string
			packers[i] = &s
			types += "string,"
		case Bytes32:
			var b [32]byte
			packers[i] = &b
			types += "bytes32,"
		case Address:
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
	case Int32:
		return *ret.(*int32)
	case Uint32:
		return *ret.(*uint32)
	case Int8:
		return *ret.(*int8)
	case Uint8:
		return *ret.(*uint8)
	case Int64:
		return *ret.(*int64)
	case Uint64:
		return *ret.(*uint64)
	case "uint64[]":
		return reflect.ValueOf(ret).Interface()
	case Bool:
		return *ret.(*bool)
	case String:
		return *ret.(*string)
	case Bytes32:
		return *ret.(*[32]byte)
	case Address:
		return *ret.(*common.Address)
	case "address[]":
		return reflect.ValueOf(ret).Interface()
	}

	return nil
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
