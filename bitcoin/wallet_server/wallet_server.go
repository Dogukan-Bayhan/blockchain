package main

// tempDir is the template path used when running from the repository root.
const tempDir = "wallet_server/templates"

// WalletServer exposes wallet creation and transaction signing over HTTP.
type WalletServer struct {
	port    uint16
	gateway string
}

// NewWalletServer creates a wallet server that forwards chain requests to a gateway.
func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

// Port returns the wallet server HTTP port.
func (ws *WalletServer) Port() uint16 {
	return ws.port
}

// Gateway returns the blockchain node HTTP base URL.
func (ws *WalletServer) Gateway() string {
	return ws.gateway
}
