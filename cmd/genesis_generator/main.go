package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "genesis generator"
	app.Usage = "generate genesis.json file for your private network consensus engine"
	app.Action = runGenesisGenerator
	app.Run(os.Args)
}

func runGenesisGenerator(c *cli.Context) error {
	makeGenerator(c.String("network")).run()
	return nil
}
