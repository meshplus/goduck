package hpc

import (
	"fmt"
	"io/ioutil"

	"github.com/meshplus/gosdk/hvm"
	"github.com/meshplus/gosdk/rpc"
	"github.com/meshplus/gosdk/utils/java"
)

func Update(configPath, codePath, typ string, local bool, conAddr string) error {
	hpc, err := New(configPath)
	if err != nil {
		return err
	}

	switch typ {
	case "hvm":
		bin, err := hvm.ReadJar(codePath)
		if err != nil {
			return err
		}

		return hpc.updateContract(bin, conAddr, rpc.HVM)
	case "jvm":
		bin, err := java.ReadJavaContract(codePath)
		if err != nil {
			return err
		}

		return hpc.updateContract(bin, conAddr, rpc.JVM)
	case "solc":
		code, err := ioutil.ReadFile(codePath)
		if err != nil {
			return err
		}

		ret, err := hpc.compileContract(string(code), local)
		if err != nil {
			return err
		}

		bin, err := mainContractBin(ret)
		if err != nil {
			return err
		}

		return hpc.updateContract(bin, conAddr, rpc.EVM)
	default:
		return fmt.Errorf("not support contract type: %s", typ)
	}
}

func (h *Hyperchain) updateContract(bin string, addr string, typ rpc.VMType) error {
	tx := rpc.NewTransaction(h.key.GetAddress().String()).Maintain(1, addr, bin).VMType(typ)
	tx.Sign(h.key)
	rec, err := h.api.MaintainContract(tx)
	if err != nil {
		return err
	}

	if !rec.Valid {
		return fmt.Errorf(rec.Ret)
	}

	return nil
}

func mainContractBin(res *rpc.CompileResult) (string, error) {
	idx := 0
	for i, r := range res.Bin {
		if len(r) != 2 {
			idx = i
			break
		}
	}

	return res.Bin[idx], nil
}
