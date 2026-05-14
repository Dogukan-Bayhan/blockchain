package node

import "time"

type Config struct {
	ID              string
	P2PAddr         string
	HTTPAddr        string
	BootstrapPeers  []string
	MaxPeers        int
	DataDir         string
	NetworkID       string
	ProtocolVersion int
	UserAgent       string
	enableMining    string
	ReadTimeout     time.Duration
	WriteTimeout 	time.Duration
	MessageMaxBytes int
}
