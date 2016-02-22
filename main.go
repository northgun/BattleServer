package main

import (
	"crypto/tls"
	"encoding/json"
	glog "github.com/golang/glog"
	"io/ioutil"
)

var (
	Cfg *Config
	S   *Server
	p   *MMpool
	rm  *RoomsManager
)

func main() {
	glog.V(3).Infof("Start gateway server")
	config_path := "gate_config.json"
	file, e := ioutil.ReadFile(config_path)
	glog.Info("Config Path:", config_path)
	if e != nil {
		panic(e)
	}
	json.Unmarshal(file, &Cfg)
	certs := make([]tls.Certificate, 0)
	if Cfg.Certfile != "" && Cfg.Keyfile != "" {
		glog.Info("TLS Cert:", Cfg.Certfile)
		glog.Info("TLS  Key:", Cfg.Keyfile)
		cert, err := tls.LoadX509KeyPair(Cfg.Certfile, Cfg.Keyfile)
		if err != nil {
			panic(err)
		}
		certs = append(certs, cert)
	}
	glog.Info(Cfg.Port)
	rm = NewRoomsManager()
	S = NewServer(
		Cfg.Port,
		certs,
	)

	S.ForeverServe()

	//p = ConnectionPool()
	/*redisclient := p.GetConnection()

	resp := redisclient.Cmd("PING")
	glog.Info(resp)*/

}
