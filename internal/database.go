package internal

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const database = "/home/wallius/localDB/voicechat.db"

func CreateConnection() *sql.DB {
	// Verify database
	if fileExist(database) {
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

func Prep() {
	createTable(CreateConnection())
}

func createTable(db *sql.DB) {
	// Users table SQL creation statement
	createUsersTableSQL := `CREATE TABLE "Users" (
		"username" TEXT PRIMARY KEY,
		"password" TEXT,
		"fullname" TEXT,
		"nickname" TEXT,
		"active" INTEGER);`

	// Friend table SQL creation statement
	createFriendTableSQL := `CREATE TABLE "Contacts" (
		"user1" TEXT,
		"user2" TEXT,
		status INTEGER,
		FOREIGN KEY(user1) REFERENCES Users(username),
		FOREIGN KEY(user2) REFERENCES Users(username),
		PRIMARY KEY (user1, user2));`

	createNotifTableSQL := `CREATE TABLE 'Notifications' (
			'id' INTEGER PRIMARY KEY,
			'uid' TEXT,
			'timestamp' DATETIME,
			'data' TEXT,
			'status' INTEGER);`

	log.Println("Creating [Users] table...")

	_, err := db.Exec(createUsersTableSQL)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Table [Users] created")

	// Create Friend table
	log.Println("Creating [Friend] table...")

	_, err = db.Exec(createFriendTableSQL)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Table [Friend] created")

	_, err = db.Exec(createNotifTableSQL)

	if err != nil {
		log.Fatal(err)
	}
}

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
