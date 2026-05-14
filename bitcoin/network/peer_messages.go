package network

import (
	"bitcoin/protocol"
	"encoding/json"
	"log"
)

// handleMessage processes post-handshake peer protocol messages.
func (p *Peer) handleMessage(msg protocol.Message) {
	switch msg.Type {
	case protocol.MessagePing:
		var payload protocol.PingPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("network: invalid ping payload from %s: %v", p.conn.RemoteAddr().String(), err)
			return
		}

		if err := p.sendPong(protocol.PongPayload{Nonce: payload.Nonce}); err != nil {
			log.Printf("network: send pong error: %v", err)
			return
		}
		log.Printf("network: received ping and sent pong to %s", p.conn.RemoteAddr().String())

	case protocol.MessagePong:
		var payload protocol.PongPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("network: invalid pong payload from %s: %v", p.conn.RemoteAddr().String(), err)
			return
		}
		log.Printf("network: received pong from %s nonce=%d", p.conn.RemoteAddr().String(), payload.Nonce)

	case protocol.MessageGetAddr:
		addrs := p.manager.PeerAddrs()
		if err := p.sendAddr(addrs); err != nil {
			log.Printf("network: send addr error: %v", err)
			return
		}
		log.Printf("network: received getaddr and sent %d addrs to %s", len(addrs), p.conn.RemoteAddr().String())

	case protocol.MessageAddr:
		var payload protocol.AddrPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("network: invalid addr payload from %s: %v", p.conn.RemoteAddr().String(), err)
			return
		}

		p.manager.AddKnownPeers(payload.Peers)

		log.Printf("network: received addr from %s peers=%v", p.conn.RemoteAddr().String(), payload.Peers)

	case protocol.MessageInv:
		var payload protocol.InventoryPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("network: invalid inv payload from %s: %v", p.conn.RemoteAddr().String(), err)
			return
		}
		log.Printf("network: received inv from %s items=%v", p.conn.RemoteAddr().String(), payload.Items)

	case protocol.MessageTx:
		// Transaction relay will validate and add the payload to the mempool.
		log.Printf("network: received tx from %s", p.conn.RemoteAddr().String())

	case protocol.MessageBlock:
		// Block relay will validate and connect the payload to the local chain.
		log.Printf("network: received block from %s", p.conn.RemoteAddr().String())

	case protocol.MessageGetData:
		var payload protocol.GetDataPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("network: invalid getdata payload from %s: %v", p.conn.RemoteAddr().String(), err)
			return
		}
		log.Printf("network: received getdata from %s items=%v", p.conn.RemoteAddr().String(), payload.Items)

	default:
		log.Printf("network: unknown message type %q from %s", msg.Type, p.conn.RemoteAddr().String())
	}
}
