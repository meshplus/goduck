package ethereum

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/urfave/cli/v2"
)

func GetEtherCMD() *cli.Command {
	return &cli.Command{
		Name:  "ether",
		Usage: "operation about ethereum chain",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start a ethereum chain",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "account_path",
						Usage:    "path to store accounts file",
						Required: true,
					},
				},
				Action: start,
			},
			contractCMD,
		},
	}
}

func start(ctx *cli.Context) error {
	accountPath := ctx.String("account_path")
	args := []string{"-a", "1", "--acctKeys", accountPath}

	cmd := exec.Command("ganache-cli", args...)
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
