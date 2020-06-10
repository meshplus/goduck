package ethereum

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

const (
	SCRIPT = "private_chain.sh"
)

func GetEtherCMD() *cli.Command {
	return &cli.Command{
		Name:  "ether",
		Usage: "operation about ethereum chain",
		Subcommands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "start a ethereum chain",
				Action: startEther,
			},
			{
				Name:   "stop",
				Usage:  "Stop ethereum chain",
				Action: stopEther,
			},
			contractCMD,
		},
	}
}

func startEther(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if err := StartEthereum(repoRoot, "binary"); err != nil {
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

	fmt.Printf("start ethereum private chain with data directory in %s/ethereum/datadir.\n", repoRoot)
	return nil
}

func StopEthereum(repoPath string) error {
	if !fileutil.Exist(filepath.Join(repoPath, SCRIPT)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return execCmd([]string{SCRIPT, "down"}, repoPath)
}

func StartEthereum(repo, mode string) error {
	switch mode {
	case "binary":
		args := []string{SCRIPT, "binary"}
		if err := execCmd(args, repo); err != nil {
			return err
		}
	case "docker":
		args := []string{SCRIPT, "docker"}
		if err := execCmd(args, repo); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not supported mode")
	}
	return nil
}

func execCmd(args []string, repoRoot string) error {
	cmd := exec.Command("/bin/bash", args...)
	cmd.Dir = repoRoot
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("execute command: %s", err.Error())
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	errStr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("execute command error:%w", err)
	}
	fmt.Println(string(errStr))
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait for command to finish: %s", err.Error())
	}
	return nil
}
