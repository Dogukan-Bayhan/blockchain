package protocol

import "encoding/json"

const (
	// MessageVersion starts the peer handshake and announces local node metadata.
	MessageVersion = "version"
	// MessageVerAck acknowledges that a peer's version message was accepted.
	MessageVerAck = "verack"
	// MessagePing checks whether an established peer connection is still alive.
	MessagePing = "ping"
	// MessagePong responds to a ping message with the same nonce.
	MessagePong = "pong"
	// MessageGetAddr asks a peer for the P2P addresses it currently knows.
	MessageGetAddr = "getaddr"
	// MessageAddr returns known peer addresses after a getaddr request.
	MessageAddr = "addr"

	// MessageInv announces transaction or block hashes without sending full data.
	MessageInv = "inv"
	// MessageGetData requests full transaction or block data for announced hashes.
	MessageGetData = "getdata"
	// MessageTx carries a full transaction payload.
	MessageTx = "tx"
	// MessageBlock carries a full block payload.
	MessageBlock = "block"
)

// Message is the common envelope exchanged by peers over the P2P connection.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// NewMessage creates a message envelope and serializes the optional payload.
func NewMessage(messageType string, payload any) (Message, error) {
	if payload == nil {
		return Message{Type: messageType}, nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Type: messageType,
		Data: data,
	}, nil
}

// VersionPayload identifies a node during the initial peer handshake.
type VersionPayload struct {
	NodeID     string `json:"node_id"`
	P2PAddr    string `json:"p2p_addr"`
	HTTPAddr   string `json:"http_addr"`
	BestHeight int    `json:"best_height"`
}

// PingPayload carries a nonce so a pong response can be matched to its ping.
type PingPayload struct {
	Nonce int64 `json:"nonce"`
}

// PongPayload carries the ping nonce back to the sender.
type PongPayload struct {
	Nonce int64 `json:"nonce"`
}

// AddrPayload carries known P2P peer addresses.
type AddrPayload struct {
	Peers []string `json:"peers"`
}
