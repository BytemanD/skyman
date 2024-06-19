package utility

import (
	"context"
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
)

type Interval interface {
	Next() time.Duration
}

type DefaultInterval struct {
	Interval time.Duration
}

func (i DefaultInterval) Next() time.Duration {
	return i.Interval
}

type StepInterval struct {
	Min     time.Duration
	Max     time.Duration
	Step    time.Duration
	current time.Duration
}

func (i *StepInterval) Next() time.Duration {
	if i.current >= i.Max {
		return i.Max
	}
	i.current = min(i.current+i.Step, i.Max)
	return i.current
}

type RetryCondition struct {
	Timeout      time.Duration
	IntervalMin  time.Duration
	IntervalMax  time.Duration
	IntervalStep time.Duration
	interval     time.Duration
}

func (c *RetryCondition) NextInterval() time.Duration {
	if c.IntervalMax > 0 && c.interval >= c.IntervalMax {
		return c.IntervalMax
	}
	if c.interval == 0 {
		c.interval = c.IntervalMin
	} else if c.IntervalStep > 0 {
		c.interval += c.IntervalStep
	}
	if c.IntervalMax > 0 {
		c.interval = min(c.interval, c.IntervalMax)
	}
	return c.interval

}
func Retry(condition RetryCondition, function func() bool) error {
	startTime := time.Now()
	for {
		retry := function()
		if !retry {
			return nil
		}
		if condition.Timeout > 0 && time.Since(startTime) >= condition.Timeout {
			return fmt.Errorf("retry timeout(%v)", condition.Timeout)
		}
		time.Sleep(condition.NextInterval())
	}
}

func RetryWithContext(ctx context.Context, condition RetryCondition, function func() error) error {
	startTime := time.Now()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := function()
		if err == nil {
			return nil
		}
		logging.Debug("error: %s", err)
		if condition.Timeout > 0 && time.Since(startTime) >= condition.Timeout {
			return fmt.Errorf("retry timeout(%v), last error: %s", condition.Timeout, err)
		}
		time.Sleep(condition.NextInterval())
	}
}

func RetryWithErrors(condition RetryCondition, matchErrors []string, function func() error) error {
	startTime := time.Now()
	for {
		err := function()
		if err == nil {
			return nil
		}
		for _, e := range matchErrors {
			if compare.IsStructOf(e, err) {
				err = nil
				break
			}
		}
		if err != nil {
			return err
		}
		if condition.Timeout > 0 && time.Since(startTime) >= condition.Timeout {
			return fmt.Errorf("retry timeout(%v), last error: %s", condition.Timeout, err)
		}
		time.Sleep(condition.NextInterval())
	}
}
