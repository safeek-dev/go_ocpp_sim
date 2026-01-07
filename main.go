package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"ocpp-simulator/internal/configs"
	"ocpp-simulator/internal/logging"
	"ocpp-simulator/internal/metrics"
	"ocpp-simulator/internal/simulator"
)

// Global state
var (
	cpManager      *simulator.Manager
	logger         *logging.Logger
	metricsTracker *metrics.Tracker
	config         *configs.SimulatorConfig
	configMutex    sync.RWMutex
)

// Config represents the simulator configuration
// type Config struct {
// 	OCPPServerURL      string `json:"ocppServerUrl"`
// 	RemoteStartURL     string `json:"remoteStartUrl"`
// 	RemoteStopURL      string `json:"remoteStopUrl"`
// 	RemoteStartToken   string `json:"remoteStartToken"`
// 	RemoteStopToken    string `json:"remoteStopToken"`
// 	HeartbeatInterval  int    `json:"heartbeatInterval"`  // seconds
// 	MeterValueInterval int    `json:"meterValueInterval"` // seconds
// 	TransactionCutoff  int    `json:"transactionCutoffMinutes"`
// 	IdTag              string `json:"idTag"`
// 	ChargePointVendor  string `json:"chargePointVendor"`
// 	ChargePointModel   string `json:"chargePointModel"`
// 	FirmwareVersion    string `json:"firmwareVersion"`
// }

// DefaultConfig returns default configuration
func DefaultConfig() *configs.SimulatorConfig {
	return &configs.SimulatorConfig{
		OCPPServerURL:      "ws://localhost:8001",
		RemoteStartURL:     "http://localhost:9097/tm/secure/api/v1/MBremoteStartTransaction",
		RemoteStopURL:      "http://localhost:9097/tm/secure/api/v1/MBremoteStopTransaction",
		HeartbeatInterval:  60,
		MeterValueInterval: 60,
		TransactionCutoff:  30,
		IdTag:              "DEFAULT_RFID",
		ChargePointVendor:  "Generic",
		ChargePointModel:   "Simulator",
		FirmwareVersion:    "1.0.0",
	}
}

func init() {
	config = DefaultConfig()

	var err error
	logger, err = logging.NewLogger("logs")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	metricsTracker = metrics.NewTracker()
	cpManager = simulator.NewManager(logger, metricsTracker)
}

func main() {
	// Setup HTTP routes
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})

	// API routes
	mux.HandleFunc("POST /api/config", handleSetConfig)
	mux.HandleFunc("GET /api/config", handleGetConfig)
	mux.HandleFunc("POST /api/csv/chargepoints", handleUploadChargepoints)
	mux.HandleFunc("POST /api/csv/profiles", handleUploadProfiles)
	mux.HandleFunc("GET /api/csv/status", handleCSVStatus)
	mux.HandleFunc("POST /api/cps/start", handleStartCPs)
	mux.HandleFunc("POST /api/cps/stop", handleStopAllCPs)
	mux.HandleFunc("POST /api/cps/{chargeBoxId}/stop", handleStopCP)
	mux.HandleFunc("GET /api/cps", handleGetCPs)
	mux.HandleFunc("GET /api/cps/{chargeBoxId}", handleGetCP)
	mux.HandleFunc("POST /api/ocpp/{chargeBoxId}/boot", handleOCPPBoot)
	mux.HandleFunc("POST /api/ocpp/{chargeBoxId}/heartbeat", handleOCPPHeartbeat)
	mux.HandleFunc("POST /api/ocpp/{chargeBoxId}/status", handleOCPPStatus)
	mux.HandleFunc("POST /api/ocpp/{chargeBoxId}/start-transaction", handleOCPPStartTransaction)
	mux.HandleFunc("POST /api/ocpp/{chargeBoxId}/{connectorId}/stop-transaction", handleOCPPStopTransaction)
	mux.HandleFunc("POST /api/remote/start/{chargeBoxId}/{connectorId}", handleRemoteStart)
	mux.HandleFunc("POST /api/remote/stop/{chargeBoxId}/{connectorId}", handleRemoteStop)
	mux.HandleFunc("POST /api/remote/stop-all", handleRemoteStopAll)
	mux.HandleFunc("GET /api/metrics", handleGetMetrics)
	mux.HandleFunc("GET /api/logs/{chargeBoxId}", handleGetCPLogs)
	mux.HandleFunc("GET /api/logs", handleGetGlobalLogs)
	mux.HandleFunc("GET /api/transactions", handleGetTransactions)

	// Serve index.html for root
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})

	// Start HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down simulator...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cpManager.StopAll(ctx)
		server.Shutdown(ctx)
		os.Exit(0)
	}()

	log.Printf("Starting OCPP Simulator on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// ============ Config Handlers ============

func handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig configs.SimulatorConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Invalid config: %v", err), http.StatusBadRequest)
		return
	}

	configMutex.Lock()
	config = &newConfig
	configMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// ============ CSV Handlers ============

func handleUploadChargepoints(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read error: %v", err), http.StatusBadRequest)
		return
	}

	// Validate and load
	err = cpManager.LoadChargepoints(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "chargepoints loaded"})
}

func handleUploadProfiles(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read error: %v", err), http.StatusBadRequest)
		return
	}

	// Validate and load
	err = cpManager.LoadRemoteStartProfiles(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "profiles loaded"})
}

func handleCSVStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cpManager.GetCSVStatus())
}

// ============ CP Simulation Handlers ============

func handleStartCPs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	configMutex.RLock()
	cfg := *config
	configMutex.RUnlock()

	if err := cpManager.StartCPs(r.Context(), req.Count, &cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start CPs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStopAllCPs(w http.ResponseWriter, r *http.Request) {
	cpManager.StopAll(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStopCP(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	cpManager.StopCP(r.Context(), chargeBoxId)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetCPs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cpManager.GetAllCPs())
}

func handleGetCP(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	cp := cpManager.GetCP(chargeBoxId)
	if cp == nil {
		http.Error(w, "CP not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cp)
}

// ============ OCPP Command Handlers ============

func handleOCPPBoot(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	if err := cpManager.SendBootNotification(r.Context(), chargeBoxId); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleOCPPHeartbeat(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	if err := cpManager.SendHeartbeat(r.Context(), chargeBoxId); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleOCPPStatus(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	var req struct {
		ConnectorId int    `json:"connectorId"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := cpManager.SendStatusNotification(r.Context(), chargeBoxId, req.ConnectorId, req.Status); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleOCPPStartTransaction(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	var req struct {
		ConnectorId int    `json:"connectorId"`
		IdTag       string `json:"idTag"`
		MeterStart  int    `json:"meterStart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.IdTag == "" {
		configMutex.RLock()
		req.IdTag = config.IdTag
		configMutex.RUnlock()
	}

	txnId, err := cpManager.StartTransaction(r.Context(), chargeBoxId, req.ConnectorId, req.IdTag, req.MeterStart)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"transactionId": txnId})
}

func handleOCPPStopTransaction(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	connectorIdStr := r.PathValue("connectorId")
	connectorId, _ := strconv.Atoi(connectorIdStr)

	if err := cpManager.StopTransaction(r.Context(), chargeBoxId, connectorId); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ============ Remote Start/Stop Handlers ============

func handleRemoteStart(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	connectorIdStr := r.PathValue("connectorId")
	connectorId, _ := strconv.Atoi(connectorIdStr)

	if err := cpManager.RemoteStartTransaction(r.Context(), chargeBoxId, connectorId); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRemoteStop(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	connectorIdStr := r.PathValue("connectorId")
	connectorId, _ := strconv.Atoi(connectorIdStr)

	if err := cpManager.RemoteStopTransaction(r.Context(), chargeBoxId, connectorId); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRemoteStopAll(w http.ResponseWriter, r *http.Request) {
	cpManager.RemoteStopAllTransactions(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ============ Metrics & Logs Handlers ============

func handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricsTracker.GetMetrics())
}

func handleGetCPLogs(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := r.PathValue("chargeBoxId")
	logs := logger.GetCPLogs(chargeBoxId)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func handleGetGlobalLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	logs := logger.GetRecentLogs(limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cpManager.GetAllTransactions())
}
