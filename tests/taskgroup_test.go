package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/BytemanD/skyman/utility"
)

func TestTaskGroup(t *testing.T) {
	nums := utility.Range(10)
	tg := utility.TaskGroup{
		Items:        nums,
		MaxWorker:    2,
		ShowProgress: true,
		Func: func(item interface{}) error {
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(3)))
			return nil
		},
	}
	if err := tg.Start(); err != nil {
		t.Errorf("test faild: %v", err)
	}
}
