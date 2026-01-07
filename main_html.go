package main

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OCPP Simulator Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
            background: #f5f5f5;
            color: #333;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }
        
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 8px;
            margin-bottom: 30px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        
        header h1 {
            margin-bottom: 10px;
            font-size: 32px;
        }
        
        header p {
            opacity: 0.9;
            font-size: 14px;
        }
        
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        
        .card h2 {
            font-size: 16px;
            margin-bottom: 15px;
            color: #667eea;
            border-bottom: 2px solid #f0f0f0;
            padding-bottom: 10px;
        }
        
        .metrics {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 10px;
        }
        
        .metric {
            background: #f9f9f9;
            padding: 12px;
            border-radius: 4px;
            border-left: 3px solid #667eea;
        }
        
        .metric-label {
            font-size: 12px;
            color: #666;
            margin-bottom: 5px;
        }
        
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #333;
        }
        
        .form-group {
            margin-bottom: 15px;
        }
        
        label {
            display: block;
            margin-bottom: 5px;
            font-size: 14px;
            font-weight: 500;
            color: #555;
        }
        
        input[type="text"],
        input[type="number"],
        input[type="password"],
        textarea,
        select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
            font-family: inherit;
        }
        
        input[type="text"]:focus,
        input[type="number"]:focus,
        input[type="password"]:focus,
        textarea:focus,
        select:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }
        
        .button-group {
            display: flex;
            gap: 10px;
            margin-top: 20px;
        }
        
        button {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        
        .btn-primary {
            background: #667eea;
            color: white;
            flex: 1;
        }
        
        .btn-primary:hover {
            background: #5568d3;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        
        .btn-secondary {
            background: #f0f0f0;
            color: #333;
        }
        
        .btn-secondary:hover {
            background: #e0e0e0;
        }
        
        .btn-danger {
            background: #ef4444;
            color: white;
        }
        
        .btn-danger:hover {
            background: #dc2626;
        }
        
        .btn-success {
            background: #10b981;
            color: white;
        }
        
        .btn-success:hover {
            background: #059669;
        }
        
        .btn-sm {
            padding: 6px 12px;
            font-size: 12px;
        }
        
        .tabs {
            display: flex;
            gap: 0;
            margin-bottom: 20px;
            border-bottom: 2px solid #f0f0f0;
        }
        
        .tab {
            padding: 15px 20px;
            cursor: pointer;
            border: none;
            background: none;
            font-size: 14px;
            font-weight: 500;
            color: #999;
            transition: all 0.3s ease;
            border-bottom: 3px solid transparent;
            margin-bottom: -2px;
        }
        
        .tab.active {
            color: #667eea;
            border-bottom-color: #667eea;
        }
        
        .tab-content {
            display: none;
        }
        
        .tab-content.active {
            display: block;
        }
        
        .table-wrapper {
            overflow-x: auto;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
        }
        
        th {
            background: #f9f9f9;
            padding: 12px;
            text-align: left;
            font-weight: 600;
            color: #555;
            border-bottom: 2px solid #f0f0f0;
            font-size: 13px;
        }
        
        td {
            padding: 12px;
            border-bottom: 1px solid #f0f0f0;
            font-size: 13px;
        }
        
        tr:hover {
            background: #f9f9f9;
        }
        
        .status-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
        }
        
        .status-connected {
            background: #d1fae5;
            color: #065f46;
        }
        
        .status-ready {
            background: #dbeafe;
            color: #0c4a6e;
        }
        
        .status-charging {
            background: #fef3c7;
            color: #78350f;
        }
        
        .status-error {
            background: #fee2e2;
            color: #7f1d1d;
        }
        
        .file-input-wrapper {
            position: relative;
            overflow: hidden;
            display: inline-block;
            width: 100%;
        }
        
        .file-input-wrapper input[type="file"] {
            position: absolute;
            left: -9999px;
        }
        
        .file-input-label {
            display: block;
            padding: 10px 20px;
            background: #667eea;
            color: white;
            border-radius: 4px;
            cursor: pointer;
            text-align: center;
            font-weight: 600;
            transition: background 0.3s ease;
        }
        
        .file-input-label:hover {
            background: #5568d3;
        }
        
        .alert {
            padding: 15px;
            border-radius: 4px;
            margin-bottom: 15px;
            font-size: 14px;
        }
        
        .alert-success {
            background: #d1fae5;
            color: #065f46;
            border-left: 4px solid #10b981;
        }
        
        .alert-error {
            background: #fee2e2;
            color: #7f1d1d;
            border-left: 4px solid #ef4444;
        }
        
        .alert-info {
            background: #dbeafe;
            color: #0c4a6e;
            border-left: 4px solid #3b82f6;
        }
        
        .logs-container {
            background: #1e1e1e;
            color: #d4d4d4;
            padding: 15px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 12px;
            max-height: 400px;
            overflow-y: auto;
            line-height: 1.5;
        }
        
        .log-entry {
            margin-bottom: 5px;
            padding: 5px;
        }
        
        .log-timestamp {
            color: #858585;
        }
        
        .log-cp {
            color: #4ec9b0;
        }
        
        .log-direction {
            color: #ce9178;
        }
        
        .log-action {
            color: #569cd6;
        }
        
        .spinner {
            border: 3px solid #f0f0f0;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            width: 20px;
            height: 20px;
            animation: spin 1s linear infinite;
            display: inline-block;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .slider-container {
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        input[type="range"] {
            flex: 1;
            height: 6px;
            border-radius: 3px;
            background: #ddd;
            outline: none;
            -webkit-appearance: none;
        }
        
        input[type="range"]::-webkit-slider-thumb {
            -webkit-appearance: none;
            appearance: none;
            width: 20px;
            height: 20px;
            border-radius: 50%;
            background: #667eea;
            cursor: pointer;
        }
        
        input[type="range"]::-moz-range-thumb {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            background: #667eea;
            cursor: pointer;
            border: none;
        }
        
        .slider-value {
            min-width: 50px;
            text-align: right;
            font-weight: 600;
            color: #667eea;
        }
        
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.5);
        }
        
        .modal.active {
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .modal-content {
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            max-width: 600px;
            width: 90%;
            max-height: 80vh;
            overflow-y: auto;
            box-shadow: 0 4px 20px rgba(0,0,0,0.2);
        }
        
        .modal-header {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 20px;
            color: #333;
        }
        
        .modal-close {
            position: absolute;
            top: 20px;
            right: 20px;
            background: #f0f0f0;
            border: none;
            width: 32px;
            height: 32px;
            border-radius: 50%;
            cursor: pointer;
            font-size: 18px;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ OCPP Load Simulator</h1>
            <p>High-performance charge point load testing for OCPP 1.6</p>
        </header>
        
        <div class="grid">
            <div class="card">
                <h2>📊 Live Metrics</h2>
                <div class="metrics">
                    <div class="metric">
                        <div class="metric-label">Active CPs</div>
                        <div class="metric-value" id="activeCPs">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Transactions</div>
                        <div class="metric-value" id="activeTransactions">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Messages/sec</div>
                        <div class="metric-value" id="messagesPerSec">0.00</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Total Messages</div>
                        <div class="metric-value" id="totalMessages">0</div>
                    </div>
                </div>
            </div>
            
            <div class="card">
                <h2>⚙️ Configuration</h2>
                <div class="form-group">
                    <label>OCPP Server URL</label>
                    <input type="text" id="ocppUrl" placeholder="ws://localhost:8001" value="ws://localhost:8001">
                </div>
                <div class="form-group">
                    <label>Heartbeat Interval (seconds)</label>
                    <div class="slider-container">
                        <input type="range" id="heartbeatInterval" min="5" max="300" value="60">
                        <span class="slider-value"><span id="heartbeatValue">60</span>s</span>
                    </div>
                </div>
                <div class="form-group">
                    <label>Meter Value Interval (seconds)</label>
                    <div class="slider-container">
                        <input type="range" id="meterValueInterval" min="5" max="300" value="60">
                        <span class="slider-value"><span id="meterValueValue">60</span>s</span>
                    </div>
                </div>
                <button class="btn-primary" onclick="saveConfig()">💾 Save Config</button>
            </div>
            
            <div class="card">
                <h2>🚀 Start Simulation</h2>
                <div class="form-group">
                    <label>Number of Charge Points</label>
                    <input type="number" id="cpCount" min="1" max="10000" value="10">
                </div>
                <div class="form-group">
                    <label>Transaction Cutoff (minutes)</label>
                    <input type="number" id="txnCutoff" min="1" max="120" value="30">
                </div>
                <div class="button-group">
                    <button class="btn-success" onclick="startCPs()">▶️ Start CPs</button>
                    <button class="btn-danger" onclick="stopAllCPs()">⏹️ Stop All</button>
                </div>
            </div>
        </div>
        
        <div class="grid">
            <div class="card">
                <h2>📁 CSV Management</h2>
                <div class="form-group">
                    <label>Upload Chargepoints CSV</label>
                    <div class="file-input-wrapper">
                        <label class="file-input-label">Choose File
                            <input type="file" id="chargepointsFile" accept=".csv">
                        </label>
                    </div>
                </div>
                <div class="form-group">
                    <label>Upload Remote Start Profiles CSV</label>
                    <div class="file-input-wrapper">
                        <label class="file-input-label">Choose File
                            <input type="file" id="profilesFile" accept=".csv">
                        </label>
                    </div>
                </div>
                <div id="csvStatus" style="margin-top: 15px; font-size: 13px; color: #666;"></div>
            </div>
            
            <div class="card">
                <h2>🔗 Remote API Configuration</h2>
                <div class="form-group">
                    <label>Remote Start URL</label>
                    <input type="text" id="remoteStartUrl" placeholder="http://localhost:9097/api/remoteStart">
                </div>
                <div class="form-group">
                    <label>Remote Start Auth Token</label>
                    <input type="password" id="remoteStartToken" placeholder="Bearer eyJ...">
                </div>
                <div class="form-group">
                    <label>Remote Stop URL</label>
                    <input type="text" id="remoteStopUrl" placeholder="http://localhost:9097/api/remoteStop">
                </div>
                <div class="form-group">
                    <label>Remote Stop Auth Token</label>
                    <input type="password" id="remoteStopToken" placeholder="Bearer eyJ...">
                </div>
                <button class="btn-primary" onclick="saveRemoteConfig()">💾 Save Remote Config</button>
            </div>
        </div>
        
        <div class="card">
            <div class="tabs">
                <button class="tab active" onclick="switchTab('overview', event)">📋 Overview</button>
                <button class="tab" onclick="switchTab('cps', event)">⚡ Charge Points</button>
                <button class="tab" onclick="switchTab('transactions', event)">💳 Transactions</button>
                <button class="tab" onclick="switchTab('logs', event)">📝 Logs</button>
            </div>
            
            <div id="overview" class="tab-content active">
                <h2 style="margin-bottom: 20px;">System Status</h2>
                <div id="statusAlert"></div>
                <div class="metrics" style="grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));">
                    <div class="metric">
                        <div class="metric-label">CSV Chargepoints Loaded</div>
                        <div class="metric-value" id="cpLoaded">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Remote Profiles Loaded</div>
                        <div class="metric-value" id="profilesLoaded">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Messages Sent</div>
                        <div class="metric-value" id="msgSent">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Messages Received</div>
                        <div class="metric-value" id="msgRecv">0</div>
                    </div>
                </div>
            </div>
            
            <div id="cps" class="tab-content">
                <h2 style="margin-bottom: 20px;">Active Charge Points</h2>
                <div class="table-wrapper">
                    <table>
                        <thead>
                            <tr>
                                <th>Charge Box ID</th>
                                <th>Status</th>
                                <th>Connectors</th>
                                <th>Active Txns</th>
                                <th>Messages</th>
                                <th>Last Boot</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody id="cpTableBody">
                            <tr><td colspan="7" style="text-align: center; color: #999;">No active CPs</td></tr>
                        </tbody>
                    </table>
                </div>
            </div>
            
            <div id="transactions" class="tab-content">
                <h2 style="margin-bottom: 20px;">Active Transactions</h2>
                <div class="button-group">
                    <button class="btn-danger" onclick="remoteStopAll()">🛑 Remote Stop All</button>
                </div>
                <div class="table-wrapper" style="margin-top: 15px;">
                    <table>
                        <thead>
                            <tr>
                                <th>Transaction ID</th>
                                <th>Charge Box</th>
                                <th>Connector</th>
                                <th>IdTag</th>
                                <th>Start Time</th>
                                <th>Status</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody id="txnTableBody">
                            <tr><td colspan="7" style="text-align: center; color: #999;">No active transactions</td></tr>
                        </tbody>
                    </table>
                </div>
            </div>
            
            <div id="logs" class="tab-content">
                <h2 style="margin-bottom: 20px;">Recent Logs</h2>
                <div class="logs-container" id="logsContainer">
                    <p style="color: #999;">No logs yet. Start the simulation to see logs.</p>
                </div>
                <button class="btn-secondary" onclick="refreshLogs()" style="margin-top: 15px;">🔄 Refresh Logs</button>
            </div>
        </div>
    </div>
    
    <div id="notificationAlert"></div>
    
    <script>
        // Configuration
        let config = {
            ocppServerUrl: 'ws://localhost:8001',
            remoteStartUrl: 'http://localhost:9097/api/remoteStart',
            remoteStopUrl: 'http://localhost:9097/api/remoteStop',
            remoteStartToken: '',
            remoteStopToken: '',
            heartbeatInterval: 60,
            meterValueInterval: 60,
            transactionCutoffMinutes: 30,
            idTag: 'DEFAULT_RFID',
        };
        
        // Load saved config from localStorage
        function loadSavedConfig() {
            const saved = localStorage.getItem('ocppSimulatorConfig');
            if (saved) {
                config = { ...config, ...JSON.parse(saved) };
                updateConfigUI();
            }
        }
        
        // Update UI with current config
        function updateConfigUI() {
            document.getElementById('ocppUrl').value = config.ocppServerUrl;
            document.getElementById('remoteStartUrl').value = config.remoteStartUrl;
            document.getElementById('remoteStopUrl').value = config.remoteStopUrl;
            document.getElementById('remoteStartToken').value = config.remoteStartToken;
            document.getElementById('remoteStopToken').value = config.remoteStopToken;
            document.getElementById('heartbeatInterval').value = config.heartbeatInterval;
            document.getElementById('meterValueInterval').value = config.meterValueInterval;
            updateSliderValues();
        }
        
        function updateSliderValues() {
            document.getElementById('heartbeatValue').textContent = document.getElementById('heartbeatInterval').value;
            document.getElementById('meterValueValue').textContent = document.getElementById('meterValueInterval').value;
        }
        
        // Save configuration
        async function saveConfig() {
            config.ocppServerUrl = document.getElementById('ocppUrl').value;
            config.heartbeatInterval = parseInt(document.getElementById('heartbeatInterval').value);
            config.meterValueInterval = parseInt(document.getElementById('meterValueInterval').value);
            
            localStorage.setItem('ocppSimulatorConfig', JSON.stringify(config));
            
            const response = await fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config),
            });
            
            showNotification('Configuration saved!', 'success');
        }
        
        async function saveRemoteConfig() {
            config.remoteStartUrl = document.getElementById('remoteStartUrl').value;
            config.remoteStopUrl = document.getElementById('remoteStopUrl').value;
            config.remoteStartToken = document.getElementById('remoteStartToken').value;
            config.remoteStopToken = document.getElementById('remoteStopToken').value;
            
            localStorage.setItem('ocppSimulatorConfig', JSON.stringify(config));
            showNotification('Remote configuration saved!', 'success');
        }
        
        // File uploads
        document.getElementById('chargepointsFile').addEventListener('change', async (e) => {
            const file = e.target.files[0];
            if (!file) return;
            
            const formData = new FormData();
            formData.append('file', file);
            
            try {
                const response = await fetch('/api/csv/chargepoints', {
                    method: 'POST',
                    body: formData,
                });
                
                if (response.ok) {
                    showNotification('Chargepoints CSV uploaded successfully!', 'success');
                    await updateCSVStatus();
                } else {
                    const err = await response.text();
                    showNotification('Error: ' + err, 'error');
                }
            } catch (err) {
                showNotification('Upload failed: ' + err.message, 'error');
            }
            
            e.target.value = '';
        });
        
        document.getElementById('profilesFile').addEventListener('change', async (e) => {
            const file = e.target.files[0];
            if (!file) return;
            
            const formData = new FormData();
            formData.append('file', file);
            
            try {
                const response = await fetch('/api/csv/profiles', {
                    method: 'POST',
                    body: formData,
                });
                
                if (response.ok) {
                    showNotification('Remote start profiles uploaded successfully!', 'success');
                    await updateCSVStatus();
                } else {
                    const err = await response.text();
                    showNotification('Error: ' + err, 'error');
                }
            } catch (err) {
                showNotification('Upload failed: ' + err.message, 'error');
            }
            
            e.target.value = '';
        });
        
        // Update CSV status
        async function updateCSVStatus() {
            try {
                const response = await fetch('/api/csv/status');
                const data = await response.json();
                document.getElementById('cpLoaded').textContent = data.chargepoints_loaded || 0;
                document.getElementById('profilesLoaded').textContent = data.profiles_loaded || 0;
            } catch (err) {
                console.error('Failed to fetch CSV status:', err);
            }
        }
        
        // Start CPs
        async function startCPs() {
            const count = parseInt(document.getElementById('cpCount').value);
            if (isNaN(count) || count < 1) {
                showNotification('Please enter a valid CP count', 'error');
                return;
            }
            
            config.transactionCutoffMinutes = parseInt(document.getElementById('txnCutoff').value);
            
            try {
                const response = await fetch('/api/cps/start', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ count }),
                });
                
                if (response.ok) {
                    showNotification("Starting " + count + " charge points...", 'success');
                    startAutoRefresh();
                } else {
                    const err = await response.text();
                    showNotification('Error: ' + err, 'error');
                }
            } catch (err) {
                showNotification('Failed to start CPs: ' + err.message, 'error');
            }
        }
        
        // Stop all CPs
        async function stopAllCPs() {
            if (!confirm('Are you sure you want to stop all charge points?')) return;
            
            try {
                const response = await fetch('/api/cps/stop', { method: 'POST' });
                if (response.ok) {
                    showNotification('All charge points stopped', 'success');
                    stopAutoRefresh();
                } else {
                    showNotification('Failed to stop CPs', 'error');
                }
            } catch (err) {
                showNotification('Error: ' + err.message, 'error');
            }
        }
        
        // Remote stop all
        async function remoteStopAll() {
            if (!confirm('Are you sure you want to stop all active transactions?')) return;
            
            try {
                const response = await fetch('/api/remote/stop-all', { method: 'POST' });
                if (response.ok) {
                    showNotification('All transactions stopped', 'success');
                } else {
                    showNotification('Failed to stop transactions', 'error');
                }
            } catch (err) {
                showNotification('Error: ' + err.message, 'error');
            }
        }
        
        // Remote start all
        async function remoteStartAll() {
            const count = parseInt(prompt('How many transactions to start?', '1'));
            if (isNaN(count) || count < 1) {
                showNotification('Please enter a valid number', 'error');
                return;
            }
            
            try {
                const response = await fetch('/api/remote/start-all', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ count: count })
                });
                if (response.ok) {
                    showNotification(count + ' transactions started', 'success');
                    await refreshTransactions();
                } else {
                    showNotification('Failed to start transactions', 'error');
                }
            } catch (err) {
                showNotification('Error: ' + err.message, 'error');
            }
        }

        
        // Tab switching
        function switchTab(tabName, event) {
            // Remove active class from all tabs and tab contents
            document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));
            document.querySelectorAll('.tab').forEach(el => el.classList.remove('active'));
            
            // Add active class to selected tab content
            document.getElementById(tabName).classList.add('active');
            
            // Add active class to clicked tab button
            if (event && event.target) {
                event.target.classList.add('active');
            }
        }
        
        // Refresh metrics
        async function refreshMetrics() {
            try {
                const response = await fetch('/api/metrics');
                const data = await response.json();
                
                document.getElementById('activeCPs').textContent = data.activeCPs || 0;
                document.getElementById('activeTransactions').textContent = data.activeTransactions || 0;
                document.getElementById('messagesPerSec').textContent = (data.messagesPerSecond || 0).toFixed(2);
                document.getElementById('totalMessages').textContent = data.totalMessages || 0;
                document.getElementById('msgSent').textContent = data.messagesSent || 0;
                document.getElementById('msgRecv').textContent = data.messagesReceived || 0;
            } catch (err) {
                console.error('Failed to fetch metrics:', err);
            }
        }
        
        // Refresh CPs table
        async function refreshCPs() {
            try {
                const response = await fetch('/api/cps');
                const cps = await response.json() || [];
                
                const tbody = document.getElementById('cpTableBody');
                if (cps.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: #999;">No active CPs</td></tr>';
                    return;
                }
                
                tbody.innerHTML = cps.map(function(cp) {
                    return '<tr>' +
                        '<td><strong>' + cp.chargeBoxId + '</strong></td>' +
                        '<td><span class="status-badge status-' + cp.status.status + '">' + cp.status.status + '</span></td>' +
                        '<td>' + Object.keys(cp.status.connectors).length + '</td>' +
                        '<td>' + cp.status.activeTransactionCount + '</td>' +
                        '<td>' + cp.status.messageCount + '</td>' +
                        '<td style="font-size: 12px;">' + new Date(cp.status.lastBootTime).toLocaleTimeString() + '</td>' +
                        '<td>' +
                            '<button class="btn-sm btn-secondary" onclick="viewCPLogs(\'' + cp.chargeBoxId + '\')">Logs</button>' +
                            '<button class="btn-sm btn-danger" onclick="stopCP(\'' + cp.chargeBoxId + '\')">Stop</button>' +
                        '</td>' +
                    '</tr>';
                }).join('');
            } catch (err) {
                console.error('Failed to fetch CPs:', err);
            }
        }
        
        // Refresh transactions
        async function refreshTransactions() {
            try {
                const response = await fetch('/api/transactions');
                const txns = await response.json() || [];
                
                const tbody = document.getElementById('txnTableBody');
                if (txns.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: #999;">No active transactions</td></tr>';
                    return;
                }
                
                tbody.innerHTML = txns.map(function(txn) {
                    return '<tr>' +
                        '<td><strong>' + txn.transactionId + '</strong></td>' +
                        '<td>' + txn.chargeBoxId + '</td>' +
                        '<td>' + txn.connectorId + '</td>' +
                        '<td>' + txn.idTag + '</td>' +
                        '<td style="font-size: 12px;">' + new Date(txn.startTime).toLocaleTimeString() + '</td>' +
                        '<td><span class="status-badge status-charging">' + txn.status + '</span></td>' +
                        '<td>' +
                            '<button class="btn-sm btn-danger" onclick="stopTransaction(\'' + txn.chargeBoxId + '\', ' + txn.connectorId + ')">Stop</button>' +
                        '</td>' +
                    '</tr>';
                }).join('');
            } catch (err) {
                console.error('Failed to fetch transactions:', err);
            }
        }
        
        // Refresh logs
        async function refreshLogs() {
            try {
                const response = await fetch('/api/logs?limit=50');
                const logs = await response.json() || [];
                
                const container = document.getElementById('logsContainer');
                if (logs.length === 0) {
                    container.innerHTML = '<p style="color: #999;">No logs available</p>';
                    return;
                }
                
                container.innerHTML = logs.map(function(log) {
                    return '<div class="log-entry">' +
                        '<span class="log-timestamp">' + log.timestamp + '</span>' +
                        '<span class="log-cp">[' + log.chargeBoxId + ']</span>' +
                        '<span class="log-direction">' + log.direction + '</span>' +
                        '<span class="log-action">' + log.action + '</span>' +
                        '<span style="color: #98c379;">' + (log.status || 'CALL') + '</span>' +
                    '</div>';
                }).join('');
            } catch (err) {
                console.error('Failed to fetch logs:', err);
            }
        }
        
        // Stop specific CP
        async function stopCP(chargeBoxId) {
            try {
                const response = await fetch('/api/cps/' + chargeBoxId + '/stop', { method: 'POST' });
                if (response.ok) {
                    showNotification('CP stopped', 'success');
                    await refreshCPs();
                }
            } catch (err) {
                showNotification('Error: ' + err.message, 'error');
            }
        }
        
        // Stop transaction
        async function stopTransaction(chargeBoxId, connectorId) {
            try {
                const response = await fetch('/api/ocpp/' + chargeBoxId + '/' + connectorId + '/stop-transaction', { method: 'POST' });
                if (response.ok) {
                    showNotification('Transaction stopped', 'success');
                    await refreshTransactions();
                }
            } catch (err) {
                showNotification('Error: ' + err.message, 'error');
            }
        }
        
        // View CP logs
        async function viewCPLogs(chargeBoxId) {
            try {
                const response = await fetch('/api/logs/' + chargeBoxId);
                const logs = await response.json() || [];
                
                const container = document.getElementById('logsContainer');
                if (logs.length === 0) {
                    container.innerHTML = '<p style="color: #999;">No logs for ' + chargeBoxId + '</p>';
                } else {
                    container.innerHTML = logs.map(function(log) {
                        return '<div class="log-entry">' +
                            '<span class="log-timestamp">' + log.timestamp + '</span>' +
                            '<span class="log-direction">' + log.direction + '</span>' +
                            '<span class="log-action">' + log.action + '</span>' +
                        '</div>';
                    }).join('');
                }
                
                switchTab('logs', { target: document.querySelector('.tab:nth-child(4)') });
            } catch (err) {
                showNotification('Error: ' + err.message, 'error');
            }
        }
        
        // Show notification
        function showNotification(message, type = 'info') {
            const alert = document.createElement('div');
            alert.className = 'alert alert-' + type;
            alert.textContent = message;
            alert.style.position = 'fixed';
            alert.style.top = '20px';
            alert.style.right = '20px';
            alert.style.zIndex = '2000';
            alert.style.maxWidth = '400px';
            document.body.appendChild(alert);
            
            setTimeout(() => alert.remove(), 4000);
        }
        
        // Auto-refresh
        let autoRefreshInterval;
        
        function startAutoRefresh() {
            if (autoRefreshInterval) clearInterval(autoRefreshInterval);
            autoRefreshInterval = setInterval(() => {
                refreshMetrics();
                refreshCPs();
                refreshTransactions();
                refreshLogs();
            }, 2000);
        }
        
        function stopAutoRefresh() {
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
                autoRefreshInterval = null;
            }
        }
        
        // Slider handlers - Use 'input' event for live updates
        function updateSliderValues() {
            const heartbeatSlider = document.getElementById('heartbeatInterval');
            const meterSlider = document.getElementById('meterValueInterval');
            document.getElementById('heartbeatValue').textContent = heartbeatSlider.value;
            document.getElementById('meterValueValue').textContent = meterSlider.value;
        }
        
        document.addEventListener('DOMContentLoaded', function() {
            const heartbeatSlider = document.getElementById('heartbeatInterval');
            const meterSlider = document.getElementById('meterValueInterval');
            
            heartbeatSlider.addEventListener('input', updateSliderValues);
            meterSlider.addEventListener('input', updateSliderValues);
            
            // Initial load
            loadSavedConfig();
            updateCSVStatus();
            refreshMetrics();
            startAutoRefresh();
            
            // Cleanup on page unload
            window.addEventListener('beforeunload', () => {
                stopAutoRefresh();
            });
        });
    </script>
</body>
</html>`
