package arithmetic

// inputs/all.go
// _ "github.com/influxdata/telegraf/plugins/inputs/arithmetic"

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Number struct {
	x int64
	Step int64
}

var NumberConfig= `
  step = 2
`

func (s *Number) SampleConfig() string {
	return NumberConfig
}

func (s *Number) Description() string {
	return "Generate arithmetic sequence starting at 0 with step i"
}

func (s *Number) Gather(acc telegraf.Accumulator) error {

	fields := make(map[string]interface{})
	fields["x"] = s.x
	fields["step"] = s.Step
	s.x += s.Step

	tags := make(map[string]string)

	acc.AddFields("arithmetic", fields, tags)

	// influx
	// arithmetic,host=centos75 x=62i,step=2i 1526884991000000000
	// arithmetic,host=centos75 step=2i,x=64i 1526884992000000000

	// json
	// {"fields":{"step":2,"x":4},"name":"arithmetic","tags":{"host":"centos75"},"timestamp":1526885111}

	return nil
}

func init() {
	inputs.Add("arithmetic", func() telegraf.Input { return &Number{x: 0} })
}

// make && ./telegraf --config temp.conf