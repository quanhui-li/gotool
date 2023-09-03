package distributed_lock

import (
	"context"
	"errors"
	"time"
)

// RetryStrategy 迭代器的模式实现重试策略
type RetryStrategy interface {
	// Next 有两个返回值，第一个标识重试的间隔，第二个标识是否进行重试
	Next(err error) (time.Duration, bool)
}

// FixTimeIntervalStrategy 固定时间间隔策略
type FixTimeIntervalStrategy struct {
	// 时间间隔
	Interval time.Duration
	// 当前的次数
	cnt int
	// 最大的次数
	maxCnt int
}

func (f *FixTimeIntervalStrategy) Next(err error) (time.Duration, bool) {
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return f.Interval, false
	}

	f.cnt++
	return f.Interval, f.cnt >= f.maxCnt
}

// ExponentialBackOffIntervalStrategy 使用指数退避实现动态的等待时间
type ExponentialBackOffIntervalStrategy struct {
	// 最大的重试间隔
	MaxInterval time.Duration
	// 当前的重试间隔
	CntInterval time.Duration
}

func (s *ExponentialBackOffIntervalStrategy) Next(err error) (time.Duration, bool) {
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return s.CntInterval, false
	}
	s.CntInterval = s.CntInterval << 1
	return s.CntInterval, s.CntInterval >= s.MaxInterval
}
