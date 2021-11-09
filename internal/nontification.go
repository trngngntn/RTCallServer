package internal

import (
	"encoding/json"
	"log"
	"time"
)

type Notification struct {
	Id        int64                  `json:"id"`
	Uid       string                 `json:"uid"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Status    int                    `json:"status"`
}

func FetchAll(uid string) (notiList []Notification) {
	db := CreateConnection()
	sql := `
		SELECT * FROM "Notifications"
		WHERE "uid" = ? AND "status" >= 0`
	row, err := db.Query(sql, uid, 0)
	if err != nil {
		log.Fatal(err)
	}

	for row.Next() {
		noti := Notification{}
		var data string
		row.Scan(&noti.Id, &noti.Uid, &noti.Timestamp, &data, &noti.Status)
		_ = json.Unmarshal([]byte(data), &noti.Data)
		notiList = append(notiList, noti)
	}

	return
}

func FetchUnread(uid string) (notiList []Notification) {
	db := CreateConnection()
	sql := `
		SELECT * FROM "Notifications"
		WHERE "uid" = ? AND "status" = 1
	`
	row, err := db.Query(sql, uid, 0)
	if err != nil {
		log.Fatal(err)
	}

	for row.Next() {
		noti := Notification{}
		var data string
		_ = row.Scan(&noti.Id, &noti.Uid, &noti.Timestamp, &data, &noti.Status)
		_ = json.Unmarshal([]byte(data), &noti.Data)
		notiList = append(notiList, noti)
	}

	return
}

func Push(noti *Notification) {
	db := CreateConnection()
	sql := `
		INSERT INTO "Notifications"("uid", "timestamp", "data", "status")
		VALUES(?,?,?,0)`
	data, err := json.Marshal(noti.Data)
	if err != nil {
		log.Fatal(err)
	}

	result, err := db.Exec(sql, noti.Uid, noti.Timestamp, string(data))
	if err != nil {
		log.Fatal(err)
	}

	noti.Id, err = result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
}

func MarkAsRead(id int) {
	db := CreateConnection()
	sql := `
		UPDATE "Notifications" SET "status" = 1
		WHERE "id" = ?`
	_, err := db.Exec(sql, id)
	if err != nil {
		log.Fatal(err)
	}
}

func Hide(id int) {
	db := CreateConnection()
	sql := `
		UPDATE "Notifications" SET "status" = -1
		WHERE "id" = ?`
	_, err := db.Exec(sql, id)
	if err != nil {
		log.Fatal(err)
	}
}
