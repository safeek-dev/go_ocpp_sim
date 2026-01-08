package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"ocpp-simulator/internal/configs"
	"ocpp-simulator/internal/logging"
	"ocpp-simulator/internal/metrics"
	"ocpp-simulator/internal/models"

	"github.com/gocarina/gocsv"
)

// Manager manages all charge point instances
type Manager struct {
	mu                  sync.RWMutex
	instances           map[string]*CPInstance
	chargepointConfigs  map[string]*models.ChargePoint
	remoteStartProfiles []*models.RemoteStartProfile
	transactions        map[string]*models.Transaction // keyed by chargeBoxId:connectorId
	logger              *logging.Logger
	metricsTracker      *metrics.Tracker
	nextTransactionId   int
	transactionIdMutex  sync.Mutex
}

// NewManager creates a new manager
func NewManager(logger *logging.Logger, tracker *metrics.Tracker) *Manager {
	return &Manager{
		instances:           make(map[string]*CPInstance),
		chargepointConfigs:  make(map[string]*models.ChargePoint),
		remoteStartProfiles: make([]*models.RemoteStartProfile, 0),
		transactions:        make(map[string]*models.Transaction),
		logger:              logger,
		metricsTracker:      tracker,
		nextTransactionId:   1,
	}
}

// LoadChargepoints loads and validates chargepoints CSV
func (m *Manager) LoadChargepoints(data []byte) error {
	var chargepoints []*models.ChargePoint
	if err := gocsv.UnmarshalBytes(data, &chargepoints); err != nil {
		return fmt.Errorf("CSV parse error: %w", err)
	}

	// Validate: no duplicates, no empty chargeBoxId, connectorCount > 0
	seen := make(map[string]bool)
	for i, cp := range chargepoints {
		if cp.ChargeBoxId == "" {
			return fmt.Errorf("row %d: chargeBoxId cannot be empty", i+2) // +2 for header
		}
		if cp.ConnectorCount <= 0 {
			return fmt.Errorf("row %d: connectorCount must be > 0", i+2)
		}
		if seen[cp.ChargeBoxId] {
			return fmt.Errorf("row %d: duplicate chargeBoxId '%s'", i+2, cp.ChargeBoxId)
		}
		seen[cp.ChargeBoxId] = true

		// Fill defaults
		if cp.Vendor == "" {
			cp.Vendor = "Generic"
		}
		if cp.Model == "" {
			cp.Model = "Simulator"
		}
		if cp.FirmwareVersion == "" {
			cp.FirmwareVersion = "1.0.0"
		}
		if cp.MeterSerialNo == "" {
			cp.MeterSerialNo = fmt.Sprintf("MSN%s", cp.ChargeBoxId)
		}
		if cp.MeterType == "" {
			cp.MeterType = "ACMeter"
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.chargepointConfigs = make(map[string]*models.ChargePoint)
	for _, cp := range chargepoints {
		m.chargepointConfigs[cp.ChargeBoxId] = cp
	}

	m.logger.LogInfo("Manager", fmt.Sprintf("Loaded %d chargepoints", len(chargepoints)))
	return nil
}

// LoadRemoteStartProfiles loads and validates remote start profiles CSV
func (m *Manager) LoadRemoteStartProfiles(data []byte) error {
	var profiles []*models.RemoteStartProfile
	if err := gocsv.UnmarshalBytes(data, &profiles); err != nil {
		return fmt.Errorf("CSV parse error: %w", err)
	}

	// Validate profiles
	for i, p := range profiles {
		if p.ProfileName == "" {
			return fmt.Errorf("row %d: profileName cannot be empty", i+2)
		}
		if p.ChargeBoxId == "" && p.ConnectorId == "" {
			return fmt.Errorf("row %d: at least one of chargeBoxId or connectorId must be specified", i+2)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.remoteStartProfiles = profiles
	m.logger.LogInfo("Manager", fmt.Sprintf("Loaded %d remote start profiles", len(profiles)))
	return nil
}

// GetCSVStatus returns the current CSV status
func (m *Manager) GetCSVStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return map[string]interface{}{
		"chargepoints_loaded": len(m.chargepointConfigs),
		"profiles_loaded":     len(m.remoteStartProfiles),
	}
}

// StartCPs starts N charge points - FIXED VERSION WITH DETAILED RESULTS
func (m *Manager) StartCPs(ctx context.Context, count int, cfg *configs.SimulatorConfig) (map[string]interface{}, error) {
	m.mu.RLock()
	if count > len(m.chargepointConfigs) {
		m.mu.RUnlock()
		return map[string]interface{}{
			"requested": count,
			"available": len(m.chargepointConfigs),
		}, fmt.Errorf("requested %d CPs but only %d available in config", count, len(m.chargepointConfigs))
	}
	m.mu.RUnlock()

	results := map[string]interface{}{
		"requested":  count,
		"started":    0,
		"failed":     0,
		"cp_details": make([]map[string]interface{}, 0),
		"error_logs": make([]string, 0),
	}

	i := 0
	m.mu.RLock()

	// Collect configs to start
	configsToStart := make([]*models.ChargePoint, 0)
	for _, cpConfig := range m.chargepointConfigs {
		if i >= count {
			break
		}
		configsToStart = append(configsToStart, cpConfig)
		i++
	}
	m.mu.RUnlock()

	// Start each CP
	for _, cpConfig := range configsToStart {
		m.mu.Lock()
		if _, exists := m.instances[cpConfig.ChargeBoxId]; exists {
			m.mu.Unlock()
			m.logger.LogInfo(cpConfig.ChargeBoxId, "CP already running, skipping")
			continue
		}

		instance := NewCPInstance(
			cpConfig,
			cfg.OCPPServerURL,
			cfg.HeartbeatInterval,
			cfg.MeterValueInterval,
			cfg.TransactionCutoff,
			cfg.RemoteStartURL,
			cfg.RemoteStopURL,
			cfg.RemoteStartToken,
			cfg.RemoteStopToken,
			m.logger,
			m.metricsTracker,
		)

		m.instances[cpConfig.ChargeBoxId] = instance
		m.mu.Unlock()

		// Start the CP in a goroutine
		// IMPORTANT: Increment metrics BEFORE starting goroutine to ensure accurate count
		m.metricsTracker.IncrementActiveCPs()
		
		go func(inst *CPInstance) {
			// Defer cleanup to ensure metrics are decremented even if CP fails
			defer func() {
				m.metricsTracker.DecrementActiveCPs()
				m.mu.Lock()
				// Remove from instances map if still there (cleanup)
				if _, exists := m.instances[inst.config.ChargeBoxId]; exists {
					delete(m.instances, inst.config.ChargeBoxId)
				}
				m.mu.Unlock()
				m.logger.LogInfo(inst.config.ChargeBoxId, "CP run completed and cleaned up")
			}()
			
			inst.Run(ctx)
		}(instance)

		startedCount := results["started"].(int)
		startedCount++
		results["started"] = startedCount

		// Add CP detail
		details := map[string]interface{}{
			"chargeBoxId": cpConfig.ChargeBoxId,
			"status":      "started",
			"connectors":  cpConfig.ConnectorCount,
			"vendor":      cpConfig.Vendor,
			"model":       cpConfig.Model,
		}
		results["cp_details"] = append(results["cp_details"].([]map[string]interface{}), details)

		m.logger.LogInfo(cpConfig.ChargeBoxId, fmt.Sprintf("CP instance created, connecting to %s", cfg.OCPPServerURL))
	}

	if results["started"].(int) == 0 {
		errorMsg := "No chargepoints were started. Check if configs are loaded and OCPP server is reachable."
		results["error_logs"] = append(results["error_logs"].([]string), errorMsg)
	}

	return results, nil
}

// StopAll stops all charge points
func (m *Manager) StopAll(ctx context.Context) {
	m.mu.Lock()
	instances := make([]*CPInstance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, instance)
	}
	// Clear instances map immediately to prevent new operations
	m.instances = make(map[string]*CPInstance)
	m.mu.Unlock()

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *CPInstance) {
			defer wg.Done()
			inst.Disconnect(ctx)
			// Metrics will be decremented by the goroutine's defer in Run()
			// But we also decrement here as a safety measure
			m.metricsTracker.DecrementActiveCPs()
		}(instance)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

// StopCP stops a specific charge point
func (m *Manager) StopCP(ctx context.Context, chargeBoxId string) {
	m.mu.Lock()
	instance, exists := m.instances[chargeBoxId]
	if !exists {
		m.mu.Unlock()
		return
	}
	delete(m.instances, chargeBoxId)
	m.mu.Unlock()
	
	instance.Disconnect(ctx)
	// Metrics will be decremented by the goroutine's defer in Run()
	// But we also decrement here as a safety measure
	m.metricsTracker.DecrementActiveCPs()
}

// GetAllCPs returns status of all active CPs
func (m *Manager) GetAllCPs() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]map[string]interface{}, 0, len(m.instances))
	for id, instance := range m.instances {
		result = append(result, map[string]interface{}{
			"chargeBoxId": id,
			"status":      instance.GetStatus(),
		})
	}
	return result
}

// GetCP returns status of a specific CP
func (m *Manager) GetCP(chargeBoxId string) map[string]interface{} {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	m.mu.RUnlock()
	if !exists {
		return nil
	}
	return map[string]interface{}{
		"chargeBoxId": chargeBoxId,
		"status":      instance.GetStatus(),
	}
}

// GetAllTransactions returns all active transactions
func (m *Manager) GetAllTransactions() []*models.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*models.Transaction, 0, len(m.transactions))
	for _, txn := range m.transactions {
		result = append(result, txn)
	}
	return result
}

// ============ OCPP Commands ============

func (m *Manager) SendBootNotification(ctx context.Context, chargeBoxId string) error {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("CP not found")
	}
	instance.SendBootNotification(ctx)
	return nil
}

func (m *Manager) SendHeartbeat(ctx context.Context, chargeBoxId string) error {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("CP not found")
	}
	instance.SendHeartbeat(ctx)
	return nil
}

func (m *Manager) SendStatusNotification(ctx context.Context, chargeBoxId string, connectorId int, status string) error {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("CP not found")
	}
	instance.SendStatusNotification(ctx, connectorId, status)
	return nil
}

func (m *Manager) StartTransaction(ctx context.Context, chargeBoxId string, connectorId int, idTag string, meterStart int) (int, error) {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	m.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("CP not found")
	}

	m.transactionIdMutex.Lock()
	txnId := m.nextTransactionId
	m.nextTransactionId++
	m.transactionIdMutex.Unlock()

	instance.StartTransaction(ctx, connectorId, idTag, meterStart, txnId)
	txn := &models.Transaction{
		TransactionId: txnId,
		ChargeBoxId:   chargeBoxId,
		ConnectorId:   connectorId,
		IdTag:         idTag,
		MeterStart:    meterStart,
		StartTime:     time.Now().UTC().Format(time.RFC3339),
		Status:        "active",
	}

	m.mu.Lock()
	key := fmt.Sprintf("%s:%d", chargeBoxId, connectorId)
	m.transactions[key] = txn
	m.mu.Unlock()

	return txnId, nil
}

func (m *Manager) StopTransaction(ctx context.Context, chargeBoxId string, connectorId int) error {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("CP not found")
	}

	instance.StopTransaction(ctx, connectorId)
	return nil
}

// ============ Remote Start/Stop ============

func (m *Manager) RemoteStartTransaction(ctx context.Context, chargeBoxId string, connectorId int) error {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	profiles := m.remoteStartProfiles
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("CP not found")
	}

	// Find matching profile
	var profile *models.RemoteStartProfile
	for _, p := range profiles {
		if (p.ChargeBoxId == chargeBoxId || p.ChargeBoxId == "*") &&
			(p.ConnectorId == fmt.Sprintf("%d", connectorId) || p.ConnectorId == "*") {
			profile = p
			break
		}
	}
	if profile == nil {
		return fmt.Errorf("no matching remote start profile for CP")
	}

	// Call remote start HTTP API
	payload := map[string]interface{}{
		"locationId":              profile.LocationId,
		"chrgPointId":             profile.ChrgPointId,
		"chrgPointConnectorDetId": profile.ChrgPointConnectorDetId,
		"chargingMethodId":        profile.ChargingMethodId,
		"chargingValue":           profile.ChargingValue,
		"chargingUnitId":          profile.ChargingUnitId,
		"isReservationTrans":      profile.IsReservationTrans,
		"reservationId":           profile.ReservationId,
		"selectedWalletType":      profile.SelectedWalletType,
		"vehicleId":               profile.VehicleId,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", instance.remoteStartURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if instance.remoteStartToken != "" {
		req.Header.Set("Authorization", instance.remoteStartToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		m.logger.LogError(chargeBoxId, fmt.Sprintf("remote start HTTP error: %v", err))
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		m.logger.LogError(chargeBoxId, fmt.Sprintf("remote start HTTP error: %d - %s", resp.StatusCode, string(body)))
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Start transaction in CP
	m.transactionIdMutex.Lock()
	txnId := m.nextTransactionId
	m.nextTransactionId++
	m.transactionIdMutex.Unlock()

	instance.StartTransaction(ctx, connectorId, "REMOTE_RFID", 0, txnId)
	txn := &models.Transaction{
		TransactionId: txnId,
		ChargeBoxId:   chargeBoxId,
		ConnectorId:   connectorId,
		IdTag:         "REMOTE_RFID",
		MeterStart:    0,
		StartTime:     time.Now().UTC().Format(time.RFC3339),
		Status:        "active",
	}

	m.mu.Lock()
	key := fmt.Sprintf("%s:%d", chargeBoxId, connectorId)
	m.transactions[key] = txn
	m.mu.Unlock()

	return nil
}

func (m *Manager) RemoteStopTransaction(ctx context.Context, chargeBoxId string, connectorId int) error {
	m.mu.RLock()
	instance, exists := m.instances[chargeBoxId]
	txn, txnExists := m.transactions[fmt.Sprintf("%s:%d", chargeBoxId, connectorId)]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("CP not found")
	}
	if !txnExists {
		return fmt.Errorf("no active transaction")
	}

	// Call remote stop HTTP API
	payload := map[string]interface{}{
		"chrgDetId": txn.ChrgDetId,
	}
	if txn.ChrgDetId == 0 {
		payload["transactionId"] = txn.TransactionId
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", instance.remoteStopURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if instance.remoteStopToken != "" {
		req.Header.Set("Authorization", instance.remoteStopToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		m.logger.LogError(chargeBoxId, fmt.Sprintf("remote stop HTTP error: %v", err))
	} else {
		resp.Body.Close()
	}

	// Send OCPP stop
	instance.StopTransaction(ctx, connectorId)
	return nil
}

func (m *Manager) RemoteStopAllTransactions(ctx context.Context) {
	m.mu.RLock()
	txns := make([]*models.Transaction, 0, len(m.transactions))
	for _, txn := range m.transactions {
		txns = append(txns, txn)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, txn := range txns {
		wg.Add(1)
		go func(t *models.Transaction) {
			defer wg.Done()
			m.RemoteStopTransaction(ctx, t.ChargeBoxId, t.ConnectorId)
		}(txn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}
