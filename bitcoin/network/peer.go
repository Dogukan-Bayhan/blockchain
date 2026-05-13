package network

import (
	"bitcoin/protocol"
	"encoding/json"
	"log"
	"net"
)

func HandleConn(conn net.Conn, outbound bool, nodeID string, p2pAddr string) {
	defer conn.Close()

	codec := protocol.NewCodec(conn)

	if outbound {
		if err := sendVersion(codec, nodeID, p2pAddr); err != nil {
			log.Printf("network: send version error: %v", err)
			return
		}
		log.Printf("network: sent version to %s", conn.RemoteAddr().String())
	}

	for {
		msg, err := codec.ReadMessage()
		if err != nil {
			log.Printf("network: read error from %s: %v", conn.RemoteAddr().String(), err)
			return
		}

		switch msg.Type {
		case protocol.MessageVersion:
			var payload protocol.VersionPayload
			if err := json.Unmarshal(msg.Data, &payload); err != nil {
				log.Printf("network: invalid version payload: %v", err)
				return
			}

			log.Printf("network: received version from %s at %s", payload.NodeID, payload.P2PAddr)

			if err := codec.WriteMessage(protocol.Message{Type: protocol.MessageVerAck}); err != nil {
				log.Printf("network: send verack error: %v", err)
				return
			}
			log.Printf("network: sent verack to %s", payload.NodeID)

		case protocol.MessageVerAck:
			log.Printf("network: received verack from %s", conn.RemoteAddr().String())

		default:
			log.Printf("network: unknown message type %q from %s", msg.Type, conn.RemoteAddr().String())
		}
	}
}

func sendVersion(codec *protocol.Codec, nodeID string, p2pAddr string) error {
	payload := protocol.VersionPayload{
		NodeID:     nodeID,
		P2PAddr:    p2pAddr,
		BestHeight: 0,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return codec.WriteMessage(protocol.Message{
		Type: protocol.MessageVersion,
		Data: data,
	})
}
