package main

import (
	"encoding/json"
	"fmt"
	glog "github.com/golang/glog"
	"time"
)

type RoomsManager struct {
	RecvChan       chan []string
	NewClientChan  chan *Client
	QuitClientChan chan *Client
	Clients        map[string]*Client
	// Rooms control
	Rooms map[string]*Room
}

type Room struct {
	Name    string
	Clients map[string]*Client
	Creator string
}

const (
	ROOMCAPACITY = 4
	ROOMNAME     = "TEST"
)

func NewRoomsManager() *RoomsManager {
	rm := &RoomsManager{
		make(chan []string, 1024),
		make(chan *Client, 256),
		make(chan *Client, 256),
		make(map[string]*Client),
		make(map[string]*Room),
	}
	go rm.Serve()
	return rm
}

func (rm *RoomsManager) handleClientMessage(msg []string) {
	glog.V(4).Info("[Handle Internal]", msg)
	cr := NewClientResponse(msg[1], msg[2], msg[3], msg[4])

	switch cr.Command {
	/*case PRIVATEMSG:
	if c, ok := h.Clients[cr.To]; ok {
		// Clients on the same server
		c.In <- *cr
	} else {
		S.ClusterSend <- MessageToBytes(msg)
	}
	return
	*/
	case JOIN:
		glog.Info("JOIN:", cr.Content)
		rm.joinRoom(cr.From, cr.To, cr.Content)
	case QUIT:
		rm.quitRoom(cr.From, cr.To)
	case NORMAL:
		rm.broadcast(cr.To, cr)
	}

}

func (rm *RoomsManager) newRoom(creator string, room_name string) {

	room := &Room{room_name, make(map[string]*Client, 0), creator}
	rm.Rooms[room_name] = room

}

func (rm *RoomsManager) joinRoom(client_name string, room_name string, content string) error {
	if client, ok := rm.Clients[client_name]; ok {
		glog.V(2).Infof("JoinRoom:client:%v", client)
		if room, ok := rm.Rooms[room_name]; ok {
			room.Clients[client_name] = client
			client.RoomList[room_name] = true
		} else {
			rm.newRoom(client_name, room_name)
			rm.Rooms[room_name].Clients[client_name] = client
			client.RoomList[room_name] = true
		}

		return nil
	}
	return fmt.Errorf("%s not existed", client_name)
}

func (rm *RoomsManager) quitRoom(client_name string, room_name string) error {
	// XXX N-M problem
	glog.V(2).Infof("%s quit %v", client_name, room_name)

	if room, ok := rm.Rooms[room_name]; ok {
		if c, c_ok := room.Clients[client_name]; c_ok {
			delete(c.RoomList, room_name)
			delete(room.Clients, client_name)
		} else {
			return fmt.Errorf("%s not in %s", c.Name, room_name)
		}
		return nil
	} else {
		return fmt.Errorf("%s not existed", room_name)
	}
}

func (rm *RoomsManager) destoryRoom(name string) {

	if room, ok := rm.Rooms[name]; ok {
		if len(room.Clients) == 0 {
			delete(rm.Rooms, name)
		}
	}
}

func (rm *RoomsManager) broadcast(room_name string, cr *ClientResponse) {

	glog.V(8).Infof("Broadcasting @ %s with %s", room_name, cr)
	if room, ok := rm.Rooms[room_name]; ok {
		// convert to bytes to save Marshal
		b, _ := json.Marshal(cr)
		for _, c := range room.Clients {
			glog.Info("Broadcasting send to %s", c.Name)
			glog.V(8).Infof("Broadcasting send to %s", c.Name)
			c.Raw <- b
		}
	}
}

func (rm *RoomsManager) Serve() {
	for {
		select {
		case msg := <-rm.RecvChan:
			rm.handleClientMessage(msg)
		case c := <-rm.NewClientChan:
			rm.Clients[c.Name] = c
		case c := <-rm.QuitClientChan:
			glog.V(4).Infof("Client:%s QUIT, rooms:%s", c.Name, c.RoomList)
			for k, _ := range c.RoomList {
				rm.quitRoom(c.Name, k)
			}
			delete(rm.Clients, c.Name)
			// All done
			c.Quit <- true
			c.QuitRaw <- true
		case <-time.After(time.Second * 5):
			glog.V(8).Info("No Message")
		}

	}
}
