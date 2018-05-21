package geometric

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"fmt"
)

// inputs/all.go
// 	_ "github.com/influxdata/telegraf/plugins/inputs/geometric"


type SockerReader struct {
	x int64
	Step int64
	telegraf.Accumulator
}

var NumberConfig= `
  step = 2
`

func (s *SockerReader) SampleConfig() string {
	return NumberConfig
}

func (s *SockerReader) Description() string {
	return "Generate geometric sequence starting at 0 with step i"
}

func (s *SockerReader) Gather(_ telegraf.Accumulator) error {
	//fmt.Print("Gathering")
	return nil
}

func (s *SockerReader) Send() error {
	for {
		fields := make(map[string]interface{})
		fields["x"] = s.x
		fields["step"] = s.Step
		s.x *= s.Step

		tags := make(map[string]string)

		s.AddFields("geometric", fields, tags)
	}
}

func (s *SockerReader) Start(acc telegraf.Accumulator) error {
	fmt.Print("Start")
	s.Accumulator = acc
	go s.Send()
	return nil

	// influx
	// geometric,host=centos75 x=62i,step=2i 1526884991000000000
	// geometric,host=centos75 step=2i,x=64i 1526884992000000000

	// json
	// {"fields":{"step":2,"x":4},"name":"geometric","tags":{"host":"centos75"},"timestamp":1526885111}
}

func (s *SockerReader) Stop(acc telegraf.Accumulator) {
}

func init() {
	inputs.Add("geometric", func() telegraf.Input { return &SockerReader{x: 1} })
}

// make && ./telegraf --config temp.conf