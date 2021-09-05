package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
							err = parse_log(_log)
							if err != nil {
								fmt.Println(err)
								os.Exit(1)
							}

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

type ExtraceEvent struct {
	PID        int64
	Started    time.Time
	Ended      time.Time
	ExitCode   int
	Duration   string
	EventType  string
	User       string
	Exec       string
	Args       []string
	ExecString string
}

var extrace_events = map[int64]ExtraceEvent{}

func ParseLine(line string) error {
	items := strings.Split(line, ` `)
	if len(items) < 3 {
		return errors.New(fmt.Sprintf("Invalid line: %s", line))
	}
	et := ``
	eu := ``
	ES := ``
	EE := ``
	EA := []string{}
	if strings.Contains(items[0], `+`) {
		et = `start`
		if !strings.Contains(items[1], `<`) || !strings.Contains(items[1], `>`) {
			return errors.New(fmt.Sprintf("Invalid start line: %s", line))
		}
		eu = strings.Replace(items[1], `<`, ``, 1)
		eu = strings.Replace(eu, `>`, ``, 1)
		EE = items[2]
		if len(items) == 4 {
			EA = []string{items[3]}
		} else if len(items) > 4 {
			EA = items[3 : len(items)-1]
		}
		ES = fmt.Sprintf(`%s %s`, EE, strings.Join(EA, " "))
	} else if strings.Contains(items[0], `-`) {
		et = `end`
	} else {
		return errors.New(fmt.Sprintf("Invalid line: %s", line))
	}

	_pid := strings.Replace(strings.Replace(items[0], `-`, ``, 1), `+`, ``, 1)
	__pid, err := strconv.ParseInt(_pid, 10, 0)
	if err == nil && __pid > 0 {
		extrace_events[__pid] = ExtraceEvent{
			PID:        __pid,
			EventType:  et,
			User:       eu,
			Exec:       EE,
			Args:       EA,
			ExecString: ES,
		}
	}
	return nil
}

func parse_log(log_data []byte) error {
	show_qty := 20
	lines := strings.Split(string(log_data), "\n")
	for l_no, l := range lines {
		if l_no < show_qty || l_no > (len(lines)-show_qty) {
			pp.Println(l_no, l)
		}
		ParseLine(l)
	}
	qty := 0
	for _, E := range extrace_events {
		if qty > 5 {
			break
		}
		qty += 1

		pp.Println(E)

	}
	return nil
}
func loadExtraceLogFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)

}
