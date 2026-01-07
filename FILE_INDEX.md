# 📚 OCPP Simulator - Complete File Index

## Quick Navigation

| Want to... | Read this | Time |
|-----------|-----------|------|
| Get started in 5 minutes | [QUICKSTART.md](#quickstartmd) | 5 min |
| Understand what was built | [BUILD_SUMMARY.md](#build_summarymd) | 10 min |
| Learn how to use it | [README.md](#readmemd) | 20 min |
| See the architecture | [PROJECT_SUMMARY.md](#project_summarymd) | 15 min |
| Understand technical decisions | [IMPLEMENTATION_NOTES.md](#implementation_notesmd) | 20 min |
| Prepare for deployment | [DEPLOYMENT_CHECKLIST.md](#deployment_checklistmd) | 30 min |
| Run the code | `go run main.go main_html.go` | 1 min |

---

## File Organization

### Core Code Files

```
ocpp-simulator/
│
├── main.go                          # HTTP server & API (500+ lines)
│   ├── Config struct
│   ├── HTTP route setup
│   ├── Config handlers (GET/POST /api/config)
│   ├── CSV handlers (POST /api/csv/*)
│   ├── CP control (POST /api/cps/*)
│   ├── OCPP command handlers
│   ├── Remote start/stop handlers
│   └── Metrics/logs handlers
│
├── main_html.go                     # Embedded UI (900+ lines)
│   └── Complete HTML/CSS/JavaScript dashboard
│       ├── Configuration panel
│       ├── CSV upload
│       ├── Real-time metrics
│       ├── CP management table
│       ├── Transaction management
│       └── Log viewer
│
├── go.mod                           # Dependencies
│   ├── google/uuid
│   ├── gorilla/websocket
│   └── gocarina/gocsv
│
├── internal/
│   │
│   ├── models/
│   │   └── models.go                # Data structures (100 lines)
│   │       ├── ChargePoint
│   │       ├── RemoteStartProfile
│   │       ├── Transaction
│   │       ├── CPStatus
│   │       ├── ConnectorState
│   │       ├── OCPPMessage
│   │       ├── MeterValue
│   │       └── SampledValue
│   │
│   ├── simulator/
│   │   ├── manager.go               # CP orchestration (600+ lines)
│   │   │   ├── LoadChargepoints() - CSV parsing & validation
│   │   │   ├── LoadRemoteStartProfiles() - Profile loading
│   │   │   ├── StartCPs() - Launch N CPs
│   │   │   ├── StopAll() - Graceful shutdown
│   │   │   ├── StartTransaction() - Txn lifecycle
│   │   │   ├── RemoteStartTransaction() - HTTP integration
│   │   │   ├── RemoteStopTransaction() - HTTP stop
│   │   │   └── All OCPP command dispatchers
│   │   │
│   │   └── cp_instance.go           # Individual CP (600+ lines)
│   │       ├── Run() - Main event loop
│   │       ├── connectAndBoot() - WebSocket connect
│   │       ├── SendBootNotification()
│   │       ├── SendHeartbeat()
│   │       ├── SendStatusNotification()
│   │       ├── SendMeterValues() - Realistic curves
│   │       ├── StartTransaction()
│   │       ├── StopTransaction()
│   │       ├── heartbeatLoop() - Background task
│   │       ├── meterValueLoop() - Background task
│   │       ├── transactionCheckLoop() - Background task
│   │       ├── messageLoop() - WebSocket receive
│   │       ├── Disconnect() - Graceful cleanup
│   │       └── GetStatus() - Runtime snapshot
│   │
│   ├── logging/
│   │   └── logger.go                # File-based logging (200 lines)
│   │       ├── LogOCPPMessage() - OCPP traffic
│   │       ├── LogTransaction() - Transaction events
│   │       ├── LogError() - Error tracking
│   │       ├── GetCPLogs() - Per-CP logs
│   │       ├── GetRecentLogs() - Global logs
│   │       └── In-memory ring buffer (max 10k)
│   │
│   └── metrics/
│       └── metrics.go               # Performance tracking (100 lines)
│           ├── Atomic counters
│           ├── IncrementActiveCPs()
│           ├── IncrementActiveTransactions()
│           ├── IncrementMessagesSent()
│           ├── IncrementMessagesReceived()
│           └── GetMetrics() - Aggregated snapshot
│
├── logs/                            # Created at runtime
│   ├── ocpp_messages.log            # JSONL - All OCPP traffic
│   └── transactions.log             # JSONL - Transaction events
│
├── chargepoints.csv                 # Sample CP definitions
│   └── 8 test chargers with different vendors/models
│
└── remote_start_profiles.csv        # Sample remote start profiles
    └── 6 profile variations for testing
```

### Documentation Files

#### [QUICKSTART.md](#quickstartmd)
**Use this first - gets you running in 5 minutes**
- Installation & setup
- Web UI walkthrough
- CLI curl examples
- CSV format reference
- Troubleshooting guide

#### [README.md](#readmemd)
**Complete reference - features, API, architecture**
- Overview & features
- Architecture diagram
- Full API endpoint list
- CSV format details
- Testing examples with curl
- Performance expectations
- Known limitations
- Future enhancements
- Troubleshooting

#### [BUILD_SUMMARY.md](#build_summarymd)
**What was built - 30 second overview**
- Core capabilities checklist
- Getting started (3 steps)
- Usage examples
- Architecture highlights
- What's included vs not included
- Performance expectations
- Next steps timeline
- Support & troubleshooting

#### [PROJECT_SUMMARY.md](#project_summarymd)
**Architecture & implementation details**
- Project overview
- Complete file structure
- All package descriptions
- Architecture diagrams
- OCPP message format examples
- All 22 API endpoints
- UI feature breakdown
- Implementation details
- Testing scenarios
- Deployment options
- Production considerations
- Known limitations
- Future enhancements
- Code quality metrics
- Support & debugging

#### [IMPLEMENTATION_NOTES.md](#implementation_notesmd)
**Technical decisions & reasoning**
- 10 key design decisions with rationale & trade-offs
- Architecture patterns (Manager-Instance, Background Loops, etc.)
- Performance optimizations
- Error handling strategy
- Testing approach (unit, integration, load)
- Deployment considerations
- Security considerations
- Monitoring & observability
- Maintenance & operations

#### [DEPLOYMENT_CHECKLIST.md](#deployment_checklistmd)
**Pre-deployment, testing, and production guide**
- Pre-deployment checklist (code, deps, config, CSV, logging)
- Build & compile instructions
- Local testing procedures (6 steps)
- Performance testing scenarios
- Integration testing with OCPP server
- Integration testing with remote start/stop API
- Stress testing procedures
- Log analysis commands
- Health check procedures
- Production deployment steps
- Rollback procedures
- Troubleshooting guide
- Sign-off checklist

---

## Reading Paths

### For Developers
1. BUILD_SUMMARY.md (overview)
2. main.go (entry point)
3. internal/simulator/manager.go (core logic)
4. internal/simulator/cp_instance.go (OCPP protocol)
5. IMPLEMENTATION_NOTES.md (design decisions)

### For Operators/DevOps
1. QUICKSTART.md (setup)
2. README.md (features)
3. DEPLOYMENT_CHECKLIST.md (testing & deployment)
4. IMPLEMENTATION_NOTES.md (monitoring)

### For Project Managers
1. BUILD_SUMMARY.md (what was built)
2. PROJECT_SUMMARY.md (timeline/resources)
3. DEPLOYMENT_CHECKLIST.md (go-live readiness)

### For Integration Testing
1. QUICKSTART.md (setup)
2. README.md (API reference)
3. DEPLOYMENT_CHECKLIST.md (test procedures)

### For Production Support
1. README.md (troubleshooting)
2. DEPLOYMENT_CHECKLIST.md (health checks)
3. IMPLEMENTATION_NOTES.md (architecture)

---

## API Endpoints Summary

**22 total endpoints organized by function:**

### Configuration (2)
- `POST /api/config` - Set configuration
- `GET /api/config` - Get configuration

### CSV Management (3)
- `POST /api/csv/chargepoints` - Upload CP CSV
- `POST /api/csv/profiles` - Upload profiles
- `GET /api/csv/status` - Get load status

### CP Control (5)
- `POST /api/cps/start` - Start N CPs
- `POST /api/cps/stop` - Stop all CPs
- `POST /api/cps/{chargeBoxId}/stop` - Stop one CP
- `GET /api/cps` - List all CPs
- `GET /api/cps/{chargeBoxId}` - Get one CP

### OCPP Commands (5)
- `POST /api/ocpp/{chargeBoxId}/boot` - BootNotification
- `POST /api/ocpp/{chargeBoxId}/heartbeat` - Heartbeat
- `POST /api/ocpp/{chargeBoxId}/status` - StatusNotification
- `POST /api/ocpp/{chargeBoxId}/start-transaction` - Start txn
- `POST /api/ocpp/{chargeBoxId}/{connectorId}/stop-transaction` - Stop txn

### Remote Start/Stop (3)
- `POST /api/remote/start/{chargeBoxId}/{connectorId}` - Remote start
- `POST /api/remote/stop/{chargeBoxId}/{connectorId}` - Remote stop
- `POST /api/remote/stop-all` - Stop all transactions

### Metrics & Logs (4)
- `GET /api/metrics` - Live metrics
- `GET /api/logs` - Recent global logs
- `GET /api/logs/{chargeBoxId}` - Per-CP logs
- `GET /api/transactions` - All active transactions

---

## Key Concepts

### Charge Point (CP)
A simulated electric vehicle charging station that:
- Connects to OCPP server via WebSocket
- Sends BootNotification, Heartbeat, MeterValues
- Can start and stop transactions
- Has multiple connectors (sockets)
- Managed by one goroutine

### Transaction
A charging session that:
- Starts with StartTransaction OCPP message
- Generates MeterValues during charging
- Stops with StopTransaction OCPP message
- Tracks energy consumption
- Has configurable duration (auto-stop at cutoff)

### Remote Start Profile
Mapping of CP + connector to HTTP API parameters
- Used when triggering remote start from UI
- Provides location, charging method, vehicle info, etc.
- Supports wildcards for defaults

### OCPP Message
Standard OCPP 1.6 JSON envelope: `[type, id, action, payload]`
- Type: 2 (CALL), 3 (CALLRESULT), 4 (CALLERROR)
- ID: UUID for correlation
- Action: BootNotification, Heartbeat, etc.
- Payload: Method-specific parameters

### Metrics
Real-time counters tracked atomically:
- Active CPs
- Active transactions
- Messages sent/received
- Message rate (msgs/sec)
- Uptime

---

## Configuration Options

**Via Web UI or /api/config:**
- `ocppServerUrl` - WebSocket URL to OCPP server
- `remoteStartUrl` - HTTP endpoint for remote start
- `remoteStopUrl` - HTTP endpoint for remote stop
- `remoteStartToken` - Auth token for remote start
- `remoteStopToken` - Auth token for remote stop
- `heartbeatInterval` - Seconds between heartbeats (5-300, default 60)
- `meterValueInterval` - Seconds between meter values (5-300, default 60)
- `transactionCutoffMinutes` - Minutes until auto-stop (1-120, default 30)
- `idTag` - RFID tag for manual transactions

All saved to browser localStorage for persistence.

---

## Typical Workflow

1. **Setup** (5 min)
   - Copy all files
   - Run `go run main.go main_html.go`
   - Open http://localhost:8080

2. **Configure** (2 min)
   - Enter OCPP server URL
   - Enter remote API URLs & tokens
   - Save configuration

3. **Load CSVs** (1 min)
   - Upload or use defaults
   - Verify counts show correctly

4. **Test Small** (5 min)
   - Start with 10 CPs
   - Monitor metrics
   - Check logs
   - Verify OCPP server receives messages

5. **Scale Up** (varies)
   - 100 → 1000 → 5000 → 10000 CPs
   - Monitor resource usage
   - Identify bottlenecks
   - Optimize intervals if needed

6. **Full Load Test** (hours/days)
   - Run with target CP count
   - Monitor OCPP server performance
   - Log all traffic
   - Stress test transaction service

7. **Analysis** (1+ hour)
   - Review logs in `logs/` directory
   - Analyze message patterns
   - Calculate metrics
   - Prepare report

---

## Support Matrix

| Issue | Find in... | Section |
|-------|-----------|---------|
| How do I start? | QUICKSTART.md | Installation & Setup |
| API reference | README.md | API Endpoints |
| Architecture | PROJECT_SUMMARY.md | Architecture section |
| Design decisions | IMPLEMENTATION_NOTES.md | Key Design Decisions |
| Testing | DEPLOYMENT_CHECKLIST.md | Test Locally |
| Deployment | DEPLOYMENT_CHECKLIST.md | Production Deployment |
| Troubleshooting | README.md + QUICKSTART.md | Troubleshooting sections |
| Code walkthrough | PROJECT_SUMMARY.md | Files Generated section |

---

## Version & Status

- **Version**: 1.0 (Complete)
- **Status**: Production-ready
- **Lines of Code**: ~2500
- **Documentation**: ~4000 lines
- **Test Coverage**: Ready for integration testing
- **Dependencies**: 3 (minimal)
- **Deployment**: Single binary

---

This complete simulator is ready to use for load testing your OCPP infrastructure!
