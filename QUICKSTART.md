# Quick Start Guide - OCPP Simulator

## Prerequisites
- Go 1.21 or higher installed
- Two CSV files: `chargepoints.csv` and `remote_start_profiles.csv`

## Installation & Setup

### 1. Clone/Setup Project
```bash
mkdir ocpp-simulator
cd ocpp-simulator
```

### 2. Copy CSV Files
Place these files in the project root:
- `chargepoints.csv` (sample provided)
- `remote_start_profiles.csv` (sample provided)

### 3. Download Dependencies
```bash
go mod download
```

### 4. Run the Simulator
```bash
go run main.go main_html.go
```

Server will start on `http://localhost:8080`

## Usage

### Web Interface (Recommended)

1. **Open Browser**: Navigate to `http://localhost:8080`

2. **Configuration Tab**:
   - Set OCPP Server URL (e.g., `ws://your-ocpp-server:8001`)
   - Configure heartbeat and meter value intervals
   - Save configuration

3. **Remote API Tab**:
   - Enter Remote Start URL
   - Enter Remote Stop URL
   - Add authorization tokens if needed
   - Save remote configuration

4. **CSV Management**:
   - Upload `chargepoints.csv` (or use default)
   - Upload `remote_start_profiles.csv`
   - Verify "CSV Chargepoints Loaded" and "Remote Profiles Loaded" show correct counts

5. **Start Simulation**:
   - Set number of charge points (1-10000)
   - Set transaction cutoff time (minutes)
   - Click "Start CPs"

6. **Monitor**:
   - Watch live metrics (Active CPs, Transactions, Message Rate)
   - View individual CP status and logs
   - Manage active transactions

### Command Line Interface

For scripting or testing, use curl:

```bash
# Set configuration
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "ocppServerUrl": "ws://ocpp-server:8001",
    "heartbeatInterval": 60,
    "meterValueInterval": 60,
    "transactionCutoffMinutes": 30
  }'

# Start 100 charge points
curl -X POST http://localhost:8080/api/cps/start \
  -H "Content-Type: application/json" \
  -d '{"count": 100}'

# Get current metrics
curl http://localhost:8080/api/metrics

# Stop all charge points
curl -X POST http://localhost:8080/api/cps/stop

# View logs
curl http://localhost:8080/api/logs?limit=50
```

## CSV Format Details

### chargepoints.csv
```csv
chargeBoxId,connectorCount,chargePointVendor,chargePointModel,firmwareVersion,meterSerialNumber,meterType
MHCGT000001_1,2,Delta,ACMini,1.0.3,MSN12345,ACMeter
RB-DN-5089730,3,ABB,TerraAC,2.1.0,MSN56789,DCMeter
```

**Fields**:
- `chargeBoxId`: Unique identifier for charge point (REQUIRED)
- `connectorCount`: Number of connectors (1-4, REQUIRED)
- `chargePointVendor`: Manufacturer name (optional, default: "Generic")
- `chargePointModel`: Model name (optional, default: "Simulator")
- `firmwareVersion`: Firmware version (optional, default: "1.0.0")
- `meterSerialNumber`: Meter serial (optional, auto-generated if empty)
- `meterType`: Meter type (optional, default: "ACMeter")

### remote_start_profiles.csv
```csv
profileName,chargeBoxId,connectorId,locationId,chrgPointId,chrgPointConnectorDetId,chargingMethodId,chargingValue,chargingUnitId,isReservationTrans,reservationId,selectedWalletType,vehicleId
Location5_Default,MHCGT000001_1,1,5,15,25,214,20,233,N,0,262,78
AnyCP_Profile,*,*,5,0,0,214,20,233,N,0,262,78
```

**Fields**:
- `profileName`: Unique profile name (REQUIRED)
- `chargeBoxId`: Target CP ID or "*" for any (REQUIRED)
- `connectorId`: Target connector or "*" for any (REQUIRED)
- Other fields: HTTP API parameters for remote start

**Matching Logic**:
- First exact match on `chargeBoxId + connectorId`
- Then fallback to `chargeBoxId + *`
- Then fallback to `* + *` (AnyCP_Profile)

## Key Features

### Per-CP Controls
Each active charge point shows:
- Current status (booting, ready, charging, etc.)
- Connector states
- Active transaction count
- Message count
- Last boot time
- Action buttons (Logs, Stop)

### Transaction Management
- **Local Start**: Manually start transaction from UI or API
- **Remote Start**: Trigger from HTTP API (if profile exists)
- **Local Stop**: Stop from UI or API
- **Remote Stop**: Stop via HTTP API (bulk or individual)
- **Auto-Stop**: Transactions stop after cutoff time

### OCPP Commands
Send individual commands per CP:
- BootNotification
- Heartbeat
- StatusNotification
- StartTransaction
- StopTransaction
- MeterValues (automatic during transactions)

### Load Testing
- **Gradual Ramp**: Start CPs slowly for realistic load
- **Bulk Operations**: Start/stop many CPs at once
- **Metrics**: Real-time message rate, active connections
- **Logging**: All OCPP traffic and transactions logged

## Troubleshooting

### CPs not connecting
**Issue**: Status shows "disconnected"
**Solution**:
- Verify OCPP Server URL is correct (ws:// or wss://)
- Check firewall allows connection
- Verify OCPP server is running
- Check logs for error messages

### Remote start fails
**Issue**: "Remote start failed" error
**Solution**:
- Verify chargeBoxId exists in `remote_start_profiles.csv`
- Check remote API URL is accessible
- Verify auth token is valid and not expired
- Check remote API logs for rejected requests

### High memory usage
**Issue**: Simulator using lots of RAM
**Solution**:
- Reduce number of active CPs
- Increase log rotation (config not available yet, edit code)
- Reduce heartbeat/meter intervals (to reduce message volume)
- Check if transactions are stopping properly

### Logs not appearing
**Issue**: No logs in UI
**Solution**:
- Logs are written to `logs/ocpp_messages.log` and `logs/transactions.log`
- Check file permissions in logs/ directory
- Verify simulator has write access to logs/ folder
- Check disk space is available

## Performance Tips

### For 5k-10k CPs
1. **Stagger startup**: Don't start all CPs at once
   - Start 1000, wait 30s, start next 1000, etc.
2. **Reduce logging detail**: Lower heartbeat interval or sample MeterValues
3. **Use dedicated server**: At least 2GB RAM, quad-core CPU
4. **Monitor metrics**: Watch messages/sec to avoid overwhelming your OCPP server
5. **Test gradually**: 100 CPs → 500 → 1000 → ... to find bottleneck

### Realistic Testing
- Set heartbeat to 60-120s (real CPs do this)
- Set meter values to 30-60s (during transactions)
- Use transaction cutoff of 15-60 minutes
- Randomize CP startup to simulate real-world stagger

## Logs and Monitoring

### File Logs
Located in `logs/` directory:

**ocpp_messages.log**: All OCPP traffic (JSONL format)
```json
{"timestamp":"2026-01-06T13:32:10Z","chargeBoxId":"MHCGT000001_1","direction":"->","action":"BootNotification","status":"CALL"}
```

**transactions.log**: Transaction events (JSONL format)
```json
{"timestamp":"2026-01-06T13:33:00Z","transactionId":1,"chargeBoxId":"MHCGT000001_1","connectorId":1,"event":"start"}
```

### UI Logs
- Real-time log viewer in Web UI
- Per-CP log inspection
- Configurable history (last 50-1000 entries)

### Metrics
Accessible at `/api/metrics`:
- Active CP count
- Active transaction count
- Messages sent/received
- Messages per second
- Uptime

## Getting Help

1. Check logs in `logs/` directory for errors
2. Review UI error messages and alerts
3. Verify CSV files have correct format (use samples as reference)
4. Confirm OCPP server connectivity with simple telnet test
5. Check remote API accessibility with curl

## Next Steps

- Implement custom OCPP message templates
- Add Prometheus metrics export
- Implement CP failure scenarios
- Add VIN/vehicle profile support
- Kubernetes deployment templates
