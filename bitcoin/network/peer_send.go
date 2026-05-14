package network

import "bitcoin/protocol"

// sendVersion announces this node's identity and advertised P2P address.
func (p *Peer) sendVersion() error {
	payload := protocol.VersionPayload{
		NodeID:     p.nodeID,
		P2PAddr:    p.p2pAddr,
		BestHeight: 0,
	}

	msg, err := protocol.NewMessage(protocol.MessageVersion, payload)
	if err != nil {
		return err
	}

	return p.writeMessage(msg)
}

// sendVerack acknowledges the peer's version message.
func (p *Peer) sendVerack() error {
	msg, err := protocol.NewMessage(protocol.MessageVerAck, nil)
	if err != nil {
		return err
	}
	return p.writeMessage(msg)
}

// sendPong replies to a ping with the matching nonce payload.
func (p *Peer) sendPong(payload protocol.PongPayload) error {
	msg, err := protocol.NewMessage(protocol.MessagePong, payload)
	if err != nil {
		return err
	}
	return p.writeMessage(msg)
}

// sendGetAddr asks the peer for its known advertised peer addresses.
func (p *Peer) sendGetAddr() error {
	msg, err := protocol.NewMessage(protocol.MessageGetAddr, nil)
	if err != nil {
		return err
	}
	return p.writeMessage(msg)
}

// sendAddr sends known advertised peer addresses to the remote peer.
func (p *Peer) sendAddr(addrs []string) error {
	msg, err := protocol.NewMessage(protocol.MessageAddr, protocol.AddrPayload{
		Peers: addrs,
	})
	if err != nil {
		return err
	}
	return p.writeMessage(msg)
}

// writeMessage serializes outbound writes so multiple goroutines can send safely.
func (p *Peer) writeMessage(msg protocol.Message) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	return p.codec.WriteMessage(msg)
}
