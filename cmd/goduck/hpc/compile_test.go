package hpc

import (
	"io/ioutil"
	"testing"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestCompileContract(t *testing.T) {
	code, err := ioutil.ReadFile("./testdata/get.sol")
	require.Nil(t, err)

	repoRoot, err := repo.PathRoot()
	require.Nil(t, err)

	hpc, err := New("", repoRoot)
	require.Nil(t, err)
	ret, err := hpc.compileContract(string(code), false)
	require.Nil(t, err)
	require.Equal(t, 2, len(ret.Bin))
}
