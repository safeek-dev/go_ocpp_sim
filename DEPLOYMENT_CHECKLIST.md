# Deployment & Testing Checklist

## Pre-Deployment Checklist

### Code Review
- [ ] All files have correct imports
- [ ] No hardcoded credentials in code
- [ ] Error handling is comprehensive
- [ ] Logging is properly configured
- [ ] Code follows Go conventions

### Dependencies
- [ ] `go mod tidy` run successfully
- [ ] No unused dependencies
- [ ] All imports are from reputable packages
- [ ] No version conflicts

### Configuration
- [ ] Default config values are sensible
- [ ] Config can be saved to localStorage
- [ ] Remote URLs are configurable
- [ ] Auth tokens can be entered securely

### CSV Files
- [ ] `chargepoints.csv` is valid and present
- [ ] `remote_start_profiles.csv` is valid and present
- [ ] Both files have correct headers
- [ ] No duplicate chargeBoxIds in chargepoints.csv
- [ ] Profiles reference existing CPs

### Logging & Storage
- [ ] `logs/` directory will be created automatically
- [ ] File permissions allow write access
- [ ] Log rotation strategy is clear (future enhancement)
- [ ] Disk space is available for logs

## Build & Compile

### Single Binary Build
```bash
cd ocpp-simulator
go mod download
go build -o ocpp-simulator main.go main_html.go
```

### Verify Binary
```bash
./ocpp-simulator --help 2>&1 | head -1
# Should show "Starting OCPP Simulator..."
```

### File Size
```bash
ls -lh ocpp-simulator
# Binary should be ~10-15 MB
```

## Test Locally

### 1. Start Server
```bash
./ocpp-simulator
# Expected output:
# Starting OCPP Simulator on http://localhost:8080
```

### 2. Verify Web UI
```bash
curl http://localhost:8080 -s | head -20
# Should return HTML starting with <!DOCTYPE html>
```

### 3. Check API
```bash
curl http://localhost:8080/api/metrics
# Should return JSON with metrics fields
```

### 4. Verify CSV Loading
```bash
curl http://localhost:8080/api/csv/status
# Should return {"chargepoints_loaded": 8, "profiles_loaded": 6}
```

### 5. Test CP Start (with mock OCPP server)
```bash
# In one terminal, create a simple mock WebSocket server:
# (Or use your real OCPP server if available)

# In another terminal:
curl -X POST http://localhost:8080/api/cps/start \
  -H "Content-Type: application/json" \
  -d '{"count": 5}'
# Expected: JSON response with status ok or error (expected if OCPP server not running)
```

### 6. Stop Server
```bash
# Press Ctrl+C in the server terminal
# Expected: Graceful shutdown message
```

## Performance Testing

### Small Load (100 CPs, 5 min test)
```bash
# 1. Start simulator
./ocpp-simulator

# 2. In browser: http://localhost:8080
#    - Start 100 CPs
#    - Monitor metrics for 5 minutes
#    - Check logs are being written

# 3. Verify resource usage
ps aux | grep ocpp-simulator
# Note: %CPU and %MEM values

# 4. Check logs
wc -l logs/ocpp_messages.log
wc -l logs/transactions.log

# 5. Calculate message rate
# Divide log lines by 5 minutes (300 seconds) = msgs/sec
```

### Medium Load (1000 CPs, 10 min test)
```bash
# Same as above but with 1000 CPs and 10 min duration
# Expected: 30-50 msgs/sec, ~300 MB RAM
```

### Large Load (5000 CPs, 15 min test)
```bash
# Same as above but with 5000 CPs and 15 min duration
# Expected: 100+ msgs/sec, ~500 MB RAM
# Monitor to ensure OCPP server doesn't get overwhelmed
```

## Integration Testing

### With Real OCPP Server

#### Setup
1. Configure OCPP server URL in UI
2. Ensure server has WebSocket endpoint
3. Configure auth if needed

#### Test Sequence
```bash
# 1. Check OCPP server is running
curl http://ocpp-server:8080/health || echo "Check server"

# 2. Start simulator with 10 CPs
# Via UI: Set count=10, click "Start CPs"

# 3. Monitor OCPP server logs for:
#    - BootNotification messages
#    - Heartbeat messages
#    - Connection count

# 4. Verify via OCPP server's metrics:
#    - Connected CPs = 10
#    - Received messages = expected count
#    - No errors

# 5. Stop simulator
# Via UI: Click "Stop All"

# 6. Verify in OCPP server:
#    - All CPs disconnected gracefully
#    - No orphaned connections
```

### With Remote Start/Stop API

#### Setup
1. Configure Remote Start URL in UI
2. Configure Remote Stop URL in UI
3. Set auth tokens if needed
4. Upload remote_start_profiles.csv with valid endpoints

#### Test Sequence
```bash
# 1. Start simulator with 10 CPs
# 2. Wait for all to reach "ready" status
# 3. For first CP with profile:
#    - Click "Remote Start"
#    - Check API server logs for HTTP request
#    - Verify request body matches profile
#    - Check auth header is present if configured

# 4. Monitor:
#    - Transaction should appear in "Transactions" tab
#    - MeterValues should start flowing
#    - Energy should increase

# 5. For active transaction:
#    - Click "Remote Stop"
#    - Check API server received stop request
#    - Verify transaction stops within 1-2 seconds

# 6. Test "Stop All":
#    - Start multiple transactions
#    - Click "Remote Stop All"
#    - All transactions should stop
```

## Stress Testing (Advanced)

### Max Connection Test
```bash
# Goal: Find maximum concurrent CPs
# Start with 1000, increase by 1000 each iteration

for count in 1000 2000 3000 4000 5000 6000 7000 8000 9000 10000; do
  echo "Testing $count CPs..."
  
  # Start CPs
  curl -X POST http://localhost:8080/api/cps/start \
    -H "Content-Type: application/json" \
    -d "{\"count\": $count}" 2>/dev/null
  
  # Wait for connections
  sleep 30
  
  # Check metrics
  curl http://localhost:8080/api/metrics | jq '.activeCPs'
  
  # Check resource usage
  ps aux | grep ocpp-simulator | head -1
  
  # Stop
  curl -X POST http://localhost:8080/api/cps/stop 2>/dev/null
  sleep 10
done
```

### Message Rate Test
```bash
# Goal: Find maximum message throughput
# Monitor msgs/sec as you increase CP count and reduce intervals

# Run with different configurations:
# - 1000 CPs, 60s heartbeat, 60s meter → ~33 msgs/sec
# - 2000 CPs, 30s heartbeat, 30s meter → ~133 msgs/sec
# - 5000 CPs, 10s heartbeat, 10s meter → ~1000 msgs/sec

# Find the point where:
# - OCPP server shows warnings/errors
# - Latency increases significantly
# - CPU on simulator hits 80%+ usage
```

### Memory Stability Test
```bash
# Goal: Detect memory leaks
# Run for 1 hour with consistent load, monitor memory

./ocpp-simulator &
SERVER_PID=$!
sleep 5

# Take initial memory reading
INITIAL_MEM=$(ps -p $SERVER_PID -o %mem= | tr -d ' ')
echo "Initial memory: ${INITIAL_MEM}%"

# Start 1000 CPs
curl -X POST http://localhost:8080/api/cps/start \
  -H "Content-Type: application/json" \
  -d '{"count": 1000}'

# Monitor every 5 minutes for 60 minutes
for i in {1..12}; do
  sleep 300
  CURRENT_MEM=$(ps -p $SERVER_PID -o %mem= | tr -d ' ')
  echo "$(date): Memory: ${CURRENT_MEM}%"
done

# Stop
curl -X POST http://localhost:8080/api/cps/stop

# Final reading should not be significantly higher than peak during run
kill $SERVER_PID
```

## Log Analysis

### Count Messages
```bash
# OCPP messages
wc -l logs/ocpp_messages.log

# Transactions
wc -l logs/transactions.log
```

### Filter by CP
```bash
# Get all messages for one CP
grep '"MHCGT000001_1"' logs/ocpp_messages.log | wc -l

# Get all transactions for one CP
grep '"MHCGT000001_1"' logs/transactions.log
```

### Find Errors
```bash
# Look for errors in logs
grep -i error logs/ocpp_messages.log | head -10

# Count by type
grep '"level":"ERROR"' logs/ocpp_messages.log | wc -l
```

### Parse JSONL
```bash
# Use jq to analyze
cat logs/ocpp_messages.log | jq '.action' | sort | uniq -c

# Get message rate per hour
cat logs/ocpp_messages.log | jq -r '.timestamp' | \
  xargs -I {} date -d {} +%Y-%m-%d-%H | sort | uniq -c
```

## Health Checks

### Endpoint Health
```bash
# Dashboard
curl -s http://localhost:8080 | grep -q "OCPP" && echo "UI: OK" || echo "UI: FAIL"

# Metrics
curl -s http://localhost:8080/api/metrics | jq .activeCPs >/dev/null && echo "Metrics: OK" || echo "Metrics: FAIL"

# Config
curl -s http://localhost:8080/api/config | grep ocppServerUrl >/dev/null && echo "Config: OK" || echo "Config: FAIL"
```

### Resource Checks
```bash
# Disk space
df -h logs/ | tail -1

# Memory
free -h

# CPU load
uptime

# Open files
lsof -p $(pgrep ocpp-simulator) | wc -l
```

### Connectivity Checks
```bash
# OCPP server reachable?
nc -zv ocpp-server 8001

# Remote API reachable?
curl -I http://remote-api/endpoint
```

## Production Deployment

### Pre-Deploy
- [ ] All tests pass locally
- [ ] Load tests completed
- [ ] Documentation reviewed
- [ ] Team trained on operation
- [ ] Rollback plan documented

### Deploy
- [ ] Copy binary to server
- [ ] Copy CSV files to server
- [ ] Ensure logs/ directory exists
- [ ] Start server with nohup or systemd
- [ ] Verify metrics visible
- [ ] Verify logs being written

### Post-Deploy
- [ ] Monitor metrics for 1 hour
- [ ] Check for errors in logs
- [ ] Verify OCPP server receives messages
- [ ] Test remote start/stop functionality
- [ ] Confirm graceful shutdown works

### Rollback
- [ ] Stop running simulator: `kill <pid>`
- [ ] Keep old logs for analysis
- [ ] Revert to previous binary version
- [ ] Restart simulator
- [ ] Verify recovery

## Troubleshooting Guide

### Issue: CPs not connecting
```bash
# Check OCPP server URL
curl http://localhost:8080/api/config | jq .ocppServerUrl

# Try to connect manually
wscat -c ws://ocpp-server:8001

# Check firewall
telnet ocpp-server 8001
```

### Issue: No metrics showing
```bash
# Check API
curl http://localhost:8080/api/metrics

# Check if CPs started
curl http://localhost:8080/api/cps | jq length

# Check logs for errors
tail -20 logs/ocpp_messages.log
```

### Issue: High memory usage
```bash
# Current memory
ps aux | grep ocpp-simulator

# Number of CPs running
curl http://localhost:8080/api/metrics | jq .activeCPs

# Stop all CPs
curl -X POST http://localhost:8080/api/cps/stop

# Monitor memory drop
watch 'ps aux | grep ocpp-simulator'
```

### Issue: Slow message rate
```bash
# Check OCPP server load
# Check network latency
ping -c 10 ocpp-server

# Reduce CP count
curl -X POST http://localhost:8080/api/cps/stop

# Increase intervals (reduce message frequency)
# Restart with smaller load
```

## Sign-Off Checklist

- [ ] Code compiles without warnings
- [ ] All endpoints tested
- [ ] UI responsive and intuitive
- [ ] CSV validation working
- [ ] Logging functional
- [ ] Metrics accurate
- [ ] Remote start/stop integrated
- [ ] Graceful shutdown verified
- [ ] Documentation complete
- [ ] Team trained
- [ ] Ready for production

**Signed by**: ________________  
**Date**: ________________  
**Environment**: [ ] Local [ ] Dev [ ] Staging [ ] Production
