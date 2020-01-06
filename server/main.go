package main

import (
	"fmt"
	"os"

	"github.com/coderun-top/coderun/server/version"
	"github.com/coderun-top/coderun/server"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "coderun-server"
	app.Version = version.Version.String()
	app.Usage = "coderun server"
	app.Action = server
	app.Flags = flags
	app.Before = before

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
