package utility

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

type TaskGroup struct {
	Items        interface{}
	Func         func(item interface{}) error
	ShowProgress bool
	MaxWorker    int
	wg           *sync.WaitGroup
}

func (g TaskGroup) Start() error {
	value := reflect.ValueOf(g.Items)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return fmt.Errorf("items must be slice or array")
	}
	g.wg = &sync.WaitGroup{}
	g.wg.Add(value.Len())
	if g.MaxWorker <= 0 {
		g.MaxWorker = runtime.NumCPU()
	}
	workers := make(chan struct{}, g.MaxWorker)
	var bar *pb.ProgressBar
	if g.ShowProgress {
		bar = pb.StartNew(value.Len())
	} else {
		bar = nil
	}
	for i := 0; i < value.Len(); i++ {
		go func(o interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			workers <- struct{}{}
			g.Func(o)
			if bar != nil {
				bar.Increment()
			}
			<-workers
		}(value.Index(i).Interface(), g.wg)
	}
	g.wg.Wait()
	if bar != nil {
		bar.Finish()
	}
	return nil
}
