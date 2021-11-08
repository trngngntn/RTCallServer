package internal

import (
	"database/sql"
	"log"
	"net"
)

type User struct {
	ID          int    `json:"uid"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	password    string
}

var MapClient = make(map[string]*Client)
var MapAddr = make(map[net.Addr]string)

type Client struct {
	SocketConn   net.Conn
	DatagramAddr string
}

func UsernameExists(username string) bool {
	query := `SELECT "username"
	FROM "Users" 
	WHERE "Users"."username" = ?`
	db := CreateConnection()

	row := db.QueryRow(query, username)

	switch err := row.Scan(&username); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		panic(err)
	}
}

func CreateNewUser(username string, password string, fullname string) {
	db := CreateConnection()
	defer db.Close()
	//TODO: Validation

	log.Println("Inserting new user...")
	insertUserSQL := `INSERT INTO Users(username, password, fullname, active) VALUES (?, ?, ?, 1)`

	statement, err := db.Prepare(insertUserSQL)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(username, password, fullname)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("User " + username + " inserted!")
}

func Login(username string, password string) bool {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	log.Println("Verifying user " + username)

	userCount := 0

	err := db.QueryRow("SELECT COUNT(*) FROM Users WHERE username=? AND password=? AND active=1", username, password).Scan(&userCount)

	if err != nil {
		return false
	} else {
		if userCount <= 0 {
			return false
		}
		return true
	}

}
