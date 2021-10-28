package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"trngngntn/rcserver/internal"
)

const (
	addr       = ":42069"
	bufferSize = 1024
)

var m = make(map[string]*net.Conn)

func main() {
	fmt.Println("Starting server...")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("TCP error!")
		return
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error!")
			continue
		}
		log.Println("Client" + conn.RemoteAddr().String() + "connected")

		go handleServiceConnection(conn)
	}

}

func exit() {
	fmt.Println("Exiting")
}

func handleServiceConnection(conn net.Conn) {
	for {
		byteSize := make([]byte, 4)
		_, err := conn.Read(byteSize)

		if err != nil {
			log.Println("Connection closed")
			conn.Close()
			return
		}

		byte := make([]byte, 4+binary.BigEndian.Uint32(byteSize))
		_, err = conn.Read(byte)

		msg := internal.ParseMessage(byte)

		go internal.ProcessMessage(msg, conn)
		//log.Println(string(buffer))
	}
}
