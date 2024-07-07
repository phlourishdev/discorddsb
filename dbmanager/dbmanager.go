package dbmanager

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

func InitDB() (*sql.DB, error) {
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		return nil, fmt.Errorf("MYSQL_USER environment variable is not set")
	}

	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("MYSQL_PASSWORD environment variable is not set")
	}

	dbName := os.Getenv("MYSQL_DATABASE")
	if dbName == "" {
		return nil, fmt.Errorf("MYSQL_DATABASE environment variable is not set")
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", user, password, dbName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer db.Close()

	time.Sleep(5 * time.Second)

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	log.Printf("Database connection established successfully.")
	return db, nil
}
