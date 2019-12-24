package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"git.quba.fr/qbarrand/quba.fr-server/pkg"
)

func main() {
	var (
		addr    string
		dir     string
		quality uint
	)

	app := cli.NewApp()

	app.Name = "server"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "addr",
			Usage:       "the address and port on which this server should listen",
			EnvVar:      "ADDR",
			Value:       ":8080",
			Destination: &addr,
		},
		cli.StringFlag{
			Name:        "dir",
			Usage:       "path to the served directory",
			EnvVar:      "DIR",
			Value:       ".",
			Destination: &dir,
		},
		cli.UintFlag{
			Name:        "quality",
			Usage:       "quality of the output JPG file",
			EnvVar:      "QUALITY",
			FilePath:    "",
			Value:       80,
			Destination: &quality,
		},
	}

	app.Action = func(_ *cli.Context) error {
		log.Print("Serving contents from " + dir)
		log.Print("Starting the server on " + addr)

		return pkg.StartServer(addr, dir, quality)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
