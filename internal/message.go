package internal

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

var Calls = make(map[string]string)

const (
	msgClientLogin uint32 = 0x01 // receive:
	msgClientDial  uint32 = 0x02 // receive: dial request from client

	/////////////////////////////////////////////////////////////////////////////////////////////////

	// msgRelayCallAccepted uint32 = 0x20
	// msgRelayCallDeclined uint32 = 0x21
	// msgRelayCallEnded    uint32 = 0x22

	// msgRelayWebRTCCandidate uint32 = 0x30
	// msgRelayWebRTCOffer     uint32 = 0x31
	// msgRelayWebRTCAnswer    uint32 = 0x32

	/////////////////////////////////////////////////////////////////////////////////////////////////

	msgServerBadIdentity uint32 = 0x00
	msgServerLoggedIn    uint32 = 0x01
	msgServerRequestCall uint32 = 0x10
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
	log.Printf("Mess: %s", string(dataByte[4:]))
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
	//log.Printf("Message: type=%x,length=%s\n\n\n\n", msg.typ, msg.jsonData)
	switch {
	//message login data contain uid
	case msg.typ == msgClientLogin:
		var uid = msg.jsonData["uid"].(string)
		MapClient[uid] = &Client{SocketConn: conn}
		MapAddr[conn.RemoteAddr()] = uid
		log.Printf("Registered UID: %s", uid)
		var sendMsg = NetMessage{typ: msgServerLoggedIn, jsonData: make(map[string]interface{})}
		sendMsg.jsonData["uid"] = uid
		go sendMessage(conn, &sendMsg)

	case msg.typ == msgClientDial:
		var fromClientUID = MapAddr[conn.RemoteAddr()]
		var toClientUID = msg.jsonData["uid"].(string)
		var toClient = MapClient[toClientUID]
		var sendMsg = NetMessage{typ: msgServerRequestCall, jsonData: make(map[string]interface{})}
		sendMsg.jsonData["caller"] = fromClientUID

		Calls[fromClientUID] = toClientUID
		Calls[toClientUID] = fromClientUID

		//log.Printf("Call between %s and %s", fromClientUID, toClient.UID)

		go sendMessage(toClient.SocketConn, &sendMsg)

	case msg.typ > 0x1F:
		var fromClientUID = MapAddr[conn.RemoteAddr()]
		var toClientUID = Calls[fromClientUID]
		go sendMessage(MapClient[toClientUID].SocketConn, msg)
	}
}

func sendMessage(conn net.Conn, msg *NetMessage) {
	conn.Write(convertMessageToByteArray(msg))
	log.Printf("Sent message to %s\n", conn.RemoteAddr().String())
}
