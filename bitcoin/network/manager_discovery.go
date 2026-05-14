package network

import (
	"log"
	"time"
)

const addressDiscoveryInterval = 30 * time.Second

// discoverAddrsPeriodically asks connected peers for their known peer addresses.
func (m *Manager) discoverAddrsPeriodically() {
	ticker := time.NewTicker(addressDiscoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.requestAddrsFromPeers()
		case <-m.ctx.Done():
			return
		}
	}
}

// requestAddrsFromPeers sends getaddr to all handshake-complete peers.
func (m *Manager) requestAddrsFromPeers() {
	peers := m.readyPeers()
	if len(peers) == 0 {
		return
	}

	for _, peer := range peers {
		select {
		case <-m.ctx.Done():
			return
		default:
		}
		if err := peer.sendGetAddr(); err != nil {
			log.Printf("network: periodic getaddr error addr=%s: %v", peer.Addr(), err)
			continue
		}
		log.Printf("network: sent periodic getaddr to %s", peer.Addr())
	}
}

// AddKnownPeer records an advertised peer address and the time it was seen.
func (m *Manager) AddKnownPeer(addr string) {
	if addr == "" {
		return
	}

	if m.isSelfAddr(addr) {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.knownPeers[addr] = time.Now()
	log.Printf("network: known peer added addr=%s total_known_peers=%d", addr, len(m.knownPeers))
}

// AddKnownPeers records multiple advertised peer addresses.
func (m *Manager) AddKnownPeers(addrs []string) {
	for _, addr := range addrs {
		m.AddKnownPeer(addr)
	}
}

// KnownPeers returns a copy of known peer addresses and their last seen times.
func (m *Manager) KnownPeers() map[string]time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peers := make(map[string]time.Time, len(m.knownPeers))
	for addr, seenAt := range m.knownPeers {
		peers[addr] = seenAt
	}

	return peers
}
