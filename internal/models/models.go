package models

// ChargePoint represents a single charge point configuration
type ChargePoint struct {
	ChargeBoxId      string `csv:"chargeBoxId"`
	ConnectorCount   int    `csv:"connectorCount"`
	Vendor           string `csv:"chargePointVendor"`
	Model            string `csv:"chargePointModel"`
	FirmwareVersion  string `csv:"firmwareVersion"`
	MeterSerialNo    string `csv:"meterSerialNumber"`
	MeterType        string `csv:"meterType"`
}

// RemoteStartProfile represents a remote start profile
type RemoteStartProfile struct {
	ProfileName              string `csv:"profileName"`
	ChargeBoxId              string `csv:"chargeBoxId"`
	ConnectorId              string `csv:"connectorId"`
	LocationId               int    `csv:"locationId"`
	ChrgPointId              int    `csv:"chrgPointId"`
	ChrgPointConnectorDetId  int    `csv:"chrgPointConnectorDetId"`
	ChargingMethodId         int    `csv:"chargingMethodId"`
	ChargingValue            int    `csv:"chargingValue"`
	ChargingUnitId           int    `csv:"chargingUnitId"`
	IsReservationTrans       string `csv:"isReservationTrans"`
	ReservationId            int    `csv:"reservationId"`
	SelectedWalletType       int    `csv:"selectedWalletType"`
	VehicleId                int    `csv:"vehicleId"`
}

// Transaction represents an active transaction
type Transaction struct {
	TransactionId      int    `json:"transactionId"`
	ChargeBoxId        string `json:"chargeBoxId"`
	ConnectorId        int    `json:"connectorId"`
	IdTag              string `json:"idTag"`
	MeterStart         int    `json:"meterStart"`
	MeterStop          int    `json:"meterStop"`
	StartTime          string `json:"startTime"`
	StopTime           string `json:"stopTime"`
	Reason             string `json:"reason"`
	Status             string `json:"status"`
	HTTPTransactionId  string `json:"httpTransactionId,omitempty"`
	ChrgDetId          int    `json:"chrgDetId,omitempty"`
}

// CPStatus represents the status of a charge point instance
type CPStatus struct {
	ChargeBoxId        string                    `json:"chargeBoxId"`
	Status             string                    `json:"status"` // connected, booting, ready, charging, faulted, disconnecting
	ConnectorStates    map[int]ConnectorState    `json:"connectorStates"`
	ActiveTransactions map[int]*Transaction      `json:"activeTransactions"`
	LastBootTime       string                    `json:"lastBootTime"`
	LastHeartbeat      string                    `json:"lastHeartbeat"`
	MessageCount       int                       `json:"messageCount"`
	Config             *ChargePoint              `json:"config"`
}

// ConnectorState represents a connector's state
type ConnectorState struct {
	ConnectorId    int    `json:"connectorId"`
	Status         string `json:"status"`           // Available, Occupied, Unavailable, Faulted
	LastStatusTime string `json:"lastStatusTime"`
	CurrentPower   float64 `json:"currentPower"`    // kW
	TotalEnergy    float64 `json:"totalEnergy"`     // kWh
}

// OCPPMessage represents an OCPP message log entry
type OCPPMessage struct {
	Timestamp     string `json:"timestamp"`
	ChargeBoxId   string `json:"chargeBoxId"`
	ConnectorId   int    `json:"connectorId,omitempty"`
	Direction     string `json:"direction"`        // "->", "<-"
	MessageType   string `json:"messageType"`      // CALL, CALLRESULT, CALLERROR
	Action        string `json:"action"`           // BootNotification, Heartbeat, etc.
	Status        string `json:"status"`           // Accepted, Rejected, etc.
	PayloadSize   int    `json:"payloadSize"`
	ErrorCode     string `json:"errorCode,omitempty"`
	RawPayload    string `json:"rawPayload,omitempty"`
}

// MeterValue represents a meter value reading
type MeterValue struct {
	Timestamp string `json:"timestamp"`
	Values    []SampledValue `json:"values"`
}

// SampledValue represents a single sampled value
type SampledValue struct {
	Value   string `json:"value"`
	Context string `json:"context,omitempty"` // Sample, Transaction
	Format  string `json:"format,omitempty"`  // Raw, SignedData
	Measurand string `json:"measurand"`       // Energy.Active.Import.Register, Power.Active.Import, Temperature, etc.
	Phase   string `json:"phase,omitempty"`   // L1, L2, L3
	Location string `json:"location,omitempty"` // Inlet, Outlet, Body
	Unit    string `json:"unit,omitempty"`    // Wh, kWh, W, kW, A, V, C
}
