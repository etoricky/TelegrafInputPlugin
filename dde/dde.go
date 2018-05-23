package dde

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
	"net"
	"bufio"
	"fmt"
	"strings"
	"strconv"
	"github.com/pkg/errors"
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
    Exchange string
    Format string
}

func (s *Quotes) SampleConfig() string {
	return `
  ## Output data format. influx or field_only
  format = "field_only"
`
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

		go func() {
			for _ = range time.NewTicker(1 * time.Second).C {
				fmt.Fprintf(conn, "> Ping" + "\n")
			}
		}()

		for {
			res, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			fn(res)
		}

	}
}

func parseResult(res string, s *Quotes) error {
	fields := strings.Fields(res) // "GOLD 1292.11 1292.61"
	if fields[0]==">" {
		return errors.New("Starting with >")
	}
	if len(fields)!=3 {
		return errors.New("Error: Not starting with > but not 3 fields")
	}

	bid, err := strconv.ParseFloat(fields[1], 64)
	if err!=nil {
		return err
	}
	ask, err := strconv.ParseFloat(fields[2], 64)
	if err!=nil {
		return err
	}

	s.Time = time.Now().Format("2006-01-02 15:04:05")
	s.Symbol = fields[0]
	s.Open = 0.0
	s.High = 0.0
	s.Low = 0.0
	s.Bid = bid
	s.Ask = ask
	s.YClose = 0.0
	s.HighSpread = 0
	s.LowSpread = 0
	s.Exchange = "Hong Kong"
	return nil
}

func (s *Quotes) Start(acc telegraf.Accumulator) error {

	switch s.Format {
	case "influx":
		fn := func(res string) {
			err := parseResult(res, s)
			if err!=nil {
				return
			}
			fields := make(map[string]interface{})
			fields["Open"] = s.Open
			fields["High"] = s.High
			fields["Low"] = s.Low
			fields["Ask"] = s.Ask
			fields["Bid"] = s.Bid
			fields["YClose"] = s.YClose
			fields["HighSpread"] = s.HighSpread
			fields["LowSpread"] = s.LowSpread
			tags := make(map[string]string)
			tags["Exchange"] = s.Exchange
			acc.AddFields(s.Symbol, fields, tags)
		}
		go start(fn)
	case "field_only":
		fn := func(res string) {
			err := parseResult(res, s)
			if err!=nil {
				return
			}
			fields := make(map[string]interface{})
			fields["Time"] = s.Time
			fields["Symbol"] = s.Symbol
			fields["Open"] = s.Open
			fields["High"] = s.High
			fields["Low"] = s.Low
			fields["Ask"] = s.Ask
			fields["Bid"] = s.Bid
			fields["YClose"] = s.YClose
			fields["HighSpread"] = s.HighSpread
			fields["LowSpread"] = s.LowSpread
			fields["Exchange"] = s.Exchange
			tags := make(map[string]string)
			acc.AddFields("dde_connector", fields, tags)
		}
		go start(fn)
	}

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