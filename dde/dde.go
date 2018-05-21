package dde

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
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
	telegraf.Accumulator
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

func (s *Quotes) Send() error {
	for {
		time.Sleep(1 * time.Second)

		fields := make(map[string]interface{})
		fields["Time"] = s.Time
		fields["Symbol"] = s.Symbol
		fields["Bid"] = s.Bid

		tags := make(map[string]string)
		s.AddFields("dde", fields, tags)
	}
}

func (s *Quotes) Start(acc telegraf.Accumulator) error {
	s.Accumulator = acc
	go s.Send()
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