package metrics

import (
	"sync/atomic"
)

type Counters struct {
	MessagesSent        int64
	MessagesFailed      int64
	BroadcastsCreated   int64
	BroadcastsCompleted int64
	BroadcastsFailed    int64
	SessionsActive      int64
	RequestsTotal       int64
	RequestsError       int64

	StartupBootstrapRuns           int64
	StartupBootstrapMismatches     int64
	StartupBootstrapStartAttempts  int64
	StartupBootstrapStartErrors    int64
	StartupBootstrapLastDurationMs int64
}

var counters Counters

func IncrementMessagesSent() {
	atomic.AddInt64(&counters.MessagesSent, 1)
}

func IncrementMessagesFailed() {
	atomic.AddInt64(&counters.MessagesFailed, 1)
}

func IncrementBroadcastsCreated() {
	atomic.AddInt64(&counters.BroadcastsCreated, 1)
}

func IncrementBroadcastsCompleted() {
	atomic.AddInt64(&counters.BroadcastsCompleted, 1)
}

func IncrementBroadcastsFailed() {
	atomic.AddInt64(&counters.BroadcastsFailed, 1)
}

func IncrementSessionsActive() {
	atomic.AddInt64(&counters.SessionsActive, 1)
}

func DecrementSessionsActive() {
	atomic.AddInt64(&counters.SessionsActive, -1)
}

func IncrementRequestsTotal() {
	atomic.AddInt64(&counters.RequestsTotal, 1)
}

func IncrementRequestsError() {
	atomic.AddInt64(&counters.RequestsError, 1)
}

func IncrementStartupBootstrapRuns() {
	atomic.AddInt64(&counters.StartupBootstrapRuns, 1)
}

func AddStartupBootstrapMismatches(n int) {
	atomic.AddInt64(&counters.StartupBootstrapMismatches, int64(n))
}

func AddStartupBootstrapStartAttempts(n int) {
	atomic.AddInt64(&counters.StartupBootstrapStartAttempts, int64(n))
}

func AddStartupBootstrapStartErrors(n int) {
	atomic.AddInt64(&counters.StartupBootstrapStartErrors, int64(n))
}

func SetStartupBootstrapLastDurationMs(durationMs int64) {
	atomic.StoreInt64(&counters.StartupBootstrapLastDurationMs, durationMs)
}

func GetCounters() Counters {
	return Counters{
		MessagesSent:        atomic.LoadInt64(&counters.MessagesSent),
		MessagesFailed:      atomic.LoadInt64(&counters.MessagesFailed),
		BroadcastsCreated:   atomic.LoadInt64(&counters.BroadcastsCreated),
		BroadcastsCompleted: atomic.LoadInt64(&counters.BroadcastsCompleted),
		BroadcastsFailed:    atomic.LoadInt64(&counters.BroadcastsFailed),
		SessionsActive:      atomic.LoadInt64(&counters.SessionsActive),
		RequestsTotal:       atomic.LoadInt64(&counters.RequestsTotal),
		RequestsError:       atomic.LoadInt64(&counters.RequestsError),

		StartupBootstrapRuns:           atomic.LoadInt64(&counters.StartupBootstrapRuns),
		StartupBootstrapMismatches:     atomic.LoadInt64(&counters.StartupBootstrapMismatches),
		StartupBootstrapStartAttempts:  atomic.LoadInt64(&counters.StartupBootstrapStartAttempts),
		StartupBootstrapStartErrors:    atomic.LoadInt64(&counters.StartupBootstrapStartErrors),
		StartupBootstrapLastDurationMs: atomic.LoadInt64(&counters.StartupBootstrapLastDurationMs),
	}
}

func ResetCounters() {
	atomic.StoreInt64(&counters.MessagesSent, 0)
	atomic.StoreInt64(&counters.MessagesFailed, 0)
	atomic.StoreInt64(&counters.BroadcastsCreated, 0)
	atomic.StoreInt64(&counters.BroadcastsCompleted, 0)
	atomic.StoreInt64(&counters.BroadcastsFailed, 0)
	atomic.StoreInt64(&counters.SessionsActive, 0)
	atomic.StoreInt64(&counters.RequestsTotal, 0)
	atomic.StoreInt64(&counters.RequestsError, 0)
	atomic.StoreInt64(&counters.StartupBootstrapRuns, 0)
	atomic.StoreInt64(&counters.StartupBootstrapMismatches, 0)
	atomic.StoreInt64(&counters.StartupBootstrapStartAttempts, 0)
	atomic.StoreInt64(&counters.StartupBootstrapStartErrors, 0)
	atomic.StoreInt64(&counters.StartupBootstrapLastDurationMs, 0)
}
