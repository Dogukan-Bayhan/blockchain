package network

import (
	"log"
	"net"
	"time"
)

const knownPeerDialInterval = 10 * time.Second

// dialBootstrap connects to the configured bootstrap peer.
func (m *Manager) dialBootstrap() {
	m.dialPeer(m.bootstrap, "bootstrap")
}

// dialPeer opens one outbound connection and tracks it as an active peer.
func (m *Manager) dialPeer(addr string, reason string) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(m.ctx, "tcp", addr)
	if err != nil {
		log.Printf("network: %s dial error %s: %v", reason, addr, err)
		return
	}

	select {
	case <-m.ctx.Done():
		conn.Close()
		return
	default:
	}

	log.Printf("network: connected to %s peer %s", reason, addr)
	peer := NewPeer(conn, true, m.nodeID, m.p2pAddr, m)
	m.addPeer(peer)
	go func() {
		defer m.removePeer(peer.Addr())
		peer.Run()
	}()
}

// connectKnownPeersPeriodically tries to connect to discovered peers over time.
func (m *Manager) connectKnownPeersPeriodically() {
	ticker := time.NewTicker(knownPeerDialInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.connectKnownPeers()
		case <-m.ctx.Done():
			return
		}
	}
}

// connectKnownPeers dials known addresses that are not already connected or dialing.
func (m *Manager) connectKnownPeers() {
	knownPeers := m.KnownPeers()
	for addr := range knownPeers {
		select {
		case <-m.ctx.Done():
			return
		default:
		}
		if !m.markKnownPeerDialing(addr) {
			continue
		}

		go func(addr string) {
			defer m.unmarkKnownPeerDialing(addr)
			m.dialPeer(addr, "known")
		}(addr)
	}
}

// markKnownPeerDialing reserves an address for one outbound dial attempt.
func (m *Manager) markKnownPeerDialing(addr string) bool {
	if addr == "" || m.isSelfAddr(addr) {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.ctx.Done():
		return false
	default:
	}

	if m.dialing[addr] {
		return false
	}

	for _, peer := range m.peers {
		if peer.Addr() == addr || peer.AdvertisedAddr() == addr {
			return false
		}
	}

	m.dialing[addr] = true
	return true
}

// unmarkKnownPeerDialing releases an address after its dial attempt finishes.
func (m *Manager) unmarkKnownPeerDialing(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.dialing, addr)
}

// isSelfAddr reports whether an advertised address points back to this node.
func (m *Manager) isSelfAddr(addr string) bool {
	if addr == m.p2pAddr {
		return true
	}

	localHost, localPort, localErr := net.SplitHostPort(m.p2pAddr)
	remoteHost, remotePort, remoteErr := net.SplitHostPort(addr)
	if localErr != nil || remoteErr != nil {
		return false
	}

	if localPort != remotePort {
		return false
	}

	return localHost == "" &&
		(remoteHost == "127.0.0.1" || remoteHost == "localhost" || remoteHost == "::1")
}
