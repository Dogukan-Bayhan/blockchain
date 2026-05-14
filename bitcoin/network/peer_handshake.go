package network

import (
	"bitcoin/protocol"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// handleHandshakeMessage processes version/verack messages before normal traffic.
func (p *Peer) handleHandshakeMessage(msg protocol.Message) error {
	switch msg.Type {
	case protocol.MessageVersion:
		var payload protocol.VersionPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid version payload: %w", err)
		}

		log.Printf("network: received version from %s at %s", payload.NodeID, payload.P2PAddr)
		p.versionReceived = true
		p.remoteNodeID = payload.NodeID
		p.remoteP2PAddr = payload.P2PAddr

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

// completeHandshakeIfReady marks the peer ready once both sides exchanged version and verack.
func (p *Peer) completeHandshakeIfReady() {
	if p.handshakeDone {
		return
	}

	if p.versionSent && p.versionReceived && p.verackSent && p.verackReceived {
		p.handshakeDone = true
		p.conn.SetReadDeadline(time.Time{})
		log.Printf("network: handshake complete with %s", p.conn.RemoteAddr().String())

		// Only the dialing peer asks for addresses immediately to avoid duplicate
		// getaddr chatter during the first discovery round.
		if p.outbound {
			if err := p.sendGetAddr(); err != nil {
				log.Printf("network: send getaddr error: %v", err)
				return
			}
			log.Printf("network: sent getaddr to %s", p.conn.RemoteAddr().String())
		}
	}
}
