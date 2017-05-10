package models

type (
	WebType int

	WebConfig struct {
		EtcdServers []string
		RpcxPort    uint16
		StartType   WebType
		MyIpAddr    string
	}
)

const (
	WebType_Manager WebType = iota
	WebType_Worker
)
