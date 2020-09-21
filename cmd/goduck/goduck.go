package main

import (
	"fmt"
	"os"
	"time"

	"github.com/meshplus/goduck/cmd/goduck/ethereum"
	"github.com/meshplus/goduck/cmd/goduck/fabric"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "GoDuck"
	app.Usage = "GoDuck is a command-line management tool that can help to run BitXHub."
	app.Compiled = time.Now()

	cli.VersionPrinter = func(c *cli.Context) {
		printVersion()
	}

	// global flags
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "repo",
			Usage: "GoDuck storage repo path",
		},
	}

	app.Commands = []*cli.Command{
		deployCMD(),
		getVersionCMD(),
		GetInitCMD(),
		GetStatusCMD(),
		fabric.GetFabricCMD(),
		ethereum.GetEtherCMD(),
		keyCMD(),
		bitxhubCMD(),
		pierCMD,
		playgroundCMD(),
		infoCMD(),
		prometheusCMD(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
