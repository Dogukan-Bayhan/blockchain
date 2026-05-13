package network

import (
	"log"
	"net"
	"sync"
	"time"
)

type Config struct {
	NodeID    string
	P2PAddr   string
	Bootstrap string
}

type Manager struct {
	nodeID    string
	p2pAddr   string
	bootstrap string

	peers map[string]*Peer
	mu    sync.RWMutex
}

func NewManager(config Config) *Manager {
	return &Manager{
		nodeID:    config.NodeID,
		p2pAddr:   config.P2PAddr,
		bootstrap: config.Bootstrap,
		peers:     make(map[string]*Peer),
	}
}

func (m *Manager) Start() {
	listener, err := net.Listen("tcp", m.p2pAddr)
	if err != nil {
		log.Printf("network: listen error on %s: %v", m.p2pAddr, err)
		return
	}
	defer listener.Close()

	log.Printf("network: listening on %s", m.p2pAddr)

	if m.bootstrap != "" {
		go m.dialBootstrap()
	}
	go m.logPeersPeriodically()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("network: accept error: %v", err)
			continue
		}

		log.Printf("network: inbound peer connected from %s", conn.RemoteAddr().String())
		peer := NewPeer(conn, false, m.nodeID, m.p2pAddr)
		m.addPeer(peer)
		go func() {
			defer m.removePeer(peer.Addr())
			peer.Run()
		}()
	}
}

func (m *Manager) dialBootstrap() {
	conn, err := net.Dial("tcp", m.bootstrap)
	if err != nil {
		log.Printf("network: bootstrap dial error %s: %v", m.bootstrap, err)
		return
	}

	log.Printf("network: connected to bootstrap peer %s", m.bootstrap)
	peer := NewPeer(conn, true, m.nodeID, m.p2pAddr)
	m.addPeer(peer)
	go func() {
		defer m.removePeer(peer.Addr())
		peer.Run()
	}()
}

func (m *Manager) addPeer(peer *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.peers[peer.Addr()] = peer
	log.Printf("network: peer added addr=%s total_peers=%d", peer.Addr(), len(m.peers))
}

func (m *Manager) removePeer(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.peers, addr)
	log.Printf("network: peer removed addr=%s total_peers=%d", addr, len(m.peers))
}

func (m *Manager) PeerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.peers)
}

func (m *Manager) Peers() []PeerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peers := make([]PeerInfo, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer.Info())
	}
	return peers
}

func (m *Manager) LogPeers() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.peers) == 0 {
		log.Printf("network: peers empty")
		return
	}

	for _, peer := range m.peers {
		info := peer.Info()
		log.Printf("network: peer addr=%s outbound=%t handshake_complete=%t",
			info.Addr, info.Outbound, info.HandshakeComplete)
	}
}

func (m *Manager) logPeersPeriodically() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.LogPeers()
	}
}
