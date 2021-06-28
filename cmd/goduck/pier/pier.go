package pier

import (
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

func GeneratePier(scriptPath, repoRoot, pierRepo, configPath, appchainType, pierBinPath, pluginPath string) error {
	args := []string{scriptPath, "-a", appchainType, "-p", pierRepo, "-c", configPath, "-b", pierBinPath, "-g", pluginPath}
	return utils.ExecuteShell(args, repoRoot)
}

func StartPier(repoRoot, appchainType, pierRepo, upType, configPath, version string) error {
	args := []string{types.PierScript, "up", "-a", appchainType, "-p", pierRepo, "-u", upType, "-c", configPath, "-v", version}
	return utils.ExecuteShell(args, repoRoot)
}

func RegisterPier(repoRoot, pierRepo, appchainType, upType, method, version, cid string) error {
	args := []string{types.PierScript, "register", "-a", appchainType, "-p", pierRepo, "-u", upType, "-m", method, "-v", version, "-i", cid}
	return utils.ExecuteShell(args, repoRoot)
}

func DeployRule(repoRoot, appchainType, pierRepo, ruleRepo, upType, method, version, cid string) error {
	args := []string{types.PierScript, "rule", "-a", appchainType, "-p", pierRepo, "-r", ruleRepo, "-u", upType, "-m", method, "-v", version, "-i", cid}
	return utils.ExecuteShell(args, repoRoot)
}

func StopPier(repoRoot, appchainType string) error {
	args := []string{types.PierScript, "down", "-a", appchainType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}
	return nil
}

func CleanPier(repoRoot, appchainType string) error {
	args := []string{types.PierScript, "clean", "-a", appchainType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}

	return nil
}
