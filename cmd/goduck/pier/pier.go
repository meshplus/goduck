package pier

import (
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
)

func RegisterPier(repoRoot, chainType, cryptoPath, pierUpType, version, tls, http, pprof, api, overwrite, appchainIP, appchainAddr, appchainPorts, appchainContractAddr, pierRepo, adminKey, method string) error {
	args := []string{types.PierScript, "register", "-m", chainType, "-t", pierUpType,
		"-v", version, "-c", cryptoPath, "-f", pprof, "-a", api, "-l", tls, "-p", http,
		"-o", overwrite, "-i", appchainIP, "-d", appchainAddr, "-s", appchainPorts,
		"-n", appchainContractAddr, "-r", pierRepo, "-k", adminKey, "-e", method}
	return utils.ExecuteShell(args, repoRoot)
}

func DeployRule(repoRoot, chainType, pierRepo, ruleRepo, pierUpType, adminKey, method, version string) error {
	args := []string{types.PierScript, "rule", "-m", chainType, "-r", pierRepo, "-u", ruleRepo, "-t", pierUpType, "-k", adminKey, "-e", method, "-v", version}
	return utils.ExecuteShell(args, repoRoot)
}

func StartPier(repoRoot, chainType, cryptoPath, pierUpType, version, tls, http, pprof, api, overwrite, appchainIP, appchainAddr, appchainPorts, appchainContractAddr, pierRepo string) error {
	args := []string{types.PierScript, "up", "-m", chainType, "-t", pierUpType,
		"-v", version, "-c", cryptoPath, "-f", pprof, "-a", api, "-l", tls, "-p",
		http, "-o", overwrite, "-i", appchainIP, "-d", appchainAddr, "-s", appchainPorts,
		"-n", appchainContractAddr, "-r", pierRepo}
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
