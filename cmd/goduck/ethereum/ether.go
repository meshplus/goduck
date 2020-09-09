package ethereum

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

func StopEthereum(repoPath string) error {
	if !fileutil.Exist(filepath.Join(repoPath, types.EthereumScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return utils.ExecuteShell([]string{types.EthereumScript, "down"}, repoPath)
}

func StartEthereum(repoPath, mod string) error {
	if !fileutil.Exist(filepath.Join(repoPath, types.EthereumScript)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return utils.ExecuteShell([]string{types.EthereumScript, mod}, repoPath)
}
