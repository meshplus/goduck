package ethereum

import (
	"fmt"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)


func startEther(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if err := StartEthereum(repoRoot, ctx.String("type")); err != nil {
		return err
	}

	fmt.Printf("start ethereum private chain with data directory in %s/ethereum/datadir.\n", repoRoot)
	return nil
}

func stopEther(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if err := StopEthereum(repoRoot); err != nil {
		return err
	}

	fmt.Println("Stop ethereum private chain")
	return nil
}

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
