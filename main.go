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
	"strings"
	"sync"
	"syscall"
	"time"

	"ocpp-simulator/internal/configs"
	"ocpp-simulator/internal/logging"
	"ocpp-simulator/internal/metrics"
	"ocpp-simulator/internal/simulator"

	"github.com/gorilla/websocket"
)

// Global state
var (
	cpManager      *simulator.Manager
	logger         *logging.Logger
	metricsTracker *metrics.Tracker
	config         *configs.SimulatorConfig
	configMutex    sync.RWMutex
)

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
	
	// Try to load CSV files on startup if they exist
	loadCSVFilesIfExists()
}

func main() {
	mux := http.NewServeMux()

	// API routes FIRST (before catch-all "/")
	// Using method checks for Go 1.21 compatibility (method-specific routing requires Go 1.22+)
	mux.HandleFunc("/api/metrics", methodCheck("GET", handleGetMetrics))
	mux.HandleFunc("/api/csv/status", methodCheck("GET", handleCSVStatus))
	mux.HandleFunc("/api/transactions", methodCheck("GET", handleGetTransactions))
	mux.HandleFunc("/api/cps", handleCPsRouter) // Handles both GET and POST
	mux.HandleFunc("/api/logs", handleLogsRouter) // Handles GET with optional path param
	mux.HandleFunc("/api/config", handleConfigRouter) // Handles both GET and POST
	mux.HandleFunc("/api/csv/chargepoints", methodCheck("POST", handleUploadChargepoints))
	mux.HandleFunc("/api/csv/profiles", methodCheck("POST", handleUploadProfiles))
	mux.HandleFunc("/api/cps/start", methodCheck("POST", handleStartCPs))
	mux.HandleFunc("/api/cps/stop", methodCheck("POST", handleStopAllCPs))
	mux.HandleFunc("/api/remote/stop-all", methodCheck("POST", handleRemoteStopAll))
	mux.HandleFunc("/api/ocpp/", handleOCPPRouter) // Handles all OCPP routes
	mux.HandleFunc("/api/remote/", handleRemoteRouter) // Handles remote start/stop routes

	// NEW DEBUG ROUTES
	mux.HandleFunc("/api/debug/health", methodCheck("GET", handleHealth))
	mux.HandleFunc("/api/debug/ocpp-connection", methodCheck("GET", handleOCPPConnectionTest))

	// Catch-all for root LAST
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

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

// ============ Helper Functions ============

func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// methodCheck wraps a handler to check HTTP method (for Go 1.21 compatibility)
func methodCheck(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

// handleCPsRouter handles /api/cps routes with path parameters
func handleCPsRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api/cps" {
		if r.Method == "GET" {
			handleGetCPs(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	
	// Handle /api/cps/{chargeBoxId} or /api/cps/{chargeBoxId}/stop
	parts := splitPath(path[len("/api/cps/"):])
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	
	chargeBoxId := parts[0]
	if len(parts) == 1 {
		if r.Method == "GET" {
			r = setPathValue(r, "chargeBoxId", chargeBoxId)
			handleGetCP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 2 && parts[1] == "stop" {
		if r.Method == "POST" {
			r = setPathValue(r, "chargeBoxId", chargeBoxId)
			handleStopCP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.NotFound(w, r)
	}
}

// handleLogsRouter handles /api/logs routes
func handleLogsRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := r.URL.Path
	if path == "/api/logs" {
		handleGetGlobalLogs(w, r)
		return
	}
	
	// Handle /api/logs/{chargeBoxId}
	if strings.HasPrefix(path, "/api/logs/") {
		chargeBoxId := path[len("/api/logs/"):]
		r = setPathValue(r, "chargeBoxId", chargeBoxId)
		handleGetCPLogs(w, r)
		return
	}
	
	http.NotFound(w, r)
}

// handleConfigRouter handles /api/config routes
func handleConfigRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		handleGetConfig(w, r)
	} else if r.Method == "POST" {
		handleSetConfig(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleOCPPRouter handles /api/ocpp/{chargeBoxId}/... routes
func handleOCPPRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/ocpp/") {
		http.NotFound(w, r)
		return
	}
	
	parts := splitPath(path[len("/api/ocpp/"):])
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	
	chargeBoxId := parts[0]
	action := parts[1]
	
	r = setPathValue(r, "chargeBoxId", chargeBoxId)
	
	switch action {
	case "boot":
		if r.Method == "POST" {
			handleOCPPBoot(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "heartbeat":
		if r.Method == "POST" {
			handleOCPPHeartbeat(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "status":
		if r.Method == "POST" {
			handleOCPPStatus(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "start-transaction":
		if r.Method == "POST" {
			handleOCPPStartTransaction(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		// Check for stop-transaction: /api/ocpp/{chargeBoxId}/{connectorId}/stop-transaction
		if len(parts) == 3 && parts[2] == "stop-transaction" {
			if r.Method == "POST" {
				r = setPathValue(r, "connectorId", parts[1])
				handleOCPPStopTransaction(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	}
}

// handleRemoteRouter handles /api/remote/... routes
func handleRemoteRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/remote/") {
		http.NotFound(w, r)
		return
	}
	
	parts := splitPath(path[len("/api/remote/"):])
	
	if len(parts) == 1 && parts[0] == "stop-all" {
		if r.Method == "POST" {
			handleRemoteStopAll(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	
	if len(parts) == 3 {
		action := parts[0]
		chargeBoxId := parts[1]
		connectorId := parts[2]
		
		r = setPathValue(r, "chargeBoxId", chargeBoxId)
		r = setPathValue(r, "connectorId", connectorId)
		
		if r.Method == "POST" {
			if action == "start" {
				handleRemoteStart(w, r)
			} else if action == "stop" {
				handleRemoteStop(w, r)
			} else {
				http.NotFound(w, r)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	
	http.NotFound(w, r)
}

// Helper functions for path parsing
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// setPathValue creates a new request with a path value (simulating Go 1.22+ PathValue)
func setPathValue(r *http.Request, key, value string) *http.Request {
	// For Go 1.21, we'll use a context value approach
	ctx := context.WithValue(r.Context(), "pathValue_"+key, value)
	return r.WithContext(ctx)
}

// getPathValue retrieves a path value from request context
func getPathValue(r *http.Request, key string) string {
	if val := r.Context().Value("pathValue_" + key); val != nil {
		return val.(string)
	}
	return ""
}

// loadCSVFilesIfExists attempts to load CSV files from the current directory
func loadCSVFilesIfExists() {
	// Try to load chargepoints.csv
	if data, err := os.ReadFile("chargepoints.csv"); err == nil {
		if err := cpManager.LoadChargepoints(data); err != nil {
			log.Printf("Warning: Failed to load chargepoints.csv: %v", err)
		} else {
			log.Printf("Loaded chargepoints.csv on startup")
		}
	} else {
		log.Printf("Info: chargepoints.csv not found, will need to upload via UI")
	}
	
	// Try to load remote_start_profiles.csv
	if data, err := os.ReadFile("remote_start_profiles.csv"); err == nil {
		if err := cpManager.LoadRemoteStartProfiles(data); err != nil {
			log.Printf("Warning: Failed to load remote_start_profiles.csv: %v", err)
		} else {
			log.Printf("Loaded remote_start_profiles.csv on startup")
		}
	} else {
		log.Printf("Info: remote_start_profiles.csv not found, will need to upload via UI")
	}
}

// ============ Config Handlers ============

func handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig configs.SimulatorConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid config"}, http.StatusBadRequest)
		return
	}

	configMutex.Lock()
	config = &newConfig
	configMutex.Unlock()

	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	jsonResponse(w, config, http.StatusOK)
}

// ============ CSV Handlers ============

func handleUploadChargepoints(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		jsonResponse(w, map[string]string{"error": "Missing file"}, http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		jsonResponse(w, map[string]string{"error": "Read error"}, http.StatusBadRequest)
		return
	}

	err = cpManager.LoadChargepoints(data)
	if err != nil {
		jsonResponse(w, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}, http.StatusBadRequest)
		return
	}

	jsonResponse(w, map[string]string{"status": "chargepoints loaded"}, http.StatusOK)
}

func handleUploadProfiles(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		jsonResponse(w, map[string]string{"error": "Missing file"}, http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		jsonResponse(w, map[string]string{"error": "Read error"}, http.StatusBadRequest)
		return
	}

	err = cpManager.LoadRemoteStartProfiles(data)
	if err != nil {
		jsonResponse(w, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}, http.StatusBadRequest)
		return
	}

	jsonResponse(w, map[string]string{"status": "profiles loaded"}, http.StatusOK)
}

func handleCSVStatus(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, cpManager.GetCSVStatus(), http.StatusOK)
}

// ============ CP Simulation Handlers ============

func handleStartCPs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Count int `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid request"}, http.StatusBadRequest)
		return
	}

	if req.Count <= 0 {
		jsonResponse(w, map[string]string{"error": "count must be > 0"}, http.StatusBadRequest)
		return
	}

	configMutex.RLock()
	cfg := *config
	configMutex.RUnlock()

	// Check if OCPP server URL is configured
	if cfg.OCPPServerURL == "" {
		jsonResponse(w, map[string]interface{}{
			"status":   "error",
			"message":  "OCPP Server URL not configured",
			"ocpp_url": cfg.OCPPServerURL,
		}, http.StatusBadRequest)
		return
	}

	// Start the CPs
	results, err := cpManager.StartCPs(r.Context(), req.Count, &cfg)

	if err != nil {
		jsonResponse(w, map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
			"results": results,
		}, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"status":  "ok",
		"message": fmt.Sprintf("Started %d chargepoints", req.Count),
		"results": results,
	}, http.StatusOK)
}

func handleStopAllCPs(w http.ResponseWriter, r *http.Request) {
	cpManager.StopAll(r.Context())
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleStopCP(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	cpManager.StopCP(r.Context(), chargeBoxId)
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleGetCPs(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, cpManager.GetAllCPs(), http.StatusOK)
}

func handleGetCP(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	cp := cpManager.GetCP(chargeBoxId)
	if cp == nil {
		jsonResponse(w, map[string]string{"error": "CP not found"}, http.StatusNotFound)
		return
	}
	jsonResponse(w, cp, http.StatusOK)
}

// ============ OCPP Command Handlers ============

func handleOCPPBoot(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	if err := cpManager.SendBootNotification(r.Context(), chargeBoxId); err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleOCPPHeartbeat(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	if err := cpManager.SendHeartbeat(r.Context(), chargeBoxId); err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleOCPPStatus(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	var req struct {
		ConnectorId int    `json:"connectorId"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid request"}, http.StatusBadRequest)
		return
	}

	if err := cpManager.SendStatusNotification(r.Context(), chargeBoxId, req.ConnectorId, req.Status); err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleOCPPStartTransaction(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	var req struct {
		ConnectorId int    `json:"connectorId"`
		IdTag       string `json:"idTag"`
		MeterStart  int    `json:"meterStart"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid request"}, http.StatusBadRequest)
		return
	}

	if req.IdTag == "" {
		configMutex.RLock()
		req.IdTag = config.IdTag
		configMutex.RUnlock()
	}

	txnId, err := cpManager.StartTransaction(r.Context(), chargeBoxId, req.ConnectorId, req.IdTag, req.MeterStart)
	if err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"transactionId": txnId}, http.StatusOK)
}

func handleOCPPStopTransaction(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	connectorIdStr := getPathValue(r, "connectorId")
	connectorId, _ := strconv.Atoi(connectorIdStr)

	if err := cpManager.StopTransaction(r.Context(), chargeBoxId, connectorId); err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

// ============ Remote Start/Stop Handlers ============

func handleRemoteStart(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	connectorIdStr := getPathValue(r, "connectorId")
	connectorId, _ := strconv.Atoi(connectorIdStr)

	if err := cpManager.RemoteStartTransaction(r.Context(), chargeBoxId, connectorId); err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleRemoteStop(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	connectorIdStr := getPathValue(r, "connectorId")
	connectorId, _ := strconv.Atoi(connectorIdStr)

	if err := cpManager.RemoteStopTransaction(r.Context(), chargeBoxId, connectorId); err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func handleRemoteStopAll(w http.ResponseWriter, r *http.Request) {
	cpManager.RemoteStopAllTransactions(r.Context())
	jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

// ============ Metrics & Logs Handlers ============

func handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, metricsTracker.GetMetrics(), http.StatusOK)
}

func handleGetCPLogs(w http.ResponseWriter, r *http.Request) {
	chargeBoxId := getPathValue(r, "chargeBoxId")
	logs := logger.GetCPLogs(chargeBoxId)
	jsonResponse(w, logs, http.StatusOK)
}

func handleGetGlobalLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	logs := logger.GetRecentLogs(limit)
	jsonResponse(w, logs, http.StatusOK)
}

func handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, cpManager.GetAllTransactions(), http.StatusOK)
}

// ============ DEBUG HANDLERS ============

func handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, http.StatusOK)
}

func handleOCPPConnectionTest(w http.ResponseWriter, r *http.Request) {
	configMutex.RLock()
	ocppURL := config.OCPPServerURL
	configMutex.RUnlock()

	if ocppURL == "" {
		jsonResponse(w, map[string]interface{}{
			"error":    "OCPP_URL not configured",
			"ocpp_url": "",
		}, http.StatusBadRequest)
		return
	}

	// Try to connect with short timeout
	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	conn, _, err := dialer.DialContext(ctx, ocppURL, nil)

	result := map[string]interface{}{
		"ocpp_url":  ocppURL,
		"reachable": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
		jsonResponse(w, result, http.StatusServiceUnavailable)
	} else {
		conn.Close()
		result["message"] = "OCPP server is reachable"
		jsonResponse(w, result, http.StatusOK)
	}
}
