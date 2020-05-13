package main

import (
	"fmt"

	"github.com/meshplus/bitxman"
	"github.com/urfave/cli/v2"
)

func getVersionCMD() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "BitXMan version",
		Action: version,
	}
}

func version(ctx *cli.Context) error {
	printVersion()

	return nil
}

func printVersion() {
	fmt.Printf("Bitxman version: %s-%s-%s\n", bitxman.CurrentVersion, bitxman.CurrentBranch, bitxman.CurrentCommit)
	fmt.Printf("App build date: %s\n", bitxman.BuildDate)
	fmt.Printf("System version: %s\n", bitxman.Platform)
	fmt.Printf("Golang version: %s\n", bitxman.GoVersion)
}
