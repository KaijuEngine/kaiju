package network

import (
	"kaiju/engine"
	"kaiju/klib"
	"net"
	"sync"
	"time"
)

const reliableRetryDelay = time.Millisecond * 15

type PendingNetworkPacketUDP struct {
	target *ServerClient
	packet NetworkPacketUDP
}

type NetworkUDP struct {
	conn           *net.UDPConn
	pendingPackets []PendingNetworkPacketUDP
	pendingMutex   sync.RWMutex
	updateId       int
	isReading      bool
}

func (n *NetworkUDP) IsLive() bool { return n.conn != nil }

func (n *NetworkUDP) Close(updater *engine.Updater) {
	n.isReading = false
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	updater.RemoveUpdate(n.updateId)
	n.updateId = 0
}

func (n *NetworkUDP) removePendingPacket(id int64) {
	n.pendingMutex.Lock()
	defer n.pendingMutex.Unlock()
	for i := range n.pendingPackets {
		if n.pendingPackets[i].packet.timestamp == id {
			n.pendingPackets = klib.RemoveUnordered(n.pendingPackets, i)
			break
		}
	}
}

func (n *NetworkUDP) createUnreliable(message []byte) NetworkPacketUDP {
	packet := NetworkPacketUDP{
		timestamp:  time.Now().UTC().UnixMicro(),
		messageLen: uint16(len(message)),
	}
	copy(packet.message[:], message)
	return packet
}

func (n *NetworkUDP) createReliable(message []byte, target *ServerClient) NetworkPacketUDP {
	packet := NetworkPacketUDP{
		timestamp:  time.Now().UTC().UnixMicro(),
		order:      target.reliableOrder,
		messageLen: uint16(len(message)),
		typeFlags:  udpPacketTypeReliable,
		nextRetry:  time.Now().Add(reliableRetryDelay),
	}
	target.reliableOrder++
	copy(packet.message[:], message)
	n.pendingMutex.Lock()
	n.pendingPackets = append(n.pendingPackets, PendingNetworkPacketUDP{
		target: target,
		packet: packet,
	})
	n.pendingMutex.Unlock()
	return packet
}

func (n *NetworkUDP) createAck(fromTimestamp []byte) NetworkPacketUDP {
	packet := NetworkPacketUDP{
		timestamp:  time.Now().UTC().UnixMicro(),
		messageLen: uint16(len(fromTimestamp)),
		typeFlags:  udpPacketTypeAck,
	}
	copy(packet.message[:], fromTimestamp)
	return packet
}
