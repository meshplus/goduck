package fabric

import (
	"path/filepath"

	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

// start a fabric network
func Start(repoRoot, cryptoConfigPath string) error {
	var (
		cryptoPath string
		err        error
	)
	if cryptoConfigPath != "" {
		cryptoPath, err = filepath.Abs(cryptoConfigPath)
		if err != nil {
			return err
		}
	}

	args := []string{filepath.Join(repoRoot, types.FabricScript), "up", cryptoPath}

	return utils.ExecuteShell(args, repoRoot)
}

// stop a fabric network
func Stop(repoRoot string) error {
	args := []string{filepath.Join(repoRoot, types.FabricScript), "down"}

	return utils.ExecuteShell(args, repoRoot)
}
