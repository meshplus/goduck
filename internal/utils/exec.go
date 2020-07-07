package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type ExecTask struct {
	Command string
	Args    []string
	Env     []string
	Repo    string

	StreamStdio  bool
	PrintCommand bool
}

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func ExecuteShell(args []string, repoRoot string) error {
	task := ExecTask{
		Command:      "/bin/bash",
		Args:         args,
		Repo:         repoRoot,
		StreamStdio:  true,
		PrintCommand: true,
	}
	result, err := task.Execute()
	if err != nil {
		return fmt.Errorf("execute shell error:%w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("execute shell error:%s", result.Stderr)
	}
	return nil
}

func (et ExecTask) Execute() (ExecResult, error) {
	argsSt := ""
	if len(et.Args) > 0 {
		argsSt = strings.Join(et.Args, " ")
	}

	if et.PrintCommand {
		fmt.Println("exec: ", et.Command, argsSt)
	}

	var cmd *exec.Cmd

	if strings.Index(et.Command, " ") > 0 {
		parts := strings.Split(et.Command, " ")
		command := parts[0]
		args := parts[1:]
		cmd = exec.Command(command, args...)

	} else {
		cmd = exec.Command(et.Command, et.Args...)
	}

	cmd.Dir = et.Repo

	if len(et.Env) > 0 {
		overrides := map[string]bool{}
		for _, env := range et.Env {
			key := strings.Split(env, "=")[0]
			overrides[key] = true
			cmd.Env = append(cmd.Env, env)
		}

		for _, env := range os.Environ() {
			key := strings.Split(env, "=")[0]

			if _, ok := overrides[key]; !ok {
				cmd.Env = append(cmd.Env, env)
			}
		}
	}

	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}

	var stdoutWriters io.Writer
	var stderrWriters io.Writer

	if et.StreamStdio {
		stdoutWriters = io.MultiWriter(os.Stdout, &stdoutBuff)
		stderrWriters = io.MultiWriter(os.Stderr, &stderrBuff)
	} else {
		stdoutWriters = &stdoutBuff
		stderrWriters = &stderrBuff
	}

	cmd.Stdout = stdoutWriters
	cmd.Stderr = stderrWriters
	startErr := cmd.Start()

	if startErr != nil {
		return ExecResult{}, startErr
	}

	exitCode := 0
	execErr := cmd.Wait()
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return ExecResult{
		Stdout:   string(stdoutBuff.Bytes()),
		Stderr:   string(stderrBuff.Bytes()),
		ExitCode: exitCode,
	}, nil
}
