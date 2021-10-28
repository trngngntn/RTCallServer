package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
	reader := bufio.NewReader(conn)
	for {
		byteData := make([]byte, 4)
		_, err := io.ReadFull(reader, byteData)
		if err != nil {
			log.Println("Connection closed")
			conn.Close()
			return
		}

		size := binary.BigEndian.Uint32(byteData)

		log.Printf("Message: size=%d", size)

		byteData = make([]byte, size+4)
		n, err := io.ReadFull(reader, byteData)

		if err != nil {
			log.Println("Connection closed")
			conn.Close()
			return
		}

		if n != int(size)+4 {
			log.Panicf("Invalid size, read: %d byte ", n)
			conn.Close()
			os.Exit(-1)
		}

		if size > 5000 {
			os.Exit(-1)
		}
		go internal.ProcessMessage(byteData, conn)
		//log.Println(string(buffer))
	}
}
