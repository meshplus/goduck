package ethereum

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	compileResult, err := compileSolidityCode("solidity/broker.sol")
	require.Nil(t, err)

	data, err := json.Marshal(compileResult)
	require.Nil(t, err)

	fmt.Println(string(data))
}
