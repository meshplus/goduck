package main

import (
	"github.com/urfave/cli/v2"
)

var fabricCMD = &cli.Command{
	Name:  "fabric",
	Usage: "Interact with fabric contract about invoke and querying upgrading",
	//Subcommands: cli.Commands{
	//	{
	//		Name:      "invoke",
	//		Usage:     "Invoke fabric chaincode",
	//		ArgsUsage: "command: snake fabric invoke [chaincode_id] [function] [args(optional)]",
	//		Action: func(ctx *cli.Context) error {
	//			root := ctx.GlobalString("repo")
	//			if root != "" {
	//				repo.PathRootVar = root
	//			}
	//
	//			args := ctx.Args()
	//			if len(args) < 2 {
	//				return fmt.Errorf("args must be (chaincode_id function args[optional])")
	//			}
	//
	//			if len(args) == 2 {
	//				args = append(args, "")
	//			}
	//
	//			return fabric.Invoke(args[0], args[1], args[2])
	//		},
	//	},
	//	{
	//
	//		Name:  "query",
	//		Usage: "query fabric chaincode",
	//		Action: func(ctx *cli.Context) error {
	//			root := ctx.GlobalString("repo")
	//			if root != "" {
	//				repo.PathRootVar = root
	//			}
	//
	//			args := ctx.Args()
	//			if len(args) < 2 {
	//				return fmt.Errorf("args must be (cid function args[options])")
	//			}
	//
	//			if len(args) == 2 {
	//				args = append(args, "")
	//			}
	//
	//			return fabric.Query(args[0], args[1], args[2])
	//		},
	//	},
	//	{
	//		Name:  "deploy",
	//		Usage: "deploy fabric chaincode",
	//		Flags: []cli.Flag{
	//			cli.StringFlag{
	//				Name:     "ccid",
	//				Usage:    "chaincode id",
	//				Required: true,
	//			},
	//			cli.StringFlag{
	//				Name:     "path",
	//				Usage:    "chaincode path",
	//				Required: true,
	//			},
	//			cli.StringFlag{
	//				Name:     "version",
	//				Usage:    "chaincode version",
	//				Required: true,
	//			},
	//		},
	//		Action: func(ctx *cli.Context) error {
	//			ccid := ctx.String("ccid")
	//			path := ctx.String("path")
	//			version := ctx.String("version")
	//
	//			root := ctx.GlobalString("repo")
	//			if root != "" {
	//				repo.PathRootVar = root
	//			}
	//
	//			return fabric.Deploy(ccid, path, version)
	//		},
	//	},
	//},
}
