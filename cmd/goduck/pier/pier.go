package pier

import (
	"fmt"

	"github.com/meshplus/goduck/cmd/goduck/ethereum"
	"github.com/meshplus/goduck/cmd/goduck/fabric"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

func StartPier(repoRoot, chainType, chainUpType, pierUpType string, version string) error {
	// start appchain first
	switch chainUpType {
	case types.TypeBinary:
		switch chainType {
		case types.Fabric:
			return fmt.Errorf("fabric is not supported to start up with binary")
		case types.Ethereum:
			if err := ethereum.StartEthereum(repoRoot, types.TypeBinary); err != nil {
				return err
			}
		default:
			return fmt.Errorf("chain type %s is not supported", chainType)
		}
	case types.TypeDocker:
		switch chainType {
		case types.Fabric:
			if err := fabric.Start(repoRoot, ""); err != nil {
				return err
			}
		case types.Ethereum:
			if err := ethereum.StartEthereum(repoRoot, types.TypeDocker); err != nil {
				return err
			}
		default:
			return fmt.Errorf("chain type %s is not supported", chainType)
		}
	default:
		return fmt.Errorf("appchain start up type %s is not supported", chainUpType)
	}
	// start pier with appchain plugin, install chaincode if appchain is fabric
	if chainType == types.Fabric {
		if err := utils.ExecuteShell([]string{types.ChaincodeScript, "install"}, repoRoot); err != nil {
			return err
		}
	}

	args := []string{types.PierScript, "up", "-m", chainType, "-t", pierUpType,
		"-r", ".pier_" + chainType, "-v", version}
	return utils.ExecuteShell(args, repoRoot)
}

func StopPier(repoRoot, chainType string, pierOnly bool) error {
	// stop pier first, then appchain
	args := []string{types.PierScript, "down", "-m", chainType}
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
