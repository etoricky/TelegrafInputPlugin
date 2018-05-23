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
	"math"
)

// inputs/all.go
// 	_ "github.com/influxdata/telegraf/plugins/inputs/dde"

type Quote struct {
	Time time.Time // 2017.10.30 15:19:22
	Open float64 // 1272.90
	High float64 // 1273.86
	Low float64 // 1268.94
	YClose float64 // 1270.70
	Bid float64 // 1269.05
	Ask float64 // 1269.55
	HighSpread float64 // 12
	LowSpread float64 // 234
}

type DdeData struct {
	Timezone string
	quotes map[string]Quote
}

func (s *DdeData) SampleConfig() string {
	return `
  ## IANA Time Zone, Asia/Hong_Kong or Europe/London
  timezone = "Europe/London"
`
}

func (s *DdeData) Description() string {
	return "Generate datapoint from Universal DDE Connector TCP socket"
}

func (s *DdeData) Gather(_ telegraf.Accumulator) error {
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

func handlePrice(res string, s *DdeData, location *time.Location) (*string, *Quote, error) {
	fields := strings.Fields(res) // "GOLD 1292.11 1292.61"
	if fields[0]==">" {
		return nil, nil, errors.New("Starting with >")
	}
	if len(fields)!=3 {
		return nil, nil, errors.New("Error: Not starting with > but not 3 fields")
	}

	symbol := fields[0]
	bid, err := strconv.ParseFloat(fields[1], 64)
	if err!=nil {
		return nil, nil, err
	}
	ask, err := strconv.ParseFloat(fields[2], 64)
	if err!=nil {
		return nil, nil, err
	}
	spread := ask - bid
	now := time.Now().In(location)

	last, exist := s.quotes[symbol]
	if !exist {
		last = Quote{now, bid, bid, bid, bid, bid, ask, spread, spread}
		s.quotes[symbol] = last
	}

	curr := Quote{now, last.Open, last.High, last.Low, last.YClose, bid, ask,last.HighSpread, last.LowSpread }

	if now.Day()!=last.Time.Day() {
		curr.Open = bid
		curr.High = bid
		curr.Low = bid
		curr.YClose = last.Bid
		curr.HighSpread = spread
		curr.LowSpread = spread
	}

	curr.HighSpread = math.Max(curr.HighSpread, spread)
	curr.LowSpread = math.Min(curr.LowSpread, spread)
	curr.High = math.Max(curr.High, bid)
	curr.Low = math.Min(curr.Low, bid)

	s.quotes[symbol] = curr
	return &symbol, &curr, nil
}

func (s *DdeData) Start(acc telegraf.Accumulator) error {

	location, err := time.LoadLocation(s.Timezone)
	if err != nil {
		fmt.Println(err)
	}

	fn := func(res string) {
		symbol, quote, err := handlePrice(res, s, location)
		if err!=nil {
			return
		}
		fields := make(map[string]interface{})
		fields["Time"] = quote.Time.Format("2006-01-02 15:04:05")
		fields["Symbol"] = *symbol
		fields["Open"] = quote.Open
		fields["High"] = quote.High
		fields["Low"] = quote.Low
		fields["YClose"] = quote.YClose
		fields["Bid"] = quote.Bid
		fields["Ask"] = quote.Ask
		fields["HighSpread"] = quote.HighSpread
		fields["LowSpread"] = quote.LowSpread
		tags := make(map[string]string)
		acc.AddFields(*symbol, fields, tags, quote.Time)
	}
	go start(fn)

	return nil
}

func (s *DdeData) Stop() {
}

func init() {
	inputs.Add("dde", func() telegraf.Input { return &DdeData{"Europe/London", make(map[string]Quote)} })
}

// json
// {"fields":{"ask":0,"bid":0},"name":"gold","tags":{"host":"centos75"},"timestamp":1526896082}

// ideal
// {"Time":"2017.10.30 15:19:22","Symbol":"GOLD","Open":1272.90,"High":1273.86,"Low":1268.94,"Ask":1269.55,"Bid":1269.05,"YClose":1270.70,"HighSpread":12,"LowSpread":234}

// expect: no tag, random timestamp, empty name
// {"fields":{"Time":"2017.10.30 15:19:22","Symbol":"GOLD","Open":1272.90,"High":1273.86,"Low":1268.94,"Ask":1269.55,"Bid":1269.05,"YClose":1270.70,"HighSpread":12,"LowSpread":234},"name":"","timestamp":1526896082}