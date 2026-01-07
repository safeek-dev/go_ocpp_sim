package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Tracker tracks performance metrics
type Tracker struct {
	activeCPs              int64
	activeTransactions     int64
	messagesSent           int64
	messagesReceived       int64
	startTime              time.Time
	mu                     sync.RWMutex
}

// NewTracker creates a new metrics tracker
func NewTracker() *Tracker {
	return &Tracker{
		startTime: time.Now(),
	}
}

// IncrementActiveCPs increments the active CP count
func (t *Tracker) IncrementActiveCPs() {
	atomic.AddInt64(&t.activeCPs, 1)
}

// DecrementActiveCPs decrements the active CP count
func (t *Tracker) DecrementActiveCPs() {
	atomic.AddInt64(&t.activeCPs, -1)
}

// IncrementActiveTransactions increments the active transaction count
func (t *Tracker) IncrementActiveTransactions() {
	atomic.AddInt64(&t.activeTransactions, 1)
}

// DecrementActiveTransactions decrements the active transaction count
func (t *Tracker) DecrementActiveTransactions() {
	atomic.AddInt64(&t.activeTransactions, -1)
}

// IncrementMessagesSent increments the messages sent count
func (t *Tracker) IncrementMessagesSent() {
	atomic.AddInt64(&t.messagesSent, 1)
}

// IncrementMessagesReceived increments the messages received count
func (t *Tracker) IncrementMessagesReceived() {
	atomic.AddInt64(&t.messagesReceived, 1)
}

// GetMetrics returns current metrics
func (t *Tracker) GetMetrics() map[string]interface{} {
	activeCPs := atomic.LoadInt64(&t.activeCPs)
	activeTransactions := atomic.LoadInt64(&t.activeTransactions)
	messagesSent := atomic.LoadInt64(&t.messagesSent)
	messagesReceived := atomic.LoadInt64(&t.messagesReceived)
	uptime := time.Since(t.startTime).Seconds()
	msgsPerSec := 0.0
	if uptime > 0 {
		msgsPerSec = float64(messagesSent+messagesReceived) / uptime
	}

	return map[string]interface{}{
		"activeCPs":            activeCPs,
		"activeTransactions":   activeTransactions,
		"messagesSent":         messagesSent,
		"messagesReceived":     messagesReceived,
		"totalMessages":        messagesSent + messagesReceived,
		"uptime":               uptime,
		"messagesPerSecond":    msgsPerSec,
	}
}
