package main

import (
	"bufio"
	"encoding/binary"
	//"encoding/json"
	//"fmt"
	//zerorpc "github.com/bsphere/zerorpc"
	//goreq "github.com/franela/goreq"
	glog "github.com/golang/glog"
	"github.com/mediocregopher/radix.v2/redis"
	"net"
	//"net/url"
	"strings"
	"sync"
	"time"
)

var (
	mmclient *redis.Client
)

type Client struct {
	Name     string
	Conn     net.Conn
	Init     bool
	In       chan ClientResponse
	Raw      chan []byte
	Error    chan ClientResponse
	RoomList map[string]bool
	Closed   bool
	Quit     chan bool
	QuitRaw  chan bool
	sync.Mutex
}

func NewClient(conn net.Conn) (c *Client) {

	var mu sync.Mutex
	c = &Client{
		conn.RemoteAddr().String(),
		conn,
		false,
		make(chan ClientResponse),
		make(chan []byte, 16),
		make(chan ClientResponse),
		make(map[string]bool, 2),
		false,
		make(chan bool),
		make(chan bool),
		mu,
	}

	glog.Info("Begin to get data from Client")
	go c.RecvFromConn()
	go c.JoinAndListen()
	go c.ListenRaw()
	return
}

func (c *Client) RecvFromConn() {

	if glog.V(2) {
		glog.Infof("Client RecvFromConn: %s", c.Name)
	}
	glog.Info("Client RecvFromConn: ", c.Name)
	scanner := bufio.NewScanner(c.Conn)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		c.ParseAndSend(scanner.Bytes())
	}

	// Close Client
	c.Closed = true
	if err := scanner.Err(); err != nil {
		glog.Error("Error on %s", err)
	}
	if err := c.Conn.Close(); err != nil {
		glog.Warningf("%s Connection closed", err)
	}
	glog.Infof("%s Connection closed", c.Conn.RemoteAddr())

}

func (c *Client) ParseAndSend(line []byte) {

	// c t c
	// Command/TargetName/Content
	ctc := strings.SplitN(string(line), " ", 3)
	glog.Info(string(line))
	if len(ctc) != 3 {
		glog.Warningf("Not enough for client ParseAndSend: %s", ctc)
		cr := NewClientResponse("NONE", "SYSTEM", "", "711")
		c.Error <- *cr
		return
	}

	data := NewMessage(ctc[0], c.Name, ctc[1], ctc[2])
	glog.V(3).Info("Client Data:", data)
	cmd := data[1]
	glog.Info("Command:", ctc[0], cmd, ctc[2])
	if cmd != IDENTITY {
		if !c.Init {
			cr := NewClientResponse("NONE", "SYSTEM", "", "712")
			c.Error <- *cr
			return
		}
		switch cmd {
		case JOIN:
			// Check Joinable
			glog.V(3).Info("JOIN Message", data)
			c.In <- *NewClientResponse("JOIN", "SYSTEM", ctc[1], "200")
			glog.V(3).Info("JOIN Message End", data)
			rm.RecvChan <- data
		case NORMAL, QUIT:
			glog.V(8).Info("NORMAL Message", data)
			rm.RecvChan <- data
		case PING:
			c.In <- *NewClientResponse("PING", "SYSTEM", c.Name, ctc[2])
		}
	} else {
		/*
			params := url.Values{}
			params.Set("device_id", data[3])
			params.Set("session_id", data[4])

			res, err := goreq.Request{
				Uri:         Cfg.LoginEndpoint,
				QueryString: params}.Do()

			if err != nil || res.StatusCode != 200 {
				glog.V(3).Info("err", err, `status code`, res.StatusCode)
				cr := NewClientResponse("IDENTITY", "SYSTEM", "", "700")
				c.Error <- *cr
				return
			}

			glog.Info("Response,", res)
			var csr CheckSessionResponse
			res.Body.FromJsonTo(&csr)
			defer res.Body.Close()

			// CGS get nickname...
			tmp_server_name := strings.Replace(csr.ServerName, `\`, "", -1)
			var server_groups []string
			json.Unmarshal([]byte(tmp_server_name), &server_groups)

			fmt.Println(tmp_server_name)
			if len(server_groups) == 0 {
				cr := NewClientResponse("IDENTITY", "SYSTEM", "", "701")
				c.Error <- *cr
				return
			}

			cgs, _ := zerorpc.NewClient(Cfg.CGSEndpoint)
			rsp, err := cgs.Invoke("distribute", `{"messageType":"account.get_player_nickname"}`, fmt.Sprintf("%s:%s", server_groups[0], csr.PlayerID))
			if err != nil {
				cr := NewClientResponse("IDENTITY", "SYSTEM", "", "702")
				c.Error <- *cr
				return
			}

			if len(rsp.Args) < 1 || len(rsp.Args[0].([]interface{})) < 1 {
				cr := NewClientResponse("IDENTITY", "SYSTEM", "", "702")
				c.Error <- *cr
				return
			}

			args := fmt.Sprintf("%v", rsp.Args[0].([]interface{})[0])
			fmt.Println("Get the args", args)

			var cgs_rsp CGSResponse
			json.Unmarshal([]byte(args), &cgs_rsp)

			nick_name := cgs_rsp.Data["nickname"]

			if nick_name == "" {
				c.Error <- *NewClientResponse("IDENTITY", "SYSTEM", "", "703")
				return
			}
		c.Name = server_groups[0] + ":" + csr.PlayerID*/
		c.Name = data[3]
		c.Init = true
		rm.NewClientChan <- c
		// Init join
		c.In <- *NewClientResponse("IDENTITY", "SYSTEM", "", "200")
	}
}

type CGSResponse struct {
	Data  map[string]string `json:"return_data"`
	Extra map[string]string `json:"extra_data"`
	Code  int               `json:"return_code"`
}

type CheckSessionResponse struct {
	PlayerID   string `json:"uuid"`
	ServerName string `json:"sname"`
}

func (c *Client) ListenRaw() {
	for {
		select {
		case raw := <-c.Raw:
			c.SendData(raw)
		case <-c.QuitRaw:
			return
		}

	}

}

func (c *Client) JoinAndListen() {

	glog.Infof("%s JoinAndListen", c.Name)

	// Clean all left
	defer close(c.In)
	defer close(c.Error)

	for {
		if !c.Closed {
			select {
			case cr := <-c.In:
				glog.V(8).Info("Client:%s In %s", c.Name, cr)
				c.SendData(ClientResponseToByte(cr))
			case cr := <-c.Error:
				glog.V(8).Info("Client:%s Error %s", c.Name, cr)
				c.SendData(ClientResponseToByte(cr))
			case <-c.Quit:
				return
			}
		} else {
			<-c.Quit
			glog.Infof("%s Closed", c.Name)
			return
		}
	}
}

func (c *Client) SendData(data []byte) {

	c.Lock()
	defer c.Unlock()

	glog.V(8).Infof("%s Raw Data:%s", c.Name, data)

	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// All data should through here
	err := binary.Write(c.Conn, binary.LittleEndian, int32(len(data)))
	c.handleSendError(err)

	err = binary.Write(c.Conn, binary.LittleEndian, data)
	c.handleSendError(err)

	//Test the Redis Client
	/*redisclient := p.GetConnection()
	resp := redisclient.Cmd("PING")
	glog.Info(resp)*/
}

func (c *Client) handleSendError(err error) {

	if err != nil && !c.Closed {
		glog.Error(err)
		c.Closed = true
		//if c.Name != "" {
		//	H.QuitClientChan <- c
		//}
	}
}
