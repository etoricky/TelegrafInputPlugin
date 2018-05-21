package dde

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
	"net"
	"bufio"
	"fmt"
	"strings"
)

// inputs/all.go
// 	_ "github.com/influxdata/telegraf/plugins/inputs/dde"

type Quotes struct {
	Time string // 2017.10.30 15:19:22
	Symbol string // GOLD
	Open float64 // 1272.90
	High float64 // 1273.86
	Low float64 // 1268.94
	Ask float64 // 1269.55
	Bid float64 // 1269.05
	YClose float64 // 1270.70
	HighSpread float64 // 12
	LowSpread float64 // 234
}

var ConfigString= `
  ## no parameters yet
`

func (s *Quotes) SampleConfig() string {
	return ConfigString
}

func (s *Quotes) Description() string {
	return "Generate DataPoint"
}

func (s *Quotes) Gather(_ telegraf.Accumulator) error {
	return nil
}

func login(conn net.Conn, reader *bufio.Reader) bool {

	for {
		res, err := reader.ReadString(' ')
		if err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Printf(res)
		if res=="Login: " {
			fmt.Fprintf(conn, "dde" + "\n")
		} else if res=="Password: " {
			fmt.Fprintf(conn, "1q2w3e4r" + "\n")
			break
		}
	}

	res, err := reader.ReadString('\n')
	res, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("\n")
	fmt.Printf(res)
	if strings.Contains(res, "Access granted") {
		return true
	} else if strings.Contains(res, "Access denied") {
		return false
	} else {
		return false
	}
	return false

}

func start(fn func(string)) {
	for {

		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000") + " " + "Restarting")

		conn, err := net.Dial("tcp", "127.0.0.1:2222")
		defer conn.Close()

		if err!=nil {
			fmt.Println(err)
			continue
		}
		reader := bufio.NewReader(conn)

		if !login(conn, reader) {
			continue
		}

		for {
			res, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			fn(res)
		}

	}
}

func (s *Quotes) Start(acc telegraf.Accumulator) error {
	fn := func(res string) {

		tokens := strings.Fields(res) // "GOLD 1292.11 1292.61"

		fields := make(map[string]interface{})
		fields["Time"] = time.Now().Format("2006-01-02 15:04:05")
		fields["Symbol"] = tokens[0]
		fields["Bid"] = tokens[1]
		fields["Ask"] = tokens[2]

		tags := make(map[string]string)
		acc.AddFields("dde_connector", fields, tags)
	}
	go start(fn)
	return nil
}

func (s *Quotes) Stop() {
}

func init() {
	inputs.Add("dde", func() telegraf.Input { return &Quotes{} })
}

// json
// {"fields":{"ask":0,"bid":0},"name":"gold","tags":{"host":"centos75"},"timestamp":1526896082}

// ideal
// {"Time":"2017.10.30 15:19:22","Symbol":"GOLD","Open":1272.90,"High":1273.86,"Low":1268.94,"Ask":1269.55,"Bid":1269.05,"YClose":1270.70,"HighSpread":12,"LowSpread":234}

// expect: no tag, random timestamp, empty name
// {"fields":{"Time":"2017.10.30 15:19:22","Symbol":"GOLD","Open":1272.90,"High":1273.86,"Low":1268.94,"Ask":1269.55,"Bid":1269.05,"YClose":1270.70,"HighSpread":12,"LowSpread":234},"name":"","timestamp":1526896082}