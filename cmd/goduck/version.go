package main

import (
	"fmt"

	"github.com/meshplus/goduck"
	"github.com/urfave/cli/v2"
)

func getVersionCMD() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Goduck version",
		Action: version,
	}
}

func version(ctx *cli.Context) error {
	printVersion()

	return nil
}

func printVersion() {
	fmt.Printf("Goduck version: %s-%s-%s\n", goduck.CurrentVersion, goduck.CurrentBranch, goduck.CurrentCommit)
	fmt.Printf("App build date: %s\n", goduck.BuildDate)
	fmt.Printf("System version: %s\n", goduck.Platform)
	fmt.Printf("Golang version: %s\n", goduck.GoVersion)
}
