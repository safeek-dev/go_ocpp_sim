package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"ocpp-simulator/internal/logging"
	"ocpp-simulator/internal/metrics"
	"ocpp-simulator/internal/models"
)

// CPInstance represents a single charge point instance
type CPInstance struct {
	config             *models.ChargePoint
	ocppURL            string
	heartbeatInterval  int
	meterValueInterval int
	transactionCutoff  int
	remoteStartURL     string
	remoteStopURL      string
	remoteStartToken   string
	remoteStopToken    string
	logger             *logging.Logger
	metricsTracker     *metrics.Tracker
	conn               *websocket.Conn
	mu                 sync.RWMutex
	status             string // booting, ready, charging, disconnecting, disconnected
	connectors         map[int]*models.ConnectorState
	activeTransactions map[int]*TransactionState
	messageCount       int
	lastBootTime       time.Time
	lastHeartbeat      time.Time
	shutdownChan       chan struct{}
}

// TransactionState tracks transaction-specific state
type TransactionState struct {
	TransactionId int
	ConnectorId   int
	IdTag         string
	MeterStart    int
	MeterStop     int
	StartTime     time.Time
	Duration      time.Duration
	CurrentEnergy float64 // kWh
	Status        string  // active, stopping, stopped
	CutoffTime    time.Time
}

// NewCPInstance creates a new CP instance
func NewCPInstance(
	config *models.ChargePoint,
	ocppURL string,
	heartbeatInterval, meterValueInterval, transactionCutoff int,
	remoteStartURL, remoteStopURL, remoteStartToken, remoteStopToken string,
	logger *logging.Logger,
	metricsTracker *metrics.Tracker,
) *CPInstance {
	cp := &CPInstance{
		config:             config,
		ocppURL:            ocppURL,
		heartbeatInterval:  heartbeatInterval,
		meterValueInterval: meterValueInterval,
		transactionCutoff:  transactionCutoff,
		remoteStartURL:     remoteStartURL,
		remoteStopURL:      remoteStopURL,
		remoteStartToken:   remoteStartToken,
		remoteStopToken:    remoteStopToken,
		logger:             logger,
		metricsTracker:     metricsTracker,
		status:             "disconnected",
		connectors:         make(map[int]*models.ConnectorState),
		activeTransactions: make(map[int]*TransactionState),
		shutdownChan:       make(chan struct{}),
	}

	// Initialize connectors
	for i := 1; i <= config.ConnectorCount; i++ {
		cp.connectors[i] = &models.ConnectorState{
			ConnectorId: i,
			Status:      "Available",
		}
	}

	return cp
}

// Run starts the CP's main event loop
func (cp *CPInstance) Run(ctx context.Context) {
	defer func() {
		cp.mu.Lock()
		cp.status = "disconnected"
		cp.mu.Unlock()
	}()

	// Connect to OCPP server
	if err := cp.connectAndBoot(ctx); err != nil {
		cp.logger.LogError(cp.config.ChargeBoxId, fmt.Sprintf("Failed to connect: %v", err))
		return
	}

	// Start background tasks
	go cp.heartbeatLoop(ctx)
	go cp.meterValueLoop(ctx)
	go cp.transactionCheckLoop(ctx)

	// Message receive loop
	cp.messageLoop(ctx)
}

// connectAndBoot establishes WebSocket connection and sends BootNotification
func (cp *CPInstance) connectAndBoot(ctx context.Context) error {
	cp.mu.Lock()
	cp.status = "booting"
	cp.mu.Unlock()

	cp.logger.LogInfo(cp.config.ChargeBoxId, fmt.Sprintf("Attempting to connect to OCPP server at %s", cp.ocppURL))

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, cp.ocppURL, nil)
	if err != nil {
		cp.logger.LogError(cp.config.ChargeBoxId, fmt.Sprintf("WebSocket dial failed: %v (URL: %s)", err, cp.ocppURL))
		cp.mu.Lock()
		cp.status = "failed_to_connect"
		cp.mu.Unlock()
		return fmt.Errorf("websocket dial: %w", err)
	}

	cp.logger.LogInfo(cp.config.ChargeBoxId, "WebSocket connection established successfully")

	cp.mu.Lock()
	cp.conn = conn
	cp.lastBootTime = time.Now()
	cp.mu.Unlock()

	// Send BootNotification
	cp.logger.LogInfo(cp.config.ChargeBoxId, "Sending BootNotification")
	cp.SendBootNotification(ctx)

	cp.mu.Lock()
	cp.status = "ready"
	cp.mu.Unlock()

	cp.logger.LogInfo(cp.config.ChargeBoxId, "CP ready for operation")

	return nil
}

// ============ OCPP Message Sending ============

func (cp *CPInstance) SendBootNotification(ctx context.Context) {
	cp.mu.RLock()
	conn := cp.conn
	cp.mu.RUnlock()

	if conn == nil {
		return
	}

	messageId := uuid.New().String()
	payload := []interface{}{
		2, // CALL
		messageId,
		"BootNotification",
		map[string]string{
			"chargePointVendor": cp.config.Vendor,
			"chargePointModel":  cp.config.Model,
			"firmwareVersion":   cp.config.FirmwareVersion,
			"meterSerialNumber": cp.config.MeterSerialNo,
			"meterType":         cp.config.MeterType,
		},
	}

	cp.sendOCPPMessage("BootNotification", payload)
}

func (cp *CPInstance) SendHeartbeat(ctx context.Context) {
	cp.mu.RLock()
	conn := cp.conn
	cp.mu.RUnlock()

	if conn == nil {
		return
	}

	messageId := uuid.New().String()
	payload := []interface{}{
		2, // CALL
		messageId,
		"Heartbeat",
		map[string]interface{}{},
	}

	cp.sendOCPPMessage("Heartbeat", payload)
	cp.mu.Lock()
	cp.lastHeartbeat = time.Now()
	cp.mu.Unlock()
}

func (cp *CPInstance) SendStatusNotification(ctx context.Context, connectorId int, status string) {
	cp.mu.RLock()
	conn := cp.conn
	cp.mu.RUnlock()

	if conn == nil {
		return
	}

	messageId := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)
	payload := []interface{}{
		2, // CALL
		messageId,
		"StatusNotification",
		map[string]interface{}{
			"connectorId": connectorId,
			"errorCode":   "NoError",
			"status":      status,
			"timestamp":   timestamp,
		},
	}

	cp.sendOCPPMessage("StatusNotification", payload)

	cp.mu.Lock()
	if connector, ok := cp.connectors[connectorId]; ok {
		connector.Status = status
		connector.LastStatusTime = timestamp
	}
	cp.mu.Unlock()
}

func (cp *CPInstance) SendMeterValues(ctx context.Context, connectorId int) {
	cp.mu.RLock()
	conn := cp.conn
	txn, txnExists := cp.activeTransactions[connectorId]
	cp.mu.RUnlock()

	if conn == nil || !txnExists {
		return
	}

	messageId := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Simulate realistic energy consumption
	// Ramp up for first 2 minutes, then stable, then ramp down
	elapsedSecs := int(time.Since(txn.StartTime).Seconds())
	rampUpSecs := 120
	var power float64 = 20 // kW

	if elapsedSecs < rampUpSecs {
		// Ramp up
		power = 20.0 * float64(elapsedSecs) / float64(rampUpSecs)
	} else {
		// Stable with small variance
		power = 20.0 + (rand.Float64()-0.5)*2 // ±1 kW variance
	}

	// Update energy (kWh)
	energyDelta := power / 3600.0 // power in kW / 3600 seconds per hour
	txn.CurrentEnergy += energyDelta
	meterStop := int(txn.CurrentEnergy * 1000) // Convert to Wh

	meterValues := []interface{}{
		map[string]interface{}{
			"timestamp": timestamp,
			"sampledValue": []interface{}{
				map[string]interface{}{
					"value":     fmt.Sprintf("%.2f", power),
					"measurand": "Power.Active.Import",
					"unit":      "kW",
					"context":   "Transaction.In-Progress",
				},
				map[string]interface{}{
					"value":     fmt.Sprintf("%.2f", txn.CurrentEnergy),
					"measurand": "Energy.Active.Import.Register",
					"unit":      "kWh",
					"context":   "Transaction.In-Progress",
				},
			},
		},
	}

	payload := []interface{}{
		2, // CALL
		messageId,
		"MeterValues",
		map[string]interface{}{
			"connectorId":   connectorId,
			"meterValue":    meterValues,
			"transactionId": txn.TransactionId,
		},
	}

	cp.sendOCPPMessage("MeterValues", payload)

	cp.mu.Lock()
	if connector, ok := cp.connectors[connectorId]; ok {
		connector.CurrentPower = power
		connector.TotalEnergy = txn.CurrentEnergy
	}
	cp.activeTransactions[connectorId].MeterStop = meterStop
	cp.mu.Unlock()
}

func (cp *CPInstance) StartTransaction(ctx context.Context, connectorId int, idTag string, meterStart int, transactionId int) {
	cp.mu.RLock()
	conn := cp.conn
	cp.mu.RUnlock()

	if conn == nil {
		return
	}

	messageId := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	payload := []interface{}{
		2, // CALL
		messageId,
		"StartTransaction",
		map[string]interface{}{
			"connectorId":   connectorId,
			"idTag":         idTag,
			"meterStart":    meterStart,
			"timestamp":     timestamp,
			"reservationId": 0,
		},
	}

	cp.sendOCPPMessage("StartTransaction", payload)

	// Track transaction locally
	cp.mu.Lock()
	cp.activeTransactions[connectorId] = &TransactionState{
		TransactionId: transactionId,
		ConnectorId:   connectorId,
		IdTag:         idTag,
		MeterStart:    meterStart,
		StartTime:     time.Now(),
		Duration:      time.Duration(cp.transactionCutoff) * time.Minute,
		Status:        "active",
		CutoffTime:    time.Now().Add(time.Duration(cp.transactionCutoff) * time.Minute),
	}
	cp.connectors[connectorId].Status = "Occupied"
	cp.mu.Unlock()

	cp.metricsTracker.IncrementActiveTransactions()
}

func (cp *CPInstance) StopTransaction(ctx context.Context, connectorId int) {
	cp.mu.RLock()
	conn := cp.conn
	txn, exists := cp.activeTransactions[connectorId]
	cp.mu.RUnlock()

	if conn == nil || !exists {
		return
	}

	messageId := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	payload := []interface{}{
		2, // CALL
		messageId,
		"StopTransaction",
		map[string]interface{}{
			"transactionId": txn.TransactionId,
			"idTag":         txn.IdTag,
			"meterStop":     txn.MeterStop,
			"timestamp":     timestamp,
			"reason":        "Local",
		},
	}

	cp.sendOCPPMessage("StopTransaction", payload)

	cp.mu.Lock()
	delete(cp.activeTransactions, connectorId)
	cp.connectors[connectorId].Status = "Available"
	cp.mu.Unlock()

	cp.metricsTracker.DecrementActiveTransactions()
}

// ============ Background Loops ============

func (cp *CPInstance) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(cp.heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cp.shutdownChan:
			return
		case <-ticker.C:
			cp.SendHeartbeat(ctx)
		}
	}
}

func (cp *CPInstance) meterValueLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(cp.meterValueInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cp.shutdownChan:
			return
		case <-ticker.C:
			cp.mu.RLock()
			for connectorId := range cp.activeTransactions {
				cp.mu.RUnlock()
				cp.SendMeterValues(ctx, connectorId)
				cp.mu.RLock()
			}
			cp.mu.RUnlock()
		}
	}
}

func (cp *CPInstance) transactionCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cp.shutdownChan:
			return
		case <-ticker.C:
			cp.mu.RLock()
			now := time.Now()
			for connectorId, txn := range cp.activeTransactions {
				if now.After(txn.CutoffTime) {
					cp.mu.RUnlock()
					cp.StopTransaction(ctx, connectorId)
					cp.mu.RLock()
				}
			}
			cp.mu.RUnlock()
		}
	}
}

// messageLoop receives and processes OCPP messages
func (cp *CPInstance) messageLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-cp.shutdownChan:
			return
		default:
		}

		cp.mu.RLock()
		conn := cp.conn
		cp.mu.RUnlock()

		if conn == nil {
			break
		}

		var msg []interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			cp.logger.LogError(cp.config.ChargeBoxId, fmt.Sprintf("Read error: %v", err))
			break
		}

		if len(msg) >= 2 {
			msgType := int(msg[0].(float64))
			if msgType == 3 { // CALLRESULT
				// Log result
				cp.metricsTracker.IncrementMessagesReceived()
			}
		}
	}
}

// sendOCPPMessage sends an OCPP CALL message
func (cp *CPInstance) sendOCPPMessage(action string, payload []interface{}) {
	cp.mu.Lock()
	conn := cp.conn
	cp.messageCount++
	cp.mu.Unlock()

	if conn == nil {
		return
	}

	if err := conn.WriteJSON(payload); err != nil {
		cp.logger.LogError(cp.config.ChargeBoxId, fmt.Sprintf("Send error: %v", err))
		return
	}

	cp.metricsTracker.IncrementMessagesSent()

	// Log the message
	data, _ := json.Marshal(payload)
	cp.logger.LogOCPPMessage(cp.config.ChargeBoxId, "->", action, "CALL", string(data))
}

// ============ Lifecycle ============

func (cp *CPInstance) Disconnect(ctx context.Context) {
	cp.mu.Lock()
	cp.status = "disconnecting"
	cp.mu.Unlock()

	// Stop all active transactions
	cp.mu.RLock()
	connectors := make([]int, 0, len(cp.activeTransactions))
	for connectorId := range cp.activeTransactions {
		connectors = append(connectors, connectorId)
	}
	cp.mu.RUnlock()

	for _, connectorId := range connectors {
		cp.StopTransaction(ctx, connectorId)
	}

	// Send unavailable status for all connectors
	for i := 1; i <= cp.config.ConnectorCount; i++ {
		cp.SendStatusNotification(ctx, i, "Unavailable")
	}

	// Close WebSocket
	close(cp.shutdownChan)
	cp.mu.Lock()
	if cp.conn != nil {
		cp.conn.Close()
		cp.conn = nil
	}
	cp.mu.Unlock()
}

func (cp *CPInstance) GetStatus() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	connectorStates := make(map[int]interface{})
	for id, state := range cp.connectors {
		connectorStates[id] = state
	}

	return map[string]interface{}{
		"status":                 cp.status,
		"connectors":             connectorStates,
		"activeTransactionCount": len(cp.activeTransactions),
		"messageCount":           cp.messageCount,
		"lastBootTime":           cp.lastBootTime.Format(time.RFC3339),
		"lastHeartbeat":          cp.lastHeartbeat.Format(time.RFC3339),
	}
}

