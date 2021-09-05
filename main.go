package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
	"github.com/vektra/tai64n"
)

var LOG_FILE_PATH = ``
var cmdParse = &cobra.Command{
	Use:   "parse [string to parse]",
	Short: "Parse Log File",
	Long: `parse is for parse anything back.
Echo works a lot like print, except it has a child command.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		LOG_FILE_PATH := args[0]
		_log, err := loadExtraceLogFile(LOG_FILE_PATH)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = parse_log(_log)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		tbl()

	},
}

func main() {
	rootCmd.AddCommand(cmdParse)
	rootCmd.Execute()
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

func (EE *ExtraceEvent) GetExecBase() string {
	return filepath.Base(EE.Exec)
}
func ParseLine(line string) error {

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
		//		pp.Println(items)
		if len(items) == 5 {
			EA = []string{items[4]}
		} else if len(items) > 5 {
			EA = items[4:len(items)]
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
		//pp.Println(__EE)
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

func GetExtraceEventPids() (pids []int64) {
	for pid := range started_extrace_events {
		pids = append(pids, pid)
	}
	return pids
}

func parse_log(log_data []byte) error {
	start := time.Now()
	lines := strings.Split(string(log_data), "\n")
	for _, l := range lines {
		if len(l) == 0 {
			continue
		}
		err := ParseLine(l)
		if err != nil {
			panic(err)
		}
	}
	qty := 0
	tqty := 0
	for _, _ = range started_extrace_events {
		tqty += 1
	}
	qty = 0
	for _, E := range started_extrace_events {
		if qty < 10 || qty > (tqty-10) {
			if false {
				pp.Println(E)
			}
		}
		qty += 1
	}
	msg := fmt.Sprintf(`Parsed %d Events from %s containing %s Bytes, %d Lines, %d PIDs in %s.|`,
		qty,
		LOG_FILE_PATH,
		humanize.Bytes(uint64(len(log_data))),
		len(lines),
		len(GetExtraceEventPids()),
		time.Since(start),
	)
	fmt.Println(msg)
	GetExtraceEventExecs()
	//ExtraceEventExecsReport()

	return nil
}

var ExecPIDs = map[string][]int64{}
var PIDsQtyExecs = map[int64]string{}

func ExtraceEventExecsReport() {
	for exec_bin, pids := range ExecPIDs {
		msg := fmt.Sprintf(`%s => %d`, exec_bin, len(pids))
		fmt.Println(msg)
	}
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value > p[j].Value }

func GetExtraceEventExecs() {
	ExecPIDs = map[string][]int64{}
	for _, E := range started_extrace_events {
		_, h := ExecPIDs[E.GetExecBase()]
		if !h {
			ExecPIDs[E.GetExecBase()] = []int64{E.PID}
		} else {
			ExecPIDs[E.GetExecBase()] = append(ExecPIDs[E.GetExecBase()], E.PID)
		}
	}

	p := make(PairList, len(ExecPIDs))

	i := 0
	for k, v := range ExecPIDs {
		p[i] = Pair{k, len(v)}
		i++
	}

	sort.Sort(p)
	exec_qtys = []ExecQty{}
	for _, k := range p {
		exec_qtys = append(exec_qtys, ExecQty{Exec: k.Key, Qty: k.Value})
	}
	//pp.Println(exec_qtys[1:5])
}

var exec_qtys = []ExecQty{}

type ExecQty struct {
	Exec string
	Qty  int
}

func loadExtraceLogFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)

}
