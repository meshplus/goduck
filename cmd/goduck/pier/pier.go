package pier

import (
	"fmt"

	"github.com/meshplus/goduck/cmd/goduck/ethereum"
	"github.com/meshplus/goduck/cmd/goduck/fabric"
	"github.com/meshplus/goduck/internal/types"
)

func StartAppchain(repoRoot, chainType, mode string) error {
	switch mode {
	case types.TypeBinary:
		switch chainType {
		case types.Fabric:
			return fmt.Errorf("fabric is not supported to start up with binary")
		case types.Ethereum:
			return ethereum.StartEthereum(repoRoot, types.TypeBinary)
		default:
			return fmt.Errorf("chain type %s is not supported", chainType)
		}
	case types.TypeDocker:
		switch chainType {
		case types.Fabric:
			return fabric.Start(repoRoot)
		case types.Ethereum:
			return ethereum.StartEthereum(repoRoot, types.TypeDocker)
		default:
			return fmt.Errorf("chain type %s is not supported", chainType)
		}
	default:
		return fmt.Errorf("start up mode %s is not supported", chainType)
	}
}

func StopAppchain(repoRoot, chainType string) error {
	switch chainType {
	case types.Fabric:
		return fabric.Stop(repoRoot)
	case types.Ethereum:
		return ethereum.StopEthereum(repoRoot)
	default:
		return fmt.Errorf("chain type %s is not supported", chainType)
	}
}
