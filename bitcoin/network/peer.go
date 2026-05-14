package network

import (
	"bitcoin/protocol"
	"log"
	"net"
	"sync"
	"time"
)

// Peer owns one TCP connection to another blockchain node.
type Peer struct {
	manager *Manager

	conn     net.Conn
	codec    *protocol.Codec
	outbound bool
	writeMu  sync.Mutex

	nodeID  string
	p2pAddr string

	remoteNodeID  string
	remoteP2PAddr string

	handshakeDone   bool
	versionSent     bool
	versionReceived bool
	verackSent      bool
	verackReceived  bool
}

// PeerInfo is a stable snapshot of peer state for logs and APIs.
type PeerInfo struct {
	Addr              string `json:"addr"`
	P2PAddr           string `json:"p2p_addr"`
	NodeID            string `json:"node_id"`
	Outbound          bool   `json:"outbound"`
	HandshakeComplete bool   `json:"handshake_complete"`
}

// NewPeer creates a peer wrapper around an accepted or dialed TCP connection.
func NewPeer(conn net.Conn, outbound bool, nodeID string, p2pAddr string, manager *Manager) *Peer {
	return &Peer{
		manager:  manager,
		conn:     conn,
		codec:    protocol.NewCodec(conn),
		outbound: outbound,
		nodeID:   nodeID,
		p2pAddr:  p2pAddr,
	}
}

// Addr returns the active TCP connection address for this peer.
func (p *Peer) Addr() string {
	return p.conn.RemoteAddr().String()
}

// Info returns a snapshot of this peer's currently known metadata.
func (p *Peer) Info() PeerInfo {
	return PeerInfo{
		Addr:              p.Addr(),
		P2PAddr:           p.remoteP2PAddr,
		NodeID:            p.remoteNodeID,
		Outbound:          p.outbound,
		HandshakeComplete: p.handshakeDone,
	}
}

// AdvertisedAddr returns the peer's announced P2P address when known.
func (p *Peer) AdvertisedAddr() string {
	if p.remoteP2PAddr != "" {
		return p.remoteP2PAddr
	}
	return p.Addr()
}

// Run performs the handshake and then processes normal protocol messages.
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

// Close closes the underlying TCP connection and unblocks the peer read loop.
func (p *Peer) Close() error {
	return p.conn.Close()
}
