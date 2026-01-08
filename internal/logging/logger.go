package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger handles JSONL logging to files
type Logger struct {
	logsDir         string
	messageFile     *os.File
	transactionFile *os.File
	mu              sync.Mutex
	cpLogIndices    map[string][]int64 // chargeBoxId -> list of log offsets
	cpLogIndexMutex sync.RWMutex
	recentLogs      []*OCPPMessageLog
	recentLogsMutex sync.RWMutex
}

// OCPPMessageLog represents a logged OCPP message
type OCPPMessageLog struct {
	Timestamp   string `json:"timestamp"`
	ChargeBoxId string `json:"chargeBoxId"`
	ConnectorId int    `json:"connectorId,omitempty"`
	Direction   string `json:"direction"`        // "->", "<-"
	MessageType string `json:"messageType"`      // CALL, CALLRESULT, CALLERROR
	Action      string `json:"action"`           // BootNotification, Heartbeat, etc.
	Status      string `json:"status,omitempty"` // Accepted, Rejected, etc.
	PayloadSize int    `json:"payloadSize"`
	ErrorCode   string `json:"errorCode,omitempty"`
	RawPayload  string `json:"rawPayload,omitempty"`
}

// TransactionLog represents a logged transaction
type TransactionLog struct {
	Timestamp       string  `json:"timestamp"`
	TransactionId   int     `json:"transactionId"`
	ChargeBoxId     string  `json:"chargeBoxId"`
	ConnectorId     int     `json:"connectorId"`
	IdTag           string  `json:"idTag"`
	Event           string  `json:"event"` // start, stop
	MeterStart      int     `json:"meterStart,omitempty"`
	MeterStop       int     `json:"meterStop,omitempty"`
	Energy          float64 `json:"energy,omitempty"` // kWh
	DurationSeconds int     `json:"durationSeconds,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// NewLogger creates a new logger
func NewLogger(logsDir string) (*Logger, error) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}

	msgFile, err := os.OpenFile(
		filepath.Join(logsDir, "ocpp_messages.log"),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	txnFile, err := os.OpenFile(
		filepath.Join(logsDir, "transactions.log"),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		msgFile.Close()
		return nil, err
	}

	return &Logger{
		logsDir:         logsDir,
		messageFile:     msgFile,
		transactionFile: txnFile,
		cpLogIndices:    make(map[string][]int64),
		recentLogs:      make([]*OCPPMessageLog, 0, 1000),
	}, nil
}

// LogOCPPMessage logs an OCPP message
func (l *Logger) LogOCPPMessage(chargeBoxId, direction, action, msgType, payload string) {
	entry := &OCPPMessageLog{
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		ChargeBoxId: chargeBoxId,
		Direction:   direction,
		MessageType: msgType,
		Action:      action,
		PayloadSize: len(payload),
		RawPayload:  payload,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	l.messageFile.Write(append(data, '\n'))

	// Track in memory (limited)
	l.recentLogsMutex.Lock()
	l.recentLogs = append(l.recentLogs, entry)
	if len(l.recentLogs) > 10000 {
		l.recentLogs = l.recentLogs[1:]
	}
	l.recentLogsMutex.Unlock()
}

// LogTransaction logs a transaction event
func (l *Logger) LogTransaction(chargeBoxId string, connectorId int, event string, txnData map[string]interface{}) {
	entry := &TransactionLog{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		ChargeBoxId:   chargeBoxId,
		ConnectorId:   connectorId,
		Event:         event,
		TransactionId: int(txnData["transactionId"].(float64)),
		IdTag:         txnData["idTag"].(string),
	}

	if v, ok := txnData["meterStart"]; ok {
		entry.MeterStart = int(v.(float64))
	}
	if v, ok := txnData["meterStop"]; ok {
		entry.MeterStop = int(v.(float64))
	}
	if v, ok := txnData["reason"]; ok {
		entry.Reason = v.(string)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	l.transactionFile.Write(append(data, '\n'))
}

// LogError logs an error
func (l *Logger) LogError(chargeBoxId, message string) {
	entry := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
		"chargeBoxId": chargeBoxId,
		"level":       "ERROR",
		"message":     message,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	l.messageFile.Write(append(data, '\n'))
}

// LogInfo logs an informational message
func (l *Logger) LogInfo(chargeBoxId, message string) {
	entry := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
		"chargeBoxId": chargeBoxId,
		"level":       "INFO",
		"message":     message,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	l.messageFile.Write(append(data, '\n'))
}

// GetCPLogs returns recent logs for a specific CP
func (l *Logger) GetCPLogs(chargeBoxId string) []*OCPPMessageLog {
	l.recentLogsMutex.RLock()
	defer l.recentLogsMutex.RUnlock()

	result := make([]*OCPPMessageLog, 0)
	for _, log := range l.recentLogs {
		if log.ChargeBoxId == chargeBoxId {
			result = append(result, log)
		}
	}

	// Keep only last 100
	if len(result) > 100 {
		result = result[len(result)-100:]
	}
	return result
}

// GetRecentLogs returns the most recent N global logs
func (l *Logger) GetRecentLogs(limit int) []*OCPPMessageLog {
	l.recentLogsMutex.RLock()
	defer l.recentLogsMutex.RUnlock()

	if len(l.recentLogs) <= limit {
		return l.recentLogs
	}
	return l.recentLogs[len(l.recentLogs)-limit:]
}

// Close closes the log files
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.messageFile != nil {
		l.messageFile.Close()
	}
	if l.transactionFile != nil {
		l.transactionFile.Close()
	}
	return nil
}
