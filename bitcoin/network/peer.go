package network

import (
	"bitcoin/protocol"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type Peer struct {
	conn     net.Conn
	codec    *protocol.Codec
	outbound bool

	nodeID  string
	p2pAddr string

	handshakeDone   bool
	versionSent     bool
	versionReceived bool
	verackSent      bool
	verackReceived  bool
}

type PeerInfo struct {
	Addr              string `json:"addr"`
	Outbound          bool   `json:"outbound"`
	HandshakeComplete bool   `json:"handshake_complete"`
}

func NewPeer(conn net.Conn, outbound bool, nodeID string, p2pAddr string) *Peer {
	return &Peer{
		conn:     conn,
		codec:    protocol.NewCodec(conn),
		outbound: outbound,
		nodeID:   nodeID,
		p2pAddr:  p2pAddr,
	}
}

func (p *Peer) Addr() string {
	return p.conn.RemoteAddr().String()
}

func (p *Peer) Info() PeerInfo {
	return PeerInfo{
		Addr:              p.Addr(),
		Outbound:          p.outbound,
		HandshakeComplete: p.handshakeDone,
	}
}

func (p *Peer) Run() {
	defer p.conn.Close()

	if p.outbound {
		if err := p.sendVersion(); err != nil {
			log.Printf("network: send version error: %v", err)
			return
		}
		p.versionSent = true
		log.Printf("network: sent version to %s", p.conn.RemoteAddr().String())
	}

	p.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	for {
		msg, err := p.codec.ReadMessage()
		if err != nil {
			log.Printf("network: read error from %s: %v", p.conn.RemoteAddr().String(), err)
			return
		}

		if !p.handshakeDone {
			if err := p.handleHandshakeMessage(msg); err != nil {
				log.Printf("network: %v", err)
				return
			}
			continue
		}

		p.handleMessage(msg)
	}
}

func (p *Peer) handleHandshakeMessage(msg protocol.Message) error {
	switch msg.Type {
	case protocol.MessageVersion:
		var payload protocol.VersionPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid version payload: %w", err)
		}

		log.Printf("network: received version from %s at %s", payload.NodeID, payload.P2PAddr)
		p.versionReceived = true

		if !p.versionSent {
			if err := p.sendVersion(); err != nil {
				return fmt.Errorf("send version error: %w", err)
			}
			p.versionSent = true
			log.Printf("network: sent version to %s", payload.NodeID)
		}

		if !p.verackSent {
			if err := p.sendVerack(); err != nil {
				return fmt.Errorf("send verack error: %w", err)
			}
			p.verackSent = true
			log.Printf("network: sent verack to %s", payload.NodeID)
		}

	case protocol.MessageVerAck:
		log.Printf("network: received verack from %s", p.conn.RemoteAddr().String())
		p.verackReceived = true

	default:
		return fmt.Errorf("message before handshake: %s", msg.Type)
	}

	p.completeHandshakeIfReady()
	return nil
}

func (p *Peer) completeHandshakeIfReady() {
	if p.handshakeDone {
		return
	}

	if p.versionSent && p.versionReceived && p.verackSent && p.verackReceived {
		p.handshakeDone = true
		p.conn.SetReadDeadline(time.Time{})
		log.Printf("network: handshake complete with %s", p.conn.RemoteAddr().String())
	}
}

func (p *Peer) handleMessage(msg protocol.Message) {
	switch msg.Type {
	case protocol.MessagePing:
		// pong gonder

	case protocol.MessageGetAddr:
		// addr gonder

	case protocol.MessageInv:
		// inventory isle

	case protocol.MessageTx:
		// transaction isle

	case protocol.MessageBlock:
		// block isle

	default:
		log.Printf("network: unknown message type %q from %s", msg.Type, p.conn.RemoteAddr().String())
	}
}

func (p *Peer) sendVersion() error {
	payload := protocol.VersionPayload{
		NodeID:     p.nodeID,
		P2PAddr:    p.p2pAddr,
		BestHeight: 0,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.codec.WriteMessage(protocol.Message{
		Type: protocol.MessageVersion,
		Data: data,
	})
}

func (p *Peer) sendVerack() error {
	return p.codec.WriteMessage(protocol.Message{Type: protocol.MessageVerAck})
}
