package internal

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

var Calls = make(map[string]string)

const (
	msgC2SLogin uint32 = 0x02 // receive:
	msgC2SAck   uint32 = 0x03 // receive:

	msgC2SDial uint32 = 0x04 // receive: dial request from client
	msgC2SPing uint32 = 0x05 // receive(UDP): ping for getting UDP addr

	msgC2SAcceptCall  uint32 = 0x1A // receive(UDP): accepted call signal
	msgC2SDeclineCall uint32 = 0x1B // receive: rejected call signal

	msgC2SEndCall uint32 = 0x1C // receive

	msgS2CBadIdentity uint32 = 0x00
	msgS2CLoggedIn    uint32 = 0x01

	msgS2CRequestCall  uint32 = 0x0B // send: call request
	msgS2CCallDeclined uint32 = 0x0C // send: call request from client is rejected by other client
	msgS2CCallEnded    uint32 = 0x0C // send: call request from client is rejected by other client
	msgS2CPeerAddr     uint32 = 0x02 // send: UDP addr to other client
)

type NetMessage struct {
	typ uint32
	//data     string
	jsonData map[string]interface{}
}

func NewMessage(typ uint32, jsonData map[string]interface{}) NetMessage {
	return NetMessage{typ: typ, jsonData: jsonData}
}

func ParseMessage(dataByte []byte) *NetMessage {
	var result = NetMessage{}

	result.typ = binary.BigEndian.Uint32(dataByte[:4])

	json.Unmarshal(dataByte[4:], &result.jsonData)

	return &result
}
func convertMessageToByteArray(msg *NetMessage) []byte {
	byteData, _ := json.Marshal(msg.jsonData)

	var byteHead = make([]byte, 8)

	binary.BigEndian.PutUint32(byteHead, uint32(len(byteData)))
	binary.BigEndian.PutUint32(byteHead[4:], msg.typ)

	//log.Println(byteHead)

	return append(byteHead, byteData...)
}

func ProcessMessage(msg *NetMessage, conn net.Conn) {
	//log.Printf("Processing message type: %d", msg.typ)
	switch msg.typ {
	//message login data contain uid
	case msgC2SLogin:
		var uid = msg.jsonData["uid"].(string)
		MapClient[uid] = &Client{socketConn: conn}
		MapAddr[conn.RemoteAddr()] = uid
		log.Printf("Registered UID: %s", uid)
		var sendMsg = NetMessage{typ: msgS2CLoggedIn, jsonData: make(map[string]interface{})}
		sendMsg.jsonData["uid"] = uid
		go sendMessage(conn, &sendMsg)

	case msgC2SDial:
		var sentClientUID = MapAddr[conn.RemoteAddr()]
		var recvClient = MapClient[msg.jsonData["uid"].(string)]
		var sendMsg = NetMessage{typ: msgS2CRequestCall, jsonData: make(map[string]interface{})}
		sendMsg.jsonData["caller"] = sentClientUID
		go sendMessage(recvClient.socketConn, &sendMsg)
	}
}

func ProcessUDPMessage(msg *NetMessage, addr string) {
	var recvClientUID = msg.jsonData["otherPeer"].(string)
	var recvConn = MapClient[recvClientUID].socketConn
	var sendMsg = NetMessage{typ: msgS2CPeerAddr, jsonData: make(map[string]interface{})}
	sendMsg.jsonData["peerAddr"] = addr
	go sendMessage(recvConn, &sendMsg)
}

func sendMessage(conn net.Conn, msg *NetMessage) {
	conn.Write(convertMessageToByteArray(msg))
	log.Printf("Sent message to %s\n", conn.LocalAddr().String())
}
