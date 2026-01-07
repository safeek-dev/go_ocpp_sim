# OCPP Load Simulator

A high-performance load testing simulator for OCPP 1.6 charging stations, capable of simulating 5k-10k charge points with a web-based interactive dashboard.

## Features

- **Multiple CP Simulation**: Simulate up to 10,000 charge points concurrently using Go goroutines
- **OCPP 1.6 JSON Protocol**: Full support for BootNotification, Heartbeat, MeterValues, StartTransaction, StopTransaction
- **Interactive Dashboard**: Web UI for real-time control and monitoring
- **CSV-based Configuration**: Load CP definitions and remote start profiles from CSV
- **Remote Start/Stop**: Trigger and manage transactions via HTTP APIs
- **Realistic Load**: Configurable heartbeat intervals, meter values with realistic consumption curves
- **Comprehensive Logging**: File-based JSONL logs with per-CP message tracking
- **Performance Metrics**: Real-time metrics for active CPs, transactions, and message rates
- **Graceful Lifecycle**: Clean connection management with proper OCPP disconnect sequences

## Architecture

```
┌─────────────────────────────────────────┐
│   Web Frontend (Interactive UI)         │
│   - CP Management (Create/Delete)       │
│   - Real-time Metrics & Status          │
│   - CSV Upload                          │
│   - Per-CP OCPP Controls                │
│   - Remote Start/Stop Triggers          │
└──────────────┬──────────────────────────┘
               │ REST / WebSocket
┌──────────────▼──────────────────────────┐
│   Go Backend HTTP Server                │
│   - CP Simulator Manager                │
│   - Goroutine Pool Manager              │
│   - Transaction Service Integration     │
│   - Metrics Aggregation                 │
│   - File-based Logging                  │
└──────────────┬──────────────────────────┘
               │ WebSocket
┌──────────────▼──────────────────────────┐
│   Simulated Charge Points (CPs)         │
│   - OCPP 1.6 Protocol Handler           │
│   - State Machine (Boot→Ready→...)      │
│   - Heartbeat/MeterValue Generation     │
│   - Transaction Management              │
└─────────────────────────────────────────┘
```

## Project Structure

```
ocpp-simulator/
├── main.go                 # HTTP server and entry point
├── go.mod                  # Dependencies
├── README.md               # This file
├── chargepoints.csv        # CP definitions
├── remote_start_profiles.csv # Remote start profiles
├── internal/
│   ├── models/
│   │   ├── chargepoint.go  # CP data structures
│   │   ├── transaction.go  # Transaction tracking
│   │   └── ocpp.go         # OCPP message types
│   ├── simulator/
│   │   ├── manager.go      # CP lifecycle manager
│   │   └── cp_instance.go  # Individual CP logic
│   ├── logging/
│   │   └── logger.go       # File-based logging
│   ├── http/
│   │   └── handlers.go     # HTTP API endpoints
│   └── metrics/
│       └── metrics.go      # Performance metrics
└── web/
    └── index.html          # UI dashboard
```

## Setup and Running

### Prerequisites

- Go 1.21+
- CSV files with CP definitions and remote start profiles

### Installation

```bash
git clone <repo>
cd ocpp-simulator
go mod download
go run main.go
```

### Configuration

1. **Place CSV files** in the project root:
   - `chargepoints.csv`
   - `remote_start_profiles.csv`

2. **Open browser** to `http://localhost:8080`

3. **Configure URLs** in the UI:
   - OCPP Server WebSocket URL (e.g., `ws://your-server:8001`)
   - Remote Start API endpoint
   - Remote Stop API endpoint
   - Auth tokens

4. **Upload CSV** (or use bundled default)

5. **Start simulation**:
   - Set number of CPs
   - Configure heartbeat interval
   - Configure meter interval
   - Set transaction cut-off time
   - Click "Start CPs"

## CSV Format

### chargepoints.csv

```csv
chargeBoxId,connectorCount,chargePointVendor,chargePointModel,firmwareVersion,meterSerialNumber,meterType
MHCGT000001_1,2,Delta,ACMini,1.0.3,MSN12345,ACMeter
RB-DN-5089730,3,ABB,TerraAC,2.1.0,MSN56789,DCMeter
```

### remote_start_profiles.csv

```csv
profileName,chargeBoxId,connectorId,locationId,chrgPointId,chrgPointConnectorDetId,chargingMethodId,chargingValue,chargingUnitId,isReservationTrans,reservationId,selectedWalletType,vehicleId
Location5_Default,MHCGT000001_1,1,5,15,25,214,20,233,N,0,262,78
```

## API Endpoints

### Configuration
- `POST /api/config` - Set OCPP server URL, auth tokens, intervals
- `GET /api/config` - Retrieve current configuration

### CSV Operations
- `POST /api/csv/chargepoints` - Upload chargepoints.csv
- `POST /api/csv/profiles` - Upload remote_start_profiles.csv
- `GET /api/csv/status` - Get CSV loading status

### Simulation Control
- `POST /api/cps/start` - Start N CPs (with config)
- `POST /api/cps/stop` - Stop all CPs
- `POST /api/cps/{chargeBoxId}/stop` - Stop specific CP
- `GET /api/cps` - Get all CPs and status
- `GET /api/cps/{chargeBoxId}` - Get specific CP status

### OCPP Commands
- `POST /api/ocpp/{chargeBoxId}/boot` - Send BootNotification
- `POST /api/ocpp/{chargeBoxId}/heartbeat` - Send Heartbeat
- `POST /api/ocpp/{chargeBoxId}/status` - Send StatusNotification
- `POST /api/ocpp/{chargeBoxId}/start-transaction` - Start local transaction
- `POST /api/ocpp/{chargeBoxId}/{connectorId}/stop-transaction` - Stop transaction

### Remote Start/Stop
- `POST /api/remote/start/{chargeBoxId}/{connectorId}` - Trigger remote start
- `POST /api/remote/stop/{chargeBoxId}/{connectorId}` - Trigger remote stop
- `POST /api/remote/stop-all` - Stop all transactions

### Metrics & Logs
- `GET /api/metrics` - Get aggregated metrics
- `GET /api/logs/{chargeBoxId}` - Get logs for specific CP
- `GET /api/logs` - Get recent global logs
- `GET /api/transactions` - Get all transactions

## Testing with curl

```bash
# Set configuration
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "ocppServerUrl": "ws://ocpp-server:8001",
    "remoteStartUrl": "http://backend:9097/tm/secure/api/v1/MBremoteStartTransaction",
    "remoteStopUrl": "http://backend:9097/tm/secure/api/v1/MBremoteStopTransaction",
    "remoteStartToken": "Bearer eyJ...",
    "remoteStopToken": "Bearer eyJ...",
    "heartbeatInterval": 60,
    "meterValueInterval": 60,
    "transactionCutoffMinutes": 30
  }'

# Start 100 CPs
curl -X POST http://localhost:8080/api/cps/start \
  -H "Content-Type: application/json" \
  -d '{"count": 100}'

# Get metrics
curl http://localhost:8080/api/metrics

# Trigger remote start
curl -X POST http://localhost:8080/api/remote/start/MHCGT000001_1/1

# Stop all transactions
curl -X POST http://localhost:8080/api/remote/stop-all
```

## Performance Expectations

### Memory (10k CPs)
- **200–500 MB**: WebSocket connections and per-CP state
- **50–100 MB**: Indexes, maps, metrics
- **Total**: ~300–600 MB (safe to budget 1 GB)

### Disk (1 hour at full logging)
- **1–2 GB**: Full OCPP message logs (JSONL)
- **Configurable**: Log rotation and sampling to reduce

### Network
- **Heartbeat only** (60s interval): ~16.7 msgs/sec/10k CPs
- **With MeterValues** (60s interval): ~33 msgs/sec/10k CPs
- **Realistic**: 50–100 msgs/sec depending on transaction load

## Known Limitations and Considerations

1. **Browser UI at scale**: The metrics view aggregates data; per-CP detailed logs are fetched on-demand to avoid memory bloat
2. **Log file size**: At 10k CPs with full logging, expect ~1–2 GB/hour; configure log rotation/sampling for long-running tests
3. **OCPP server capacity**: Ensure your target OCPP server can handle the message rate (load test incrementally)
4. **Graceful shutdown**: Sending StopTransaction for all active CPs before shutdown may take time

## Troubleshooting

### CPs not connecting
- Check OCPP server URL in UI (must be WebSocket protocol: `ws://` or `wss://`)
- Check logs for connection errors: `logs/ocpp_messages.log`
- Verify firewall/network access to OCPP server

### Remote start not working
- Verify `remote_start_profiles.csv` has entry for that CP
- Check auth token is valid
- Review logs for HTTP response from transaction service

### High memory usage
- Reduce number of active CPs
- Disable full message logging (use summary mode)
- Increase log rotation frequency
- Reduce per-CP log buffer size

## Future Enhancements

- [ ] Direct OCPP server mode (CPs connect to this simulator instead of real server)
- [ ] Custom OCPP message template editor
- [ ] WebSocket server mode for CP connections
- [ ] Detailed performance profiling and benchmarks
- [ ] Kubernetes deployment templates
- [ ] Prometheus metrics export

## Support

For issues or questions, check the logs in `logs/` directory or enable debug mode in UI.
