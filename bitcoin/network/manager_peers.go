package network

import (
	"log"
	"time"
)

// addPeer tracks an active TCP peer connection.
func (m *Manager) addPeer(peer *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.peers[peer.Addr()] = peer
	log.Printf("network: peer added addr=%s total_peers=%d", peer.Addr(), len(m.peers))
}

// removePeer deletes a peer after its connection closes.
func (m *Manager) removePeer(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.peers, addr)
	log.Printf("network: peer removed addr=%s total_peers=%d", addr, len(m.peers))
}

// PeerCount returns the number of currently tracked TCP peer connections.
func (m *Manager) PeerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.peers)
}

// Peers returns snapshots for all currently tracked peer connections.
func (m *Manager) Peers() []PeerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peers := make([]PeerInfo, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer.Info())
	}
	return peers
}

// PeerAddrs returns advertised P2P addresses for handshake-complete peers.
func (m *Manager) PeerAddrs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	addrs := make([]string, 0, len(m.peers))
	for _, peer := range m.peers {
		if !peer.handshakeDone {
			continue
		}
		addrs = append(addrs, peer.AdvertisedAddr())
	}
	return addrs
}

// readyPeers returns a snapshot of peers that completed the handshake.
func (m *Manager) readyPeers() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peers := make([]*Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		if !peer.handshakeDone {
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

// LogPeers writes the current peer set to the process log.
func (m *Manager) LogPeers() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.peers) == 0 {
		log.Printf("network: peers empty")
		return
	}

	for _, peer := range m.peers {
		info := peer.Info()
		log.Printf("network: peer addr=%s p2p_addr=%s node_id=%s outbound=%t handshake_complete=%t",
			info.Addr, info.P2PAddr, info.NodeID, info.Outbound, info.HandshakeComplete)
	}
}

// logPeersPeriodically logs active peer state at a fixed interval.
func (m *Manager) logPeersPeriodically() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.LogPeers()
		case <-m.ctx.Done():
			return
		}
	}
}
