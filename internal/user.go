package internal

import "net"

type User struct {
	id   string
	name string
	addr string
}

var MapClient = make(map[string]*Client)
var MapAddr = make(map[net.Addr]string)

type Client struct {
	socketConn   net.Conn
	DatagramAddr string
}
