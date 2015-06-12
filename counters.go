package breaker

import (
	"errors"
	"time"

	"golang.org/x/net/context"
)

var (
	ErrNumberOfSecondsToStoreOutOfBounds error = errors.New("NumberOfSecondsToStore out of bounds, should be between 1 and 60 seconds")
)

type HealthSummary struct {
	Failures        int64
	Success         int64
	Total           int64
	ErrorPercentage float64

	// time for metrics
	LastFailure time.Time
	LastSuccess time.Time
}

// bucket to store the metrics
type HealthCountsBucket struct {
	failures  int64
	success   int64
	lastWrite time.Time
}

type HealthCounts struct {
	// buckets to store the counters
	values []HealthCountsBucket
	// number of buckets
	buckets int
	// time frame to store
	window time.Duration

	// time for the last event
	lastFailure time.Time
	lastSuccess time.Time

	// channels for the event loop
	successChan    chan struct{}
	failuresChan   chan struct{}
	summaryChan    chan struct{}
	summaryOutChan chan HealthSummary

	// context for cancelation
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHealthCounts(numberOfSecondsToStore int) (*HealthCounts, error) {
	if numberOfSecondsToStore <= 0 || numberOfSecondsToStore > 60 {
		return nil, ErrNumberOfSecondsToStoreOutOfBounds
	}

	hc := &HealthCounts{
		buckets: numberOfSecondsToStore,
		window:  time.Duration(numberOfSecondsToStore) * time.Second,
		values:  make([]HealthCountsBucket, numberOfSecondsToStore, numberOfSecondsToStore),

		successChan:    make(chan struct{}),
		failuresChan:   make(chan struct{}),
		summaryChan:    make(chan struct{}),
		summaryOutChan: make(chan HealthSummary),
	}

	hc.ctx, hc.cancel = context.WithCancel(context.Background())

	go hc.run()
	return hc, nil
}

func (h *HealthCounts) Fail() {
	h.failuresChan <- struct{}{}
}

func (h *HealthCounts) Success() {
	h.successChan <- struct{}{}
}

func (m *HealthCounts) Summary() HealthSummary {
	m.summaryChan <- struct{}{}
	return <-m.summaryOutChan
}

func (hc *HealthCounts) Cancel() {
	hc.cancel()
}

func (hc *HealthCounts) run() {
	for {
		select {
		case <-hc.successChan:
			hc.doSuccess()
		case <-hc.failuresChan:
			hc.doFail()
		case <-hc.summaryChan:
			hc.summaryOutChan <- hc.doSummary()
		case <-hc.ctx.Done():
			return
		}
	}

}

func (c *HealthCountsBucket) reset() {
	c.failures = 0
	c.success = 0
}

// The design of the buckets follows the leaky bucket design from Netflix Hytrix
// The limit in the store is 60 seconds
func (hc *HealthCounts) bucket() *HealthCountsBucket {
	now := time.Now()
	index := now.Second() % hc.buckets
	if !hc.values[index].lastWrite.IsZero() {
		elapsed := now.Sub(hc.values[index].lastWrite)
		if elapsed > hc.window {
			hc.values[index].reset()
		}
	}
	hc.values[index].lastWrite = now
	return &hc.values[index]
}

func (h *HealthCounts) doSuccess() {
	h.bucket().success++
	h.lastSuccess = time.Now()
}

func (h *HealthCounts) doFail() {
	h.bucket().failures++
	h.lastFailure = time.Now()
}

func (b *HealthCounts) doSummary() HealthSummary {
	var sum HealthSummary

	now := time.Now()
	for _, value := range b.values {
		if !value.lastWrite.IsZero() && (now.Sub(value.lastWrite) <= b.window) {
			sum.Success += value.success
			sum.Failures += value.failures
		}
	}
	sum.Total = sum.Success + sum.Failures
	if sum.Total == 0 {
		sum.ErrorPercentage = 0
	} else {
		sum.ErrorPercentage = float64(sum.Failures) / float64(sum.Total) * 100.0
	}

	sum.LastFailure = b.lastFailure
	sum.LastSuccess = b.lastSuccess

	return sum
}
