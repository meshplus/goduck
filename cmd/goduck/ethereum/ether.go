package ethereum

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

func GetEtherCMD() *cli.Command {
	return &cli.Command{
		Name:  "ether",
		Usage: "operation about ethereum chain",
		Subcommands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "start a ethereum chain",
				Action: start,
			},
			contractCMD,
		},
	}
}

func start(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if err := StartEthereum(repoRoot, "binary"); err != nil {
		return err
	}

	fmt.Printf("start ethereum private chain with data directory in %s/datadir.\n", repoRoot)
	return nil
}

func StartEthereum(repo, mode string) error {
	switch mode {
	case "binary":
		args := []string{"private_chain.sh", "binary"}
		if err := execCmd(args, repo); err != nil {
			return err
		}
	case "docker":
		args := []string{"private_chain.sh", "docker"}
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
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("execute command: %s", err.Error())
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait for command to finish: %s", err.Error())
	}
	return nil
}
