package network

import (
	"log"
	"net"
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
}

func NewManager(config Config) *Manager {
	return &Manager{
		nodeID:    config.NodeID,
		p2pAddr:   config.P2PAddr,
		bootstrap: config.Bootstrap,
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("network: accept error: %v", err)
			continue
		}

		log.Printf("network: inbound peer connected from %s", conn.RemoteAddr().String())
		go HandleConn(conn, false, m.nodeID, m.p2pAddr)
	}
}

func (m *Manager) dialBootstrap() {
	conn, err := net.Dial("tcp", m.bootstrap)
	if err != nil {
		log.Printf("network: bootstrap dial error %s: %v", m.bootstrap, err)
		return
	}

	log.Printf("network: connected to bootstrap peer %s", m.bootstrap)
	HandleConn(conn, true, m.nodeID, m.p2pAddr)
}