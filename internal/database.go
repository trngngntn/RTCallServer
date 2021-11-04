package internal

import (
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const database = "../data/voicechat.db"

type Friend struct {
	username string
	status int
}

func CreateConnection()(*sql.DB) {
	// Verify database
	if (fileExist(database)) {
		log.Println("Database connected")
		db, _ := sql.Open("sqlite3", database)
		return db
	} else {
		log.Println("DB file not exist! Creating...")
		file, err := os.Create(database)
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
		log.Println("Db file created")

		db, err := sql.Open("sqlite3", database)

		createTable(db)

		return db
	}
}

func createTable(db *sql.DB) {
	// Users table SQL creation statement
	createUsersTableSQL := `CREATE TABLE Users (
		"username" TEXT PRIMARY KEY,
		"password" TEXT,
		"fullname" TEXT,
		"nickname" TEXT,
		"active" INTEGER);`

	// Friend table SQL creation statement
	createFriendTableSQL := `CREATE TABLE Friend (
		"user1" TEXT,
		"user2" TEXT,
		"friendstatus" INTEGER,
		FOREIGN KEY(user1) REFERENCES Users(username),
		FOREIGN KEY(user2) REFERENCES Users(username),
		PRIMARY KEY (user1, user2));`
	
	log.Println("Creating [Users] table...")

	statement, err := db.Prepare(createUsersTableSQL)

	if (err != nil) {
		log.Fatal(err)
	}

	statement.Exec()
	log.Println("Table [Users] created")

	// Create Friend table
	log.Println("Creating [Friend] table...")

	statement, err = db.Prepare(createFriendTableSQL)

	if (err != nil) {
		log.Fatal(err)
	}

	statement.Exec()
	log.Println("Table [Friend] created")
}

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if (os.IsNotExist(err)) {
		return false
	}

	return !info.IsDir()
}

func RegisteredUser(username string, password string, fullname string, nickname string) {
	db := CreateConnection()
	defer db.Close()
	//TODO: Validation

	log.Println("Inserting new user...")
	insertUserSQL := `INSERT INTO Users(username, password, fullname, nickname, active) VALUES (?, ?, ?, ?, ?)`

	statement, err := db.Prepare(insertUserSQL)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(username, password, fullname, nickname, 1)

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

	if  err != nil {
		return false
	} else {
		if userCount <= 0 {
			return false
		}
		return true
	}
	
}

func CreateFriendRequest(username1 string, username2 string) {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	log.Println("Creating friend invitation from " + username1 + " to " + username2)

	insertFriendSQL := `INSERT INTO Friend(user1, user2, friendstatus) VALUES (?, ?, ?)`

	statement, err := db.Prepare(insertFriendSQL)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(username1, username2, 0)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(username1 + " sent a friend request to " + username2)
}

func AcceptFriendRequest(username1 string, username2 string) {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	log.Println("Creating friendship between " + username1 + " and " + username2)

	insertFriendSQL := `UPDATE Friend SET friendstatus=? WHERE user1=?, user2=?`

	statement, err := db.Prepare(insertFriendSQL)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(1, username1, username2)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Friendship between " + username1 + " and " + username2 + " established!")
}

func DeclineFriendRequest(username1 string, username2 string) {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	insertFriendSQL := `DELETE FROM Friend WHERE user1=?, user2=?`

	statement, err := db.Prepare(insertFriendSQL)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(username1, username2)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(username2 + " declined friend request of " + username1)
}

func GetFriendList(uid string) ([]Friend){
	db := CreateConnection()
	defer db.Close()

	// TODO: Validation

	row, err := db.Query("SELECT Friend.user2, Friend.friendstatus FROM Friend INNER JOIN Users ON Friend.user1 = Users.username WHERE Friend.user1 = ? AND friendstatus = 1", uid)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()
	var friends []Friend
	for row.Next() {
		var uid string
		var status int
		row.Scan(&uid, &status)
		friends = append(friends, Friend{uid, status})
	}
	log.Println(friends)
	return friends
}

func DisplayFriend(uid string) {
	db := CreateConnection()
	defer db.Close()

	// TODO: Validation

	row, err := db.Query("SELECT Friend.user2 FROM Friend INNER JOIN Users ON Friend.user1 = Users.username WHERE Friend.user1 = ? AND friendstatus = 1", uid)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()

	for row.Next() {
		var uid string
		row.Scan(&uid)
		log.Println("Friend: ", uid)
	}
}