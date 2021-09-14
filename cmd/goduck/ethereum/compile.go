package ethereum

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/compiler"
)

type CompileResult struct {
	Abis  []string
	Bins  []string
	Types []string
}

func compileSolidityCode(codePath string) (*CompileResult, error) {
	codePaths := strings.Split(codePath, ",")
	contracts, err := compiler.CompileSolidity("", codePaths...)
	if err != nil {
		return nil, fmt.Errorf("compile contract: %w", err)
	}

	var (
		abis  []string
		bins  []string
		types []string
	)
	for name, contract := range contracts {
		Abi, err := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABIs from compiler output: %w", err)
		}
		abis = append(abis, string(Abi))
		bins = append(bins, contract.Code)
		types = append(types, name)
	}

	result := &CompileResult{
		Abis:  abis,
		Bins:  bins,
		Types: types,
	}
	return result, nil
}
