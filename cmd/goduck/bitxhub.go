package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/meshplus/goduck/internal/download"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

const (
	BINARY = "binary"
	SOLO   = "solo"
	SCRIPT = "playground.sh"

	LinuxWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.so"
	MacOSWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.dylib"

	BitxhubUrlLinux = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc3/bitxhub_linux_amd64.tar.gz"
	BitxhubUrlMacOS = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc3/bitxhub_macos_x86_64.tar.gz"
)

func bitxhubCMD() *cli.Command {
	return &cli.Command{
		Name:  "bitxhub",
		Usage: "start or stop BitXHub nodes",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start bitxhub nodes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Value: BINARY,
						Usage: "configuration type, one of binary or docker",
					},
					&cli.StringFlag{
						Name:  "mode",
						Value: SOLO,
						Usage: "configuration mode, one of solo or cluster",
					},
					&cli.Uint64Flag{
						Name:  "num",
						Value: 4,
						Usage: "Node number, only useful in cluster mode, ignored in solo mode",
					},
				},
				Action: startBitXHub,
			},
			{
				Name:   "stop",
				Usage:  "Stop bitxhub nodes",
				Action: stopBitXHub,
			},
			{
				Name:  "config",
				Usage: "Generate configuration for BitXHub nodes",
				Flags: []cli.Flag{
					&cli.Uint64Flag{
						Name:  "num",
						Value: 4,
						Usage: "Node number, only useful in cluster mode, ignored in solo mode",
					},
					&cli.StringFlag{
						Name:  "type",
						Value: BinaryMode,
						Usage: "configuration type, one of binary or docker",
					},
					&cli.StringFlag{
						Name:  "mode",
						Value: ClusterMode,
						Usage: "configuration mode, one of solo or cluster",
					},
					&cli.StringSliceFlag{
						Name:  "ips",
						Usage: "node IPs, use 127.0.0.1 for all nodes by default",
					},
					&cli.StringFlag{
						Name:  "target",
						Value: ".",
						Usage: "where to put the generated configuration files",
					},
				},
				Action: generateConfig,
			},
		},
	}
}

func stopBitXHub(ctx *cli.Context) error {
	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoPath, SCRIPT)) {
		return fmt.Errorf("please `goduck init` first")
	}
	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, SCRIPT), "down")

	return execCmd(args, repoPath)
}

func startBitXHub(ctx *cli.Context) error {
	num := ctx.Int("num")
	typ := ctx.String("type")
	mode := ctx.String("mode")

	repoPath, err := repo.PathRoot()
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoPath, SCRIPT)) {
		return fmt.Errorf("please `goduck init` first")
	}

	ips := make([]string, 0)
	err = InitConfig(typ, mode, repoPath, num, ips)
	if err != nil {
		return fmt.Errorf("init config error:%w", err)
	}

	if typ == BINARY {
		err := downloadBinary(repoPath)
		if err != nil {
			return fmt.Errorf("download binary error:%w", err)
		}
	}

	args := make([]string, 0)
	args = append(args, filepath.Join(repoPath, SCRIPT), "up")
	args = append(args, mode, typ, strconv.Itoa(num))
	return execCmd(args, repoPath)
}

func downloadBinary(repoPath string) error {
	root := filepath.Join(repoPath, "bin")
	if !fileutil.Exist(root) {
		err := os.Mkdir(root, 0755)
		if err != nil {
			return err
		}
	}

	if !fileutil.Exist(filepath.Join(root, "bitxhub")) {
		if runtime.GOOS == "linux" {
			err := download.Download(root, BitxhubUrlLinux)
			if err != nil {
				return err
			}
			if !fileutil.Exist(filepath.Join(root, "libwasmer.so")) {
				err := download.Download(root, LinuxWasmLibUrl)
				if err != nil {
					return err
				}
			}
		}
		if runtime.GOOS == "darwin" {
			err := download.Download(root, BitxhubUrlMacOS)
			if err != nil {
				return err
			}
			if !fileutil.Exist(filepath.Join(root, "libwasmer.dylib")) {
				err := download.Download(root, MacOSWasmLibUrl)
				if err != nil {
					return err
				}
			}
		}
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
