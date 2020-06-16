package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
)

func ExecCmd(args []string, repoRoot string) error {
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
