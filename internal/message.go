package internal

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"time"
)

var Calls = make(map[string]string)

const (
	msgClientLogin          uint32 = 0x01 // receive:
	msgClientRegister       uint32 = 0x02 // receive:
	msgClientDial           uint32 = 0x03 // receive: dial request from client
	msgClientRequestContact uint32 = 0x04
	msgClientAddContact     uint32 = 0x05
	msgClientApproveContact uint32 = 0x06
	msgClientRejectContact  uint32 = 0x07

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

	msgServerRegistered     uint32 = 0x02
	msgServerRegisterFailed uint32 = 0x03

	msgServerContactList     uint32 = 0x04
	msgServerContactInvalid  uint32 = 0x05
	msgServerContactPending  uint32 = 0x06
	msgServerContactApproved uint32 = 0x07

	msgServerAllNotif    uint32 = 0x08
	msgServerUnreadNotif uint32 = 0x09
	msgServerNewNotif    uint32 = 0x10

	msgServerRequestCall uint32 = 0x11

	notifContactRequest uint32 = 0x01
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

	if result.typ > 0x1F {
		return nil
	}

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

func ProcessMessage(byteData []byte, conn net.Conn) {

	msg := ParseMessage(byteData)

	if msg == nil {
		go forwardMessage(byteData, conn)
		return
	}

	switch {
	//message login data contain uid
	case msg.typ == msgClientLogin:
		username := msg.jsonData["username"].(string)
		password := msg.jsonData["password"].(string)
		if Login(username, password) {
			MapClient[username] = &Client{SocketConn: conn}
			MapAddr[conn.RemoteAddr()] = username
			log.Printf("User login: %s", username)
			var sendMsg = NetMessage{typ: msgServerLoggedIn, jsonData: make(map[string]interface{})}
			sendMsg.jsonData["uid"] = username
			go sendMessage(conn, &sendMsg)
		} else {
			var sendMsg = NetMessage{typ: msgServerBadIdentity, jsonData: make(map[string]interface{})}
			go sendMessage(conn, &sendMsg)
		}

	case msg.typ == msgClientRegister:
		var username = msg.jsonData["username"].(string)
		if UsernameExists(username) {
			sendMsg := NetMessage{typ: msgServerRegisterFailed, jsonData: make(map[string]interface{})}
			go sendMessage(conn, &sendMsg)
		} else {
			var displayName = msg.jsonData["displayName"].(string)
			var password = msg.jsonData["password"].(string)
			CreateNewUser(username, password, displayName)
			sendMsg := NetMessage{typ: msgServerRegistered, jsonData: make(map[string]interface{})}
			go sendMessage(conn, &sendMsg)
		}

	case msg.typ == msgClientRequestContact:
		var fromClientUID = MapAddr[conn.RemoteAddr()]
		contactList := GetContactList(fromClientUID)
		sendMsg := NetMessage{typ: msgServerContactList, jsonData: make(map[string]interface{})}
		sendMsg.jsonData["contactList"] = contactList
		go sendMessage(conn, &sendMsg)

	case msg.typ == msgClientDial:
		var fromClientUID = MapAddr[conn.RemoteAddr()]
		var toClientUID = msg.jsonData["uid"].(string)
		var toClient = MapClient[toClientUID]
		var sendMsg = NetMessage{typ: msgServerRequestCall, jsonData: make(map[string]interface{})}
		sendMsg.jsonData["caller"] = fromClientUID

		Calls[fromClientUID] = toClientUID
		Calls[toClientUID] = fromClientUID

		log.Printf("Call between %s and %s", fromClientUID, toClientUID)

		go sendMessage(toClient.SocketConn, &sendMsg)

	case msg.typ == msgClientAddContact:
		var fromClientUID = MapAddr[conn.RemoteAddr()]
		var contactUsername = msg.jsonData["username"].(string)
		if UsernameExists(contactUsername) {
			sendMsg := NetMessage{typ: msgServerContactPending, jsonData: make(map[string]interface{})}
			go sendMessage(conn, &sendMsg)
			var noti = Notification{Uid: contactUsername, Timestamp: time.Now(), Data: make(map[string]interface{}), Status: 0}
			noti.Data["type"] = notifContactRequest
			noti.Data["fromUid"] = fromClientUID
			AddPendingContact(fromClientUID, contactUsername)
			notify(contactUsername)
		} else {
			sendMsg := NetMessage{typ: msgServerContactInvalid, jsonData: make(map[string]interface{})}
			go sendMessage(conn, &sendMsg)
		}

	case msg.typ == msgClientApproveContact:
		fromClientUID := MapAddr[conn.RemoteAddr()]
		toClientUID := msg.jsonData["uid"].(string)
		notifId := msg.jsonData["notifId"].(int)
		ApproveContact(fromClientUID, toClientUID)
		Hide(notifId)

	case msg.typ == msgClientRejectContact:
		fromClientUID := MapAddr[conn.RemoteAddr()]
		toClientUID := msg.jsonData["uid"].(string)
		notifId := msg.jsonData["notifId"].(int)
		RejectContact(fromClientUID, toClientUID)
		Hide(notifId)

	}

}

func forwardMessage(byteData []byte, conn net.Conn) {
	var fromClientUID = MapAddr[conn.RemoteAddr()]
	var toClientUID = Calls[fromClientUID]
	var byteHead = make([]byte, 4)
	binary.BigEndian.PutUint32(byteHead, uint32(len(byteData)-4))

	//log.Printf("\n%s\n\n", string(append(byteHead, byteData...)))
	go MapClient[toClientUID].SocketConn.Write(append(byteHead, byteData...))
}

func notify(uid string) {
	if MapClient["uid"] != nil {

	}
}

func sendMessage(conn net.Conn, msg *NetMessage) {
	conn.Write(convertMessageToByteArray(msg))
	log.Printf("Sent message to %s\n", conn.RemoteAddr().String())
}
