package configs

type SimulatorConfig struct {
	OCPPServerURL      string
	RemoteStartURL     string
	RemoteStopURL      string
	RemoteStartToken   string
	RemoteStopToken    string
	HeartbeatInterval  int
	MeterValueInterval int
	TransactionCutoff  int
	IdTag              string
	ChargePointVendor  string
	ChargePointModel   string
	FirmwareVersion    string
}
