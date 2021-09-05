package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "extrace-parser",
		Usage: "Extrace Log Parser",
		Commands: []*cli.Command{
			{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Log",
				Subcommands: []*cli.Command{
					{
						Name:      "parse",
						Aliases:   []string{"p"},
						Usage:     "Parse Extrace Log to JSON",
						ArgsUsage: "FILE",
						Action: func(c *cli.Context) error {
							pp.Println(c.Args())
							_log, err := loadExtraceLogFile(c.Args().Get(0))
							if err != nil {
								fmt.Println(err)
								os.Exit(1)
							}
							fmt.Println(len(_log), " bytes")
							fmt.Println(len(strings.Split(string(_log), "\n")), " lines")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func loadExtraceLogFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)

}
