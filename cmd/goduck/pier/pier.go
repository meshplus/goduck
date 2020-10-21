package pier

import (
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

func StartPier(repoRoot, chainType, cryptoPath, pierUpType string, version string, port string) error {
	args := []string{types.PierScript, "up", "-m", chainType, "-t", pierUpType,
		"-r", ".pier_" + chainType, "-v", version, "-c", cryptoPath, "-p", port}
	return utils.ExecuteShell(args, repoRoot)
}

func StopPier(repoRoot, chainType string) error {
	args := []string{types.PierScript, "down", "-m", chainType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}
	return nil
}

func CleanPier(repoRoot, chainType string) error {
	// clean pier first, then appchain
	args := []string{types.PierScript, "clean", "-m", chainType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}

	return nil
}
