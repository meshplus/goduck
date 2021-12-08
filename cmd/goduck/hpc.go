package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/hpc"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

var hpcCMD = &cli.Command{
	Name:  "hpc",
	Usage: "Interact with hyperchain contract about deploying, invoking, upgrading",
	Subcommands: cli.Commands{
		&hpcDeployCMD,
		&hpcInvokeCMD,
		&hpcUpdateCMD,
		//&hpcStartCMD,
	},
}

var hpcDeployCMD = cli.Command{
	Name:  "deploy",
	Usage: "Deploy solidity/java/jvm contract",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "specify hyperchain config path. It should be hpc.account, password, hpc.toml, certs in the catalog",
		},
		&cli.StringFlag{
			Name:     "code-path",
			Usage:    "specify contract code path",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "type,t",
			Value: "solc",
			Usage: "specify contract type: solc/java/jvm, default => solc",
		},
		&cli.BoolFlag{
			Name:  "local,l",
			Usage: "specify whether to compile locally",
		},
	},
	Action: func(ctx *cli.Context) error {
		configPath := ctx.String("config-path")
		codePath := ctx.String("code-path")
		typ := ctx.String("type")
		local := ctx.Bool("local")
		args := ""

		if configPath == "" {
			repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
			if err != nil {
				return err
			}
			configPath = filepath.Join(repoRoot, "hyperchain")
			if !fileutil.Exist(configPath) {
				return fmt.Errorf("please `goduck init` first")
			}
		}

		if ctx.NArg() > 0 {
			args = ctx.Args().Get(0)
		}
		return hpc.Deploy(configPath, codePath, typ, local, args)
	},
}

var hpcUpdateCMD = cli.Command{
	Name:  "update",
	Usage: "Update solidity/java contract",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "specify hyperchain config path. It should be hpc.account, password, hpc.toml, certs in the catalog",
		},
		&cli.StringFlag{
			Name:  "code-path",
			Usage: "specify solidity abi path",
		},
		&cli.StringFlag{
			Name:  "type,t",
			Value: "solc",
			Usage: "specify contract type: solc/hvm/jvm, default => solc",
		},
		&cli.BoolFlag{
			Name:  "local,l",
			Usage: "specify whether to compile locally",
		},
		&cli.StringFlag{
			Name:     "contract-addr",
			Usage:    "specify contract address",
			Required: true,
		},
	},

	Action: func(ctx *cli.Context) error {
		configPath := ctx.String("config-path")
		codePath := ctx.String("code-path")
		typ := ctx.String("type")
		local := ctx.Bool("local")
		conAddr := ctx.String("contract-addr")

		if configPath == "" {
			repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
			if err != nil {
				return err
			}
			configPath = filepath.Join(repoRoot, "hyperchain")
			if !fileutil.Exist(configPath) {
				return fmt.Errorf("please `goduck init` first")
			}
		}

		return hpc.Update(configPath, codePath, typ, local, conAddr)
	},
}

var hpcInvokeCMD = cli.Command{
	Name:  "invoke",
	Usage: "Invoke solidity/java/jvm contract",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "specify hyperchain config path. It should be hpc.account, password, hpc.toml, certs in the catalog",
		},
		&cli.StringFlag{
			Name:  "abi-path",
			Usage: "specify solidity abi path",
		},
		&cli.StringFlag{
			Name:  "type,t",
			Value: "solc",
			Usage: "specify contract type: solc/hvm/jvm, default => solc",
		},
	},
	Action: func(ctx *cli.Context) error {
		configPath := ctx.String("config-path")
		abiPath := ctx.String("abi-path")
		typ := ctx.String("type")

		if typ != "jvm" {
			if abiPath == "" {
				return fmt.Errorf("invoke solc/java contract must include abi path")
			}
		}

		if ctx.NArg() < 2 {
			return fmt.Errorf("invoke contract must include address and function")
		}

		args := ctx.Args()

		if configPath == "" {
			repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
			if err != nil {
				return err
			}
			configPath = filepath.Join(repoRoot, "hyperchain")
			if !fileutil.Exist(configPath) {
				return fmt.Errorf("please `goduck init` first")
			}
		}

		return hpc.Invoke(configPath, abiPath, typ, args.Get(0), args.Get(1), args.Get(2))
	},
}

var hpcStartCMD = cli.Command{
	Name:  "start",
	Usage: "Start a hyperchain",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "mode",
			Usage:    "configuration mode, one of solo or cluster",
			Value:    "solo",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "ips",
			Usage:    "servers ip, solo e.g. 188.0.0.1  cluster e.g. 188.0.0.1,188.0.0.2,188.0.0.3,188.0.0.4",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "username,u",
			Usage:    "server username",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "filepath,fp",
			Usage:    "installation file path",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "certspath,cp",
			Usage:    "certs file path",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "licensepath,lp",
			Usage:    "license file path",
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		mode := ctx.String("mode")
		ips := strings.Split(ctx.String("ips"), ",")
		if mode == "solo" {
			if len(ips) != 1 {
				return fmt.Errorf("deploy solo hyperchain need 1 server ip but not")
			}
		} else {
			if len(ips) != 4 {
				return fmt.Errorf("deploy cluster hyperchain need 4 servers ip but not")
			}
		}

		username := ctx.String("username")
		fpath := ctx.String("filepath")
		fname := filepath.Base(fpath)
		cpath := ctx.String("certspath")
		lpath := ctx.String("licensepath")
		lname := filepath.Base(lpath)
		return hpc.Start(mode, ips, username, fpath, fname, cpath, lpath, lname)
	},
}
