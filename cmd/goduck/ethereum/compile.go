package ethereum

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/compiler"
)

type CompileResult struct {
	Abi   []string
	Bins  []string
	Types []string
}

func compileSolidityCode(codePath string) (*CompileResult, error) {
	contracts, err := compiler.CompileSolidity("solc", codePath)
	if err != nil {
		return nil, fmt.Errorf("compile contract: %w", err)
	}

	var (
		abis  []string
		bins  []string
		types []string
	)
	for name, contract := range contracts {
		abi, _ := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
		abis = append(abis, string(abi))
		bins = append(bins, contract.Code)
		types = append(types, name)
	}

	result := &CompileResult{
		Abi:   abis,
		Bins:  bins,
		Types: types,
	}
	return result, nil
}
