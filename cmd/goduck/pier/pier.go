package pier

import (
	"fmt"

	"github.com/meshplus/goduck/cmd/goduck/ethereum"
	"github.com/meshplus/goduck/cmd/goduck/fabric"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

func StartPier(repoRoot, chainType, cryptoPath, pierUpType string, version string) error {
	args := []string{types.PierScript, "up", "-m", chainType, "-t", pierUpType,
		"-r", ".pier_" + chainType, "-v", version, "-p", cryptoPath}
	return utils.ExecuteShell(args, repoRoot)
}

func StopPier(repoRoot, chainType string) error {
	args := []string{types.PierScript, "down", "-m", chainType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}
	return nil
}

func CleanPier(repoRoot, chainType string, pierOnly bool) error {
	// clean pier first, then appchain
	args := []string{types.PierScript, "clean", "-m", chainType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}

	if pierOnly {
		return nil
	}

	switch chainType {
	case types.Fabric:
		return fabric.Stop(repoRoot)
	case types.Ethereum:
		return ethereum.StopEthereum(repoRoot)
	default:
		return fmt.Errorf("chain type %s is not supported", chainType)
	}
}
