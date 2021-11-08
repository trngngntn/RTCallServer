package internal

import (
	"log"
	"time"
)

type Notification struct {
	id        int
	Uid       string
	Timestamp time.Time
	Data      map[string]interface{}
	Status    int
}

func FetchAll(uid string) (notiList []Notification) {
	db := CreateConnection()
	sql := `
		SELECT * FROM 'Notifications'
		WHERE 'uid' = ? AND 'status' >= 0
	`
	row, err := db.Query(sql, uid, 0)
	if err != nil {
		log.Fatal(err)
	}

	for row.Next() {
		noti := Notification{}
		var byteData []byte
		row.Scan(&noti.id, &noti.Uid, &noti.Timestamp, &byteData, &noti.Status)
	}

	return
}

func FetchUnread(uid string) (notiList []Notification) {
	db := CreateConnection()
	sql := `
		SELECT * FROM 'Notifications'
		WHERE 'uid' = ? AND 'status' = 1
	`
	row, err := db.Query(sql, uid, 0)
	if err != nil {
		log.Fatal(err)
	}

	for row.Next() {
		noti := Notification{}
		var byteData []byte
		row.Scan(&noti.id, &noti.Uid, &noti.Timestamp, &byteData, &noti.Status)
	}

	return
}

func Push(noti *Notification) {
	/*db := CreateConnection()
	sql := `
		INSERT INTO 'Notifications'
		VALUES(?,?,?,?,?)`*/

}

func MarkAsRead(id int) {

}

func Hide(id int) {

}
