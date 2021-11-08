package internal

import "log"

type Contact struct {
	username string
	status   int
}

func AddPendingContact(username1 string, username2 string) {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	log.Println("Creating friend invitation from " + username1 + " to " + username2)

	sql := `
	INSERT INTO "Contacts"(user1, user2, friendstatus) 
	VALUES (?, ?, ?)`

	statement, err := db.Prepare(sql)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(username1, username2, 0)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(username1 + " sent a friend request to " + username2)
}

func ApproveContact(username1 string, username2 string) {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	log.Println("Creating friendship between " + username1 + " and " + username2)

	sql := `
		UPDATE "Contacts" 
		SET friendstatus=? 
		WHERE user1=?, user2=?`

	statement, err := db.Prepare(sql)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(1, username1, username2)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Friendship between " + username1 + " and " + username2 + " established!")
}

func RejectContact(username1 string, username2 string) {
	db := CreateConnection()
	defer db.Close()

	//TODO: Validation

	sql := `
		DELETE FROM "Contacts" 
		WHERE user1=?, user2=?`

	statement, err := db.Prepare(sql)

	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(username1, username2)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(username2 + " declined friend request of " + username1)
}

func GetContactList(uid string) (contactList []User) {
	db := CreateConnection()
	defer db.Close()

	// TODO: Validation

	sql := `
		SELECT "fullname", "username" 
		FROM "Users"
		WHERE "username" IN
		(SELECT "user1" FROM "Contacts" 
		WHERE "user2" = ?)
		OR "username" IN 
		(SELECT "user2" FROM "Contacts" 
		WHERE "user1" = ?)`

	row, err := db.Query(sql, uid, uid)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()

	for row.Next() {
		var user User
		row.Scan(&user.DisplayName, &user.Username)
		contactList = append(contactList, user)
	}
	log.Println(contactList)
	return
}
