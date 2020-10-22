package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/meshplus/goduck/internal/repo"

	"github.com/meshplus/goduck"
	"github.com/urfave/cli/v2"
)

func getVersionCMD() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Components version",
		Subcommands: []*cli.Command{
			{
				Name:   "all",
				Usage:  "All components version",
				Action: allVersion,
			},
			{
				Name:   "goduck",
				Usage:  "Goduck version",
				Action: goduckVersion,
			},
		},
	}
}

func allVersion(ctx *cli.Context) error {
	root, err := repo.PathRoot()
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(filepath.Join(root, "release.json"))
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func goduckVersion(ctx *cli.Context) error {
	printVersion()

	return nil
}

func printVersion() {
	fmt.Printf("Goduck version: %s-%s-%s\n", goduck.CurrentVersion, goduck.CurrentBranch, goduck.CurrentCommit)
	fmt.Printf("App build date: %s\n", goduck.BuildDate)
	fmt.Printf("System version: %s\n", goduck.Platform)
	fmt.Printf("Golang version: %s\n", goduck.GoVersion)
}
