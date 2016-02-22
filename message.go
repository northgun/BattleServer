package main

import (
	"encoding/json"
	//"strings"
	"time"
)

// Command
const (
	NORMAL     = "NORMAL"
	PRIVATEMSG = "PRIVATEMSG"
	DISCONNECT = "DISCONNECT"
	QUIT       = "QUIT"
	JOIN       = "JOIN"
	KICK       = "KICK"
	CANCEL     = "CANCEL"
	IDENTITY   = "IDENTITY"
	DISCOVERY  = "DISCOVERY"
	LIST       = "LIST"
	PING       = "PING"
	HISTORY    = "HISTORY"
)

type ClientResponse struct {
	TimeStamp string `json:"timestamp"`
	Command   string `json:"cmd"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
}

type Message [6]string

func NewMessage(cmd string, sender_name string, target string, content string) []string {

	msg := make([]string, 6, 6)
	msg[0] = time.Now().UTC().Format(time.RFC3339)
	msg[1] = cmd
	msg[2] = sender_name // sender name (Client Name@Server Endpoint)
	msg[3] = target      // target (Client Name@[Server Endpoint] / Room Name)
	msg[4] = content
	msg[5] = "\r\n"
	return msg
}

func NewClientResponse(command string, from string, to string, msg string) *ClientResponse {
	return &ClientResponse{time.Now().UTC().Format(time.RFC3339), command, from, to, msg}
}

func ClientResponseToByte(msg ClientResponse) []byte {
	b, _ := json.Marshal(msg)
	return b
}
