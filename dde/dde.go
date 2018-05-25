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
	"log"
	"os"
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
	Name string
	Password string
	quotes map[string]Quote
}

func (s *DdeData) SampleConfig() string {
	return `## Universal DDE Connector login information
name = "dde"
password = "1q2w3e4r"
`
}

func (s *DdeData) Description() string {
	return "Generate datapoint from Universal DDE Connector TCP socket"
}

func (s *DdeData) Gather(_ telegraf.Accumulator) error {
	return nil
}
var logger *log.Logger

func login(conn net.Conn, reader *bufio.Reader, name string, password string) error {

	for {
		res, err := reader.ReadString(' ')
		if err != nil {
			return err
		}
		logger.Printf(res)
		if res=="Login: " {
			fmt.Fprintf(conn, name + "\n")
		} else if res=="Password: " {
			fmt.Fprintf(conn, password + "\n")
			break
		}
	}

	res, err := reader.ReadString('\n')
	res, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	logger.Printf("\n")
	logger.Printf(res)
	if strings.Contains(res, "Access granted") {
		return nil
	} else if strings.Contains(res, "Access denied") {
		return errors.New("Access denied")
	} else {
		return errors.New("Neither access granted nor denied")
	}
}

func connect(fn func(string), name string, password string) error {

	defer func() {
		if r := recover(); r != nil {
			logger.Println("Recovered in f", r)
		}
	}()

	conn, err := net.Dial("tcp", "127.0.0.1:2222")
	if err!=nil {
		return err
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	err = login(conn, reader, name, password)
	if err!=nil {
		return err
	}

	go func() {
		for _ = range time.NewTicker(60 * time.Second).C {
			fmt.Fprintf(conn, "> Ping" + "\n")
		}
	}()

	for {
		res, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fn(res)
	}

	return nil
}

func start(fn func(string), name string, password string) {

	f, err := os.OpenFile("dde.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	logger = log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	for {
		logger.Println("Reconnecting")
		err = connect(fn, name, password)
		if err!=nil {
			logger.Println(err)
		}
	}
}


func handlePrice(res string, s *DdeData) (*string, *Quote, error) {
	now := time.Now()

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

	fn := func(res string) {
		symbol, quote, err := handlePrice(res, s)
		if err!=nil {
			return
		}
		fields := make(map[string]interface{})
		fields["Time"] = quote.Time.Format("2006-01-02 15:04:05")
		fields["TimeMicro"] = quote.Time.UnixNano() / int64(time.Microsecond)
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
	go start(fn, s.Name, s.Password)

	return nil
}

func (s *DdeData) Stop() {
}

func init() {
	inputs.Add("dde", func() telegraf.Input { return &DdeData{"", "", make(map[string]Quote)} })
}
