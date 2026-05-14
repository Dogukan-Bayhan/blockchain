package network

import (
	"context"
	"log"
	"net"
	"sync"
	"time"
)

// Config contains the local node identity and bootstrap settings.
type Config struct {
	NodeID    string
	P2PAddr   string
	Bootstrap string
}

// Manager listens for peers, dials bootstrap peers, and tracks active peers.
type Manager struct {
	nodeID    string
	p2pAddr   string
	bootstrap string

	listener net.Listener

	ctx    context.Context
	cancel context.CancelFunc

	peers      map[string]*Peer
	knownPeers map[string]time.Time
	dialing    map[string]bool
	mu         sync.RWMutex
}

// NewManager creates a manager with an empty active-peer set.
func NewManager(config Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		nodeID:     config.NodeID,
		p2pAddr:    config.P2PAddr,
		bootstrap:  config.Bootstrap,
		peers:      make(map[string]*Peer),
		knownPeers: make(map[string]time.Time),
		dialing:    make(map[string]bool),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins listening for inbound peers and optionally dials the bootstrap peer.
func (m *Manager) Start() {
	listener, err := net.Listen("tcp", m.p2pAddr)
	if err != nil {
		log.Printf("network: listen error on %s: %v", m.p2pAddr, err)
		return
	}
	defer listener.Close()

	m.listener = listener

	log.Printf("network: listening on %s", m.p2pAddr)

	if m.bootstrap != "" {
		go m.dialBootstrap()
	}
	go m.logPeersPeriodically()
	go m.discoverAddrsPeriodically()
	go m.connectKnownPeersPeriodically()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-m.ctx.Done():
				log.Printf("network: listener stopped")
				return
			default:
			}
			log.Printf("network: accept error: %v", err)
			continue
		}

		log.Printf("network: inbound peer connected from %s", conn.RemoteAddr().String())
		peer := NewPeer(conn, false, m.nodeID, m.p2pAddr, m)
		m.addPeer(peer)
		go func() {
			defer m.removePeer(peer.Addr())
			peer.Run()
		}()
	}
}

// Stop cancels background work, closes the listener, and disconnects peers.
func (m *Manager) Stop() {
	m.cancel()

	if m.listener != nil {
		if err := m.listener.Close(); err != nil {
			log.Printf("network: listener close error: %v", err)
		}
	}

	m.mu.Lock()
	peers := make([]*Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}
	m.mu.Unlock()

	for _, peer := range peers {
		if err := peer.Close(); err != nil {
			log.Printf("network: peer close error addr=%s: %v", peer.Addr(), err)
		}
	}
}
