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
	"github.com/vektra/tai64n"
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
							//pp.Println(c.Args())
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

var started_extrace_events = map[int64]ExtraceEvent{}
var extrace_events = map[int64]ExtraceEvent{}

type ExtraceEvent struct {
	PID           int64
	Started       time.Time
	Ended         time.Time
	ExitCode      int
	Duration      string
	EventType     string
	User          string
	Exec          string
	Args          []string
	ExecString    string
	EventTai64    time.Time
	EndEventTai64 time.Time
}

func ParseLine(line string) error {
	//	pp.Println(tai64n.ParseTAI64NLabel(`@400000006134298d0ce3b884`).Time())

	items := strings.Split(line, ` `)
	if len(items) < 3 {
		return errors.New(fmt.Sprintf("Invalid line: %s", line))
	}
	et := ``
	eu := ``
	ES := ``
	EE := ``
	ec := -1
	EA := []string{}
	e_t := time.Now()
	if items[0][0] == '@' {
		e_t = tai64n.ParseTAI64NLabel(items[0]).Time()
	}
	if strings.Contains(items[1], `+`) {
		et = `start`
		if !strings.Contains(items[2], `<`) || !strings.Contains(items[2], `>`) {
			return errors.New(fmt.Sprintf("Invalid start line: %s", line))
		}
		eu = strings.Replace(items[2], `<`, ``, 1)
		eu = strings.Replace(eu, `>`, ``, 1)
		EE = items[3]
		if len(items) == 5 {
			EA = []string{items[4]}
		} else if len(items) > 4 {
			EA = items[4 : len(items)-1]
		}
		ES = fmt.Sprintf(`%s %s`, EE, strings.Join(EA, " "))
	} else if strings.Contains(items[1], `-`) {
		et = `end`
		for _, _l := range items {
			if strings.HasPrefix(_l, `status=`) {
				_ec, err := strconv.ParseInt(strings.Split(_l, `=`)[1], 10, 0)
				if err != nil {
					return err
				}
				ec = int(_ec)
				break
			}
		}
	} else {
		return errors.New(fmt.Sprintf("Invalid line: %s", line))
	}

	_pid := strings.Replace(strings.Replace(items[1], `-`, ``, 1), `+`, ``, 1)
	__pid, err := strconv.ParseInt(_pid, 10, 0)
	if err == nil && __pid > 0 {
		__EE := ExtraceEvent{
			PID:        __pid,
			EventType:  et,
			User:       eu,
			Exec:       EE,
			Args:       EA,
			ExecString: ES,
			ExitCode:   ec,
			EventTai64: e_t,
		}
		if __EE.EventType == `start` {
			started_extrace_events[__pid] = __EE
		} else {
			ke, has := started_extrace_events[__EE.PID]
			if has {
				if ke.PID == __EE.PID {
					ke.EndEventTai64 = __EE.EventTai64
				}
			}
		}
	}
	return nil
}

func parse_log(log_data []byte) error {
	start := time.Now()
	show_qty := 20
	lines := strings.Split(string(log_data), "\n")
	for l_no, l := range lines {
		if len(l) == 0 {
			continue
		}
		if l_no < show_qty || l_no > (len(lines)-show_qty) {
			//pp.Println(l_no, l)
		}
		err := ParseLine(l)
		if err != nil {
			panic(err)
		}
	}
	qty := 0
	tqty := 0
	for _, _ = range started_extrace_events {
		tqty = 1
	}
	for _, E := range started_extrace_events {
		qty += 1
		if qty < 10 || qty > (tqty-10) {
			pp.Println(E)
		}
		qty += 1
	}
	msg := fmt.Sprintf("Parsed %d Events from %d Lines in %s", qty, len(lines), time.Since(start))
	fmt.Println(msg)
	return nil
}
func loadExtraceLogFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)

}
