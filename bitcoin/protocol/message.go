package protocol

import "encoding/json"

const (
	MessageVersion = "version"
	MessageVerAck  = "verack"
	MessagePing    = "ping"
	MessagePong    = "pong"
	MessageGetAddr = "getaddr"
	MessageAddr    = "addr"

	MessageInv     = "inv"
	MessageGetData = "getdata"
	MessageTx      = "tx"
	MessageBlock   = "block"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type VersionPayload struct {
	NodeID     string `json:"node_id"`
	P2PAddr    string `json:"p2p_addr"`
	HTTPAddr   string `json:"http_addr"`
	BestHeight int    `json:"best_height"`
}

type PingPayload struct {
	Nonce int64 `json:"nonce"`
}

type AddrPayload struct {
	Peers []string `json:"peers"`
}
