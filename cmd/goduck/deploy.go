package main

import (
	"fmt"
	"github.com/codeskyblue/go-sh"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	target := filepath.Join(repoRoot, "config")

	err = os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return err
	}

	username := ctx.String("username")
	version := ctx.String("version")

	dir, err := ioutil.TempDir("", "bitxhub")
	if err != nil {
		return err
	}

	ips := strings.Split(ctx.String("ips"), ",")

	generator := NewBitXHubConfigGenerator("binary", "cluster", dir, len(ips), ips)

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
		err = download.Download(binPath, "https://github.com/meshplus/bitxhub/releases/download/v1.1.0-rc1/bitxhub_linux-amd64_v1.1.0-rc1.tar.gz")
		if err != nil {
			return err
		}
	}

	//for _, ip := range ips {
	//	err = sh.Command("ssh-copy-id", fmt.Sprintf("%s@%s", username, ip)).Run()
	//	if err != nil {
	//		return err
	//	}
	//}

	for idx, ip := range ips {
		fmt.Printf("Operating at node%d\n", idx+1)
		who := fmt.Sprintf("%s@%s", username, ip)
		target := fmt.Sprintf("%s:~/", who)
		err = sh.Command("scp", filePath, target).
			Command("ssh", who, fmt.Sprintf("`tar xzf %s`", filename)).
			Command("scp", "-r",
				fmt.Sprintf("%s/node%d", dir, idx+1),
				fmt.Sprintf("%s%s", target, "build")).
			Command("ssh", who, "`export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/build`").
			Command("ssh", who, fmt.Sprintf("`mkdir -p build/node%d/plugins`", idx+1)).
			Command("ssh", who, fmt.Sprintf("`cp build/raft.so build/node%d/plugins`", idx+1)).Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Run")
	for idx, ip := range ips {
		who := fmt.Sprintf("%s@%s", username, ip)
		err = sh.Command("ssh", who, "`export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/build`").
			Command("ssh", who,
				fmt.Sprintf("`cd build && nohup ./bitxhub --repo node%d start &`", idx+1)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
