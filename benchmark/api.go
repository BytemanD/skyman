package benchmark

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type BenchmarkResult struct {
	Start time.Time
	End   time.Time
	Err   error
}

func (b BenchmarkResult) Error() error {
	return b.Err
}

func (b BenchmarkResult) Spend() time.Duration {
	if b.Err != nil {
		return 0
	}
	return b.End.Sub(b.Start)
}

type CaseInterface interface {
	PreStart() error
	Start() error
}

func RunBenchmarkTest(cases []ServerShow) []BenchmarkResult {
	results := []BenchmarkResult{}
	task := syncutils.TaskGroup[ServerShow]{
		Items:        cases,
		Title:        fmt.Sprintf("test %d case(s)", len(cases)),
		ShowProgress: true,
		MaxWorker:    len(cases),
		Func: func(p ServerShow) error {
			result := BenchmarkResult{Start: time.Now()}
			if err := p.PreStart(); err != nil {
				result.Err = err
				return err
			}
			err := p.Start()
			result.End = time.Now()
			result.Err = err
			results = append(results, result)
			return nil
		},
	}
	task.Start()
	return results
}

type ServerShow struct {
	Client *openstack.Openstack
	Server nova.Server
}

func (c *ServerShow) PreStart() error {
	return nil
}
func (c *ServerShow) Start() error {
	_, err := c.Client.NovaV2().GetServer(c.Server.Id)
	return err
}
