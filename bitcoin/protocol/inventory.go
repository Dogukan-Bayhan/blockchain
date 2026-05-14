package protocol

// InventoryType identifies the kind of object announced by an inventory item.
type InventoryType string

const (
	InventoryTx    InventoryType = "tx"
	InventoryBlock InventoryType = "block"
)

// InventoryItem announces one transaction or block by hash.
type InventoryItem struct {
	Type InventoryType `json:"type"`
	Hash string        `json:"hash"`
}

// InventoryPayload carries hashes a peer may request with getdata.
type InventoryPayload struct {
	Items []InventoryItem `json:"items"`
}

// GetDataPayload asks a peer to send full objects for announced inventory items.
type GetDataPayload struct {
	Items []InventoryItem `json:"items"`
}
