package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/fatih/color"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

func deployCMD() *cli.Command {
	return &cli.Command{
		Name:  "deploy",
		Usage: "Deploy bitxhub and pier",
		Subcommands: []*cli.Command{
			{
				Name:  "bitxhub",
				Usage: "Deploy bitxhub to remote server.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "ips",
						Usage:    "servers ip, e.g. 188.0.0.1,188.0.0.2,188.0.0.3,.188.0.0.4",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "username,u",
						Usage:    "server username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "version,v",
						Usage:    "bitxhub version",
						Required: true,
					},
				},
				Action: deployBitXHub,
			},
		},
	}
}

func deployBitXHub(ctx *cli.Context) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	username := ctx.String("username")
	version := ctx.String("version")

	data, err := ioutil.ReadFile(filepath.Join(repoRoot, "release.json"))
	if err != nil {
		return err
	}

	var release *Release
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport bitxhub verison")
	}

	target := filepath.Join(repoRoot, "config")

	err = os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "bitxhub")
	if err != nil {
		return err
	}

	ips := strings.Split(ctx.String("ips"), ",")

	generator := NewBitXHubConfigGenerator("binary", "cluster", dir, len(ips), ips, "")

	if err := generator.InitConfig(); err != nil {
		return err
	}

	binPath := filepath.Join(repoRoot, "bin")

	err = os.MkdirAll(binPath, os.ModePerm)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("bitxhub_linux-amd64_%s.tar.gz", version)
	filePath := filepath.Join(binPath, filename)

	if !fileutil.Exist(filePath) {
		url := fmt.Sprintf("https://github.com/meshplus/bitxhub/releases/download/%s/bitxhub_linux-amd64_%s.tar.gz", version, version)
		err = download.Download(binPath, url)
		if err != nil {
			return err
		}
	}

	for idx, ip := range ips {
		color.Blue("====> Operating at node%d\n", idx+1)
		who := fmt.Sprintf("%s@%s", username, ip)
		target := fmt.Sprintf("%s:~/", who)

		err = sh.
			Command("scp", "-r",
				fmt.Sprintf("%s/node%d", dir, idx+1),
				fmt.Sprintf("%s%s", target, "build")).
			Command("scp", filePath, target).
			Run()
		if err != nil {
			return err
		}

		err = sh.
			Command("ssh", who, "export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/build").
			Command("ssh", who, fmt.Sprintf("tar xzf %s && mkdir -p build/node%d/plugins && cp build/raft.so build/node%d/plugins", filename, idx+1, idx+1)).Run()
		if err != nil {
			return err
		}
	}

	color.Blue("====> Run\n")
	for idx, ip := range ips {
		who := fmt.Sprintf("%s@%s", username, ip)
		err = sh.Command("ssh", who, "export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/build").
			Command("ssh", who,
				fmt.Sprintf("cd ~/build && nohup bitxhub --repo node%d start &", idx+1)).Run()
		if err != nil {
			return err
		} else {
			color.Green("====> Start bitxhub node%d successful\n", idx+1)
		}
	}

	return nil
}
