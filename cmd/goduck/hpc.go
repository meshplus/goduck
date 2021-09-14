package main

import (
	"github.com/urfave/cli/v2"
)

var hpcCMD = &cli.Command{
	Name:  "hpc",
	Usage: "Interact with hyperchain contract about deploying, invoking, upgrading",
	//Subcommands: cli.Commands{
	//	hpcDeployCMD,
	//	hpcInvokeCMD,
	//	hpcUpdateCMD,
	//	hpcStartCMD,
	//},
}

//var hpcDeployCMD = &cli.Command{
//	Name:  "deploy",
//	Usage: "Deploy solidity/java/jvm contract",
//	Flags: []cli.Flag{
//		&cli.StringFlag{
//			Name:     "path,p",
//			Usage:    "contract code path",
//			Required: true,
//		},
//		&cli.StringFlag{
//			Name:  "type,t",
//			Value: "solc",
//			Usage: "contract type: solc/java/jvm, default => solc",
//		},
//		&cli.BoolFlag{
//			Name:  "local,l",
//			Usage: "local compile",
//		},
//	},
//	Action: func(ctx *cli.Context) error {
//		path := ctx.String("path")
//		typ := ctx.String("type")
//		local := ctx.Bool("local")
//		//root := ctx.GlobalString("repo")
//		//if root != "" {
//		//	repo.PathRootVar = root
//		//}
//
//		return hpc.DeployContract(path, typ, local)
//	},
//}
//
//var hpcUpdateCMD = cli.Command{
//	Name:  "update",
//	Usage: "Update solidity/java contract",
//	Flags: []cli.Flag{
//		&cli.StringFlag{
//			Name:     "path,p",
//			Usage:    "contract code path",
//			Required: true,
//		},
//		&cli.StringFlag{
//			Name:  "type,t",
//			Value: "solc",
//			Usage: "contract type: solc/hvm/jvm, default => solc",
//		},
//		&cli.BoolFlag{
//			Name:  "local,l",
//			Usage: "local compile",
//		},
//		&cli.StringFlag{
//			Name:     "addr,a",
//			Usage:    "contract address",
//			Required: true,
//		},
//	},
//
//	Action: func(ctx *cli.Context) error {
//		path := ctx.String("path")
//		local := ctx.Bool("local")
//		addr := ctx.String("addr")
//		typ := ctx.String("type")
//
//		//root := ctx.GlobalString("repo")
//		//if root != "" {
//		//	repo.PathRootVar = root
//		//}
//
//		switch typ {
//		case "hvm":
//			bin, err := hvm.ReadJar(path)
//			if err != nil {
//				return err
//			}
//
//			return hpc.UpdateContract(bin, addr, rpc.HVM)
//		case "jvm":
//			bin, err := java.ReadJavaContract(path)
//			if err != nil {
//				return err
//			}
//
//			return hpc.UpdateContract(bin, addr, rpc.JVM)
//		case "solc":
//			code, err := ioutil.ReadFile(path)
//			if err != nil {
//				return err
//			}
//
//			ret, err := hpc.CompileContract(string(code), local)
//			if err != nil {
//				return err
//			}
//
//			bin, err := hpc.MainContractBin(ret)
//			if err != nil {
//				return err
//			}
//
//			return hpc.UpdateContract(bin, addr, rpc.EVM)
//		default:
//			return fmt.Errorf("not support contract type: %s", typ)
//		}
//	},
//}
//
//var hpcInvokeCMD = cli.Command{
//	Name:  "invoke",
//	Usage: "Invoke solidity/java/jvm contract",
//	Flags: []cli.Flag{
//		&cli.StringFlag{
//			Name:  "path,p",
//			Usage: "solidity abi path",
//		},
//		&cli.StringFlag{
//			Name:  "type,t",
//			Value: "solc",
//			Usage: "contract type: solc/hvm/jvm, default => solc",
//		},
//	},
//	Action: func(ctx *cli.Context) error {
//		typ := ctx.String("type")
//		path := ctx.String("path")
//
//		//root := ctx.GlobalString("repo")
//		//if root != "" {
//		//	repo.PathRootVar = root
//		//}
//
//		if typ != "jvm" {
//			if path == "" {
//				return fmt.Errorf("invoke solc/java contract must include abi path")
//			}
//		}
//
//		if ctx.NArg() < 2 {
//			return fmt.Errorf("invoke contract must include address and function")
//		}
//
//		args := ctx.Args()
//
//		return hpc.InvokeContract(path, typ, args.Get(0), args.Get(1), args.Get(2))
//	},
//}
//
//var hpcStartCMD = cli.Command{
//	Name:  "start",
//	Usage: "Start a hyperchain",
//	Flags: []cli.Flag{
//		&cli.StringFlag{
//			Name:     "mode",
//			Usage:    "configuration mode, one of solo or cluster",
//			Value:    "solo",
//			Required: false,
//		},
//		&cli.StringFlag{
//			Name:     "ips",
//			Usage:    "servers ip, solo e.g. 188.0.0.1  cluster e.g. 188.0.0.1,188.0.0.2,188.0.0.3,188.0.0.4",
//			Required: true,
//		},
//		&cli.StringFlag{
//			Name:     "username,u",
//			Usage:    "server username",
//			Required: true,
//		},
//		&cli.StringFlag{
//			Name:     "filepath,fp",
//			Usage:    "installation file path",
//			Required: true,
//		},
//		&cli.StringFlag{
//			Name:     "certspath,cp",
//			Usage:    "certs file path",
//			Required: true,
//		},
//		&cli.StringFlag{
//			Name:     "licensepath,lp",
//			Usage:    "license file path",
//			Required: true,
//		},
//	},
//	Action: func(ctx *cli.Context) error {
//		mode := ctx.String("mode")
//		ips := strings.Split(ctx.String("ips"), ",")
//		if mode == "solo" {
//			if len(ips) != 1 {
//				return fmt.Errorf("deploy solo hyperchain need 1 server ip but not")
//			}
//		} else {
//			if len(ips) != 4 {
//				return fmt.Errorf("deploy cluster hyperchain need 4 servers ip but not")
//			}
//		}
//
//		username := ctx.String("username")
//		fpath := ctx.String("filepath")
//		fname := filepath.Base(fpath)
//		cpath := ctx.String("certspath")
//		lpath := ctx.String("licensepath")
//		lname := filepath.Base(lpath)
//		return hpc.Start(mode, ips, username, fpath, fname, cpath, lpath, lname)
//	},
//}
