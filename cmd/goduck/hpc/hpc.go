package hpc

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/meshplus/gosdk/account"
	"github.com/meshplus/gosdk/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logger = logrus.New().WithFields(logrus.Fields{
	"module": "hpc",
})

type Hyperchain struct {
	key          *account.ECDSAKey
	api          *rpc.RPC
	abi          abi.ABI
	contractAddr string
}

func New(configPath, repoRoot string) (*Hyperchain, error) {
	defaultConfigPath := filepath.Join(repoRoot, "hyperchain")

	var hpcPath string
	if len(configPath) == 0 {
		hpcPath = defaultConfigPath
	} else {
		hpcPath = configPath
	}

	keyPath := filepath.Join(hpcPath, "hpc.account")
	accountData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	psdPath := filepath.Join(hpcPath, "password")
	psd, err := ioutil.ReadFile(psdPath)
	if err != nil {
		return nil, err
	}
	password := strings.TrimSpace(string(psd))

	key, err := account.NewAccountFromAccountJSON(string(accountData), password)
	if err != nil {
		return nil, err
	}

	api := rpc.NewRPCWithPath(hpcPath)
	return &Hyperchain{
		key:          key,
		api:          api,
		contractAddr: viper.GetString("blockchain.contract.address"),
	}, nil
}

func (h *Hyperchain) Key() *account.ECDSAKey {
	return h.key
}

func (h *Hyperchain) compileContract(code string, local bool) (*rpc.CompileResult, error) {
	if !local {
		return h.api.CompileContract(code)
	}

	c, err := NewCompiler("")
	if err != nil {
		return nil, err
	}

	contracts, err := c.Compile(code)
	if err != nil {
		return nil, err
	}
	var (
		abis  []string
		bins  []string
		types []string
	)
	// Gather all non-excluded contract for binding
	for name, contract := range contracts {
		abi, _ := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
		abis = append(abis, string(abi))
		bins = append(bins, contract.Code)
		if c.isSolcjs {
			types = append(types, strings.Split(name, ":")[1])
		} else {
			types = append(types, name)
		}
	}
	return &rpc.CompileResult{
		Abi:   abis,
		Bin:   bins,
		Types: types,
	}, nil
}
