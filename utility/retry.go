package utility

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/go-console/console"
	"github.com/samber/lo"
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

var ErrRetryTimeout = errors.New("retry timeout")

func RetryError(condition RetryCondition, function func() (bool, error)) error {
	startTime := time.Now()
	for {
		retry, err := function()
		if !retry {
			return err
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
		console.Debug("error: %s", err)
		if condition.Timeout > 0 && time.Since(startTime) >= condition.Timeout {
			return fmt.Errorf("retry timeout(%v), last error: %s", condition.Timeout, err)
		}
		time.Sleep(condition.NextInterval())
	}
}

func RetryWithErrors(condition RetryCondition, matchErrors []string, function func() error) error {
	startTime := time.Now()
	var err error
	for {
		err = function()
		if err == nil {
			return nil
		}
		if condition.Timeout > 0 && time.Since(startTime) >= condition.Timeout {
			return fmt.Errorf("retry timeout(%v), last error: %v", condition.Timeout, err)
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
		time.Sleep(condition.NextInterval())
	}
}

// 如果匹配 matchError, 重试, 否则退出
func RetryWithError(condition RetryCondition, retryError error, function func() error, enableWarnings ...bool) error {
	enableWarning := lo.FirstOrEmpty(enableWarnings)
	startTime := time.Now()
	var err error
	for {
		err = function()
		if err == nil {
			return nil
		}
		if condition.Timeout > 0 && time.Since(startTime) >= condition.Timeout {
			return fmt.Errorf("%w(%s): last error: %v", ErrRetryTimeout, condition.Timeout, err)
		}
		if !errors.Is(err, retryError) {
			return fmt.Errorf("unexcept error: %w", err)
		}
		if enableWarning {
			console.Warn("catch error: (%s), retrying", err)
		}
		time.Sleep(condition.NextInterval())
	}
}
