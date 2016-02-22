package main

import (
	"crypto/tls"
	"fmt"
	glog "github.com/golang/glog"
	"net"
)

type Server struct {
	Conn net.Listener
}

func NewServer(addr string, certs []tls.Certificate) *Server {
	var conn net.Listener
	var err error

	if len(certs) != 0 {
		tls_config := &tls.Config{Certificates: certs}
		conn, err = tls.Listen("tcp", addr, tls_config)
		if err != nil {
			glog.Fatalf("Listen TLS Fatal:", err)
		}
	} else {
		conn, err = net.Listen("tcp", addr)
		if err != nil {
			glog.Fatalf("Listen Raw Fatal:", err)
		}
	}
	fmt.Println(conn)
	glog.Info()
	return &Server{conn}
}

func (srv *Server) ListenClientConn() {
	defer srv.Conn.Close()
	for {
		conn, err := srv.Conn.Accept()
		if err != nil {
			glog.Warning("conn: failed:", conn, err)
			continue
		}
		go func(c net.Conn) {
			// New client
			// We need names for it
			glog.Info("New Connection From: ", c.RemoteAddr())
			NewClient(c)
		}(conn)
	}
}

func (srv *Server) ForeverServe() {
	srv.ListenClientConn()
}
