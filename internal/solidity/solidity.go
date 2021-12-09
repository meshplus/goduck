package solidity

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/meshplus/bitxhub-kit/hexutil"
	"github.com/meshplus/goduck/internal/types"
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

func Encode(Abi abi.ABI, funcName string, args ...interface{}) ([]interface{}, error) {

	var method abi.Method
	var err error

	if funcName == "" {
		method = Abi.Constructor
	} else {
		method, err = getMethod(Abi, funcName)
		if err != nil {
			return nil, err
		}
	}

	if len(method.Inputs) > len(args) {
		return nil, fmt.Errorf("the num of inputs is %v, expectd %v", len(method.Inputs), len(args))
	}

	typedArgs := make([]interface{}, len(method.Inputs))

	for idx, input := range method.Inputs {
		typedArgs[idx] = convert(input.Type, args[idx])
	}

	return typedArgs, nil
}

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

func UnpackOutput(abi abi.ABI, method string, receipt string, chainTyp string) ([]interface{}, error) {
	m, err := getMethod(abi, method)
	if err != nil {
		return nil, fmt.Errorf("get method %w", err)
	}

	if len(m.Outputs) == 0 {
		return nil, nil
	}

	receiptData := []byte(receipt)
	if chainTyp == types.ChainTypeHpc {
		receiptData = hexutil.Decode(receipt)
	}
	res, err := abi.Unpack(method, receiptData)
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

// convert val into target type through certain method
// support: array / slice / bytesN / basic type
// not support: nested array or slice
func convert(t abi.Type, input interface{}) interface{} {
	// array or slice
	switch t.T {
	case abi.ArrayTy:
		// make sure that the length of input equals to the t.Size
		var (
			fmtVal = make([]interface{}, t.Size)
			idx    int
		)
		reflectInput := reflect.ValueOf(input)
		switch reflectInput.Kind() {
		case reflect.String:
			if t.Size >= 1 {
				fmtVal[idx] = input
				idx++
			}
		case reflect.Slice:
			valLen := reflectInput.Len()
			var formatLen int
			if valLen < t.Size {
				formatLen = valLen
			} else {
				formatLen = t.Size
			}
			for idx = 0; idx < formatLen; idx++ {
				fmtVal[idx] = reflectInput.Index(idx).Interface()
			}
		}

		// complete input with default "" (empty string)
		for i := idx; i < t.Size; i++ {
			fmtVal[idx] = ""
		}
		// build the array (not slice)
		data := reflect.New(t.GetType()).Elem()
		for idx, val := range fmtVal {
			elem := convert(*t.Elem, val)
			data.Index(idx).Set(reflect.ValueOf(elem))
		}
		return data.Interface()

	case abi.SliceTy:
		// todo: reflect
		var fmtVal []interface{}
		reflectInput := reflect.ValueOf(input)
		switch reflectInput.Kind() {
		case reflect.String:
			fmtVal = []interface{}{input}
		case reflect.Slice:
			inputLen := reflectInput.Len()
			fmtVal = make([]interface{}, inputLen)
			for i := 0; i < inputLen; i++ {
				fmtVal[i] = reflectInput.Index(i).Interface()
			}
		}

		data := reflect.MakeSlice(t.GetType(), len(fmtVal), len(fmtVal))
		for idx, val := range fmtVal {
			elem := convert(*t.Elem, val)
			data.Index(idx).Set(reflect.ValueOf(elem))
		}
		return data.Interface()

	case abi.FixedBytesTy:
		if str, ok := input.(string); ok {
			return newFixedBytes(t.Size, str)
		}
	default:
		if str, ok := input.(string); ok {
			return newElement(t, str)
		}

	}
	return nil
}

// convert from string to basic type element
func newElement(t abi.Type, val string) interface{} {
	if t.T == abi.SliceTy || t.T == abi.ArrayTy {
		return nil
	}
	var UNIT = 64
	var elem interface{}
	switch t.String() {
	case "uint8":
		num, _ := strconv.ParseUint(val, 10, UNIT)
		elem = uint8(num)
	case "uint16":
		num, _ := strconv.ParseUint(val, 10, UNIT)
		elem = uint16(num)
	case "uint32":
		num, _ := strconv.ParseUint(val, 10, UNIT)
		elem = uint32(num)
	case "uint64":
		num, _ := strconv.ParseUint(val, 10, UNIT)
		elem = uint64(num)
	case "uint128", "uint256", "int128", "int256":
		var num *big.Int
		if val == "" {
			num = big.NewInt(0)
		} else {
			num, _ = big.NewInt(0).SetString(val, 10)
		}
		elem = num
	case "int8":
		num, _ := strconv.ParseInt(val, 10, UNIT)
		elem = int8(num)
	case "int16":
		num, _ := strconv.ParseInt(val, 10, UNIT)
		elem = int16(num)
	case "int32":
		num, _ := strconv.ParseInt(val, 10, UNIT)
		elem = int32(num)
	case "int64":
		num, _ := strconv.ParseInt(val, 10, UNIT)
		elem = int64(num)
	case "bool":
		v, _ := strconv.ParseBool(val)
		elem = v
	case "address":
		elem = common.HexToAddress(val)
	case "string":
		elem = val
	case "bytes":
		elem = common.Hex2Bytes(val)
	default:
		// default use reflect but do not use val
		// because it's impossible to know how to convert from string to target type
		elem = reflect.New(t.GetType()).Elem().Interface()
	}

	return elem
}

var byteTy = reflect.TypeOf(byte(0))

// the return val is a byte array, not slice
func newFixedBytes(size int, val string) interface{} {
	// pre-define size 1,2,3...32 and 64, other size use reflect
	switch size {
	case 1:
		var data [1]byte
		copy(data[:], []byte(val))
		return data
	case 2:
		var data [2]byte
		copy(data[:], []byte(val))
		return data
	case 3:
		var data [3]byte
		copy(data[:], []byte(val))
		return data
	case 4:
		var data [4]byte
		copy(data[:], []byte(val))
		return data
	case 5:
		var data [5]byte
		copy(data[:], []byte(val))
		return data
	case 6:
		var data [6]byte
		copy(data[:], []byte(val))
		return data
	case 7:
		var data [7]byte
		copy(data[:], []byte(val))
		return data
	case 8:
		var data [8]byte
		copy(data[:], []byte(val))
		return data
	case 9:
		var data [9]byte
		copy(data[:], []byte(val))
		return data
	case 10:
		var data [10]byte
		copy(data[:], []byte(val))
		return data
	case 11:
		var data [11]byte
		copy(data[:], []byte(val))
		return data
	case 12:
		var data [12]byte
		copy(data[:], []byte(val))
		return data
	case 13:
		var data [13]byte
		copy(data[:], []byte(val))
		return data
	case 14:
		var data [14]byte
		copy(data[:], []byte(val))
		return data
	case 15:
		var data [15]byte
		copy(data[:], []byte(val))
		return data
	case 16:
		var data [16]byte
		copy(data[:], []byte(val))
		return data
	case 17:
		var data [17]byte
		copy(data[:], []byte(val))
		return data
	case 18:
		var data [18]byte
		copy(data[:], []byte(val))
		return data
	case 19:
		var data [19]byte
		copy(data[:], []byte(val))
		return data
	case 20:
		var data [20]byte
		copy(data[:], []byte(val))
		return data
	case 21:
		var data [21]byte
		copy(data[:], []byte(val))
		return data
	case 22:
		var data [22]byte
		copy(data[:], []byte(val))
		return data
	case 23:
		var data [23]byte
		copy(data[:], []byte(val))
		return data
	case 24:
		var data [24]byte
		copy(data[:], []byte(val))
		return data
	case 25:
		var data [25]byte
		copy(data[:], []byte(val))
		return data
	case 26:
		var data [26]byte
		copy(data[:], []byte(val))
		return data
	case 27:
		var data [27]byte
		copy(data[:], []byte(val))
		return data
	case 28:
		var data [28]byte
		copy(data[:], []byte(val))
		return data
	case 29:
		var data [29]byte
		copy(data[:], []byte(val))
		return data
	case 30:
		var data [30]byte
		copy(data[:], []byte(val))
		return data
	case 31:
		var data [31]byte
		copy(data[:], []byte(val))
		return data
	case 32:
		var data [32]byte
		copy(data[:], []byte(val))
		return data
	case 64:
		var data [64]byte
		copy(data[:], []byte(val))
		return data
	default:
		return newFixedBytesWithReflect(size, val)
	}
}

//! NOTICE: newFixedBytesWithReflect take more 15 times of time than newFixedBytes
//! So it is just use for those fixed bytes which are not commonly used.
func newFixedBytesWithReflect(size int, val string) interface{} {
	data := reflect.New(reflect.ArrayOf(size, byteTy)).Elem()
	bytes := reflect.ValueOf([]byte(val))
	reflect.Copy(data, bytes)
	return data.Interface()
}
