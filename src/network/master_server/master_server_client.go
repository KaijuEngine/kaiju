package master_server

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/network"
	"log/slog"
	"unsafe"
)

const (
	pingIntervalSeconds = 3.0
)

type MasterServerClient struct {
	client       network.NetworkClient
	pingTime     float64
	updateId     int
	isServer     bool
	OnServerList func([]ResponseServerList, uint32)
	OnServerJoin func(string)
	OnClientJoin func(string)
	OnError      func(uint8)
}

func (c *MasterServerClient) Connect(updater *engine.Updater) error {
	if c.updateId != 0 {
		c.Disconnect(updater)
	}
	c.client = network.NewClientUDP()
	if c.OnServerList == nil {
		c.OnServerList = func([]ResponseServerList, uint32) {}
	}
	if c.OnServerJoin == nil {
		c.OnServerJoin = func(s string) {}
	}
	if c.OnClientJoin == nil {
		c.OnClientJoin = func(s string) {}
	}
	if c.OnError == nil {
		c.OnError = func(u uint8) {}
	}
	err := c.client.Connect(updater, masterAddress, masterPort)
	if err != nil {
		slog.Error("failed to setup connection for master server", "address", masterAddress, "port", masterPort)
	} else {
		debug.Log("Successfully bound master server client")
	}
	c.updateId = updater.AddUpdate(c.update)
	return err
}

func (c *MasterServerClient) Disconnect(updater *engine.Updater) {
	debug.Log("Disconnecting the master server client")
	updater.RemoveUpdate(c.updateId)
	c.updateId = 0
	c.client.Close(updater)
}

func (c *MasterServerClient) RegisterServer(game, name string, maxPlayers, currentPlayers uint16) error {
	debug.Log("Registering server with master server")
	req := Request{
		Type:           RequestTypeRegisterServer,
		MaxPlayers:     maxPlayers,
		CurrentPlayers: currentPlayers,
	}
	copy(req.Game[:], game)
	copy(req.Name[:], name)
	c.isServer = true
	return c.sendRequest(req)
}

func (c *MasterServerClient) ListServers(game string) error {
	req := Request{Type: RequestTypeServerList}
	copy(req.Game[:], game)
	return c.sendRequest(req)
}

func (c *MasterServerClient) JoinServer(id uint64) error {
	return c.sendRequest(Request{Type: RequestTypeJoinServer, ServerId: id})
}

func (c *MasterServerClient) update(deltaTime float64) {
	if c.isServer {
		c.pingTime -= deltaTime
		if c.pingTime < 0 {
			c.pingTime = pingIntervalSeconds
			c.sendRequest(Request{Type: RequestTypePing})
		}
	}
	messages := c.client.ServerMessageQueue.Flush()
	for i := range messages {
		buff := messages[i].Message()
		if len(buff) != int(unsafe.Sizeof(Response{})) {
			continue
		}
		c.processMessage(messages[i])
	}
}

func (c *MasterServerClient) processMessage(msg network.ClientMessage) {
	res := DeserializeResponse(msg.Message())
	switch res.Type {
	case ResponseTypeConfirmRegister:
		debug.Log("<- Confirm register")
	case ResponseTypeServerList:
		debug.Log("<- Server list")
		c.OnServerList(res.List[:], res.TotalList)
	case ResponseTypeJoinServerInfo:
		debug.Log("<- Join server")
		c.OnServerJoin(klib.ByteArrayToString(res.Address[:]))
	case ResponseTypeClientJoinInfo:
		debug.Log("<- Client join")
		c.OnClientJoin(klib.ByteArrayToString(res.Address[:]))
	case ResponseTypeError:
		debug.Log("<- Error", "error", res.Error)
		c.OnError(res.Error)
	}
}

func (c *MasterServerClient) sendRequest(req Request) error {
	switch req.Type {
	case RequestTypeRegisterServer:
		debug.Log("-> Register")
	case RequestTypeUnregisterServer:
		debug.Log("-> Unregister")
	case RequestTypePing:
		debug.Log("-> Ping")
	case RequestTypeServerList:
		debug.Log("-> Server list")
	case RequestTypeJoinServer:
		debug.Log("-> Connect to server")
	}
	buff := [unsafe.Sizeof(Request{})]byte{}
	req.Serialize(buff[:])
	err := c.client.SendMessageReliable(buff[:])
	if err != nil {
		slog.Error("failed to send the message to the master server")
	}
	return err
}
