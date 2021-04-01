package ethereum

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeploy(t *testing.T) {
	compileResult, err := compileSolidityCode("solidity/broker.sol")
	require.Nil(t, err)

	data, err := json.Marshal(compileResult)
	require.Nil(t, err)


	fmt.Println(string(data))
}
