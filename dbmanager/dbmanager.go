package dbmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

var db = connectToDB()

// getDBCredentials retrieves DB credentials from environment variables
func getDBCredentials() ([]string, error) {
	// get environment variables
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

	return []string{user, password, dbName}, nil
}

// getConnectionString creates a string used for sql.Open()
func getConectionString() string {
	dbCreds, err := getDBCredentials()

	user := dbCreds[0]
	password := dbCreds[1]
	dbName := dbCreds[2]

	if err != nil {
		log.Fatalf("Attaining database credentials failed: %v", err.Error())
	}

	// configure db connection
	connectionString := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", user, password, dbName)

	return connectionString
}

// ConnectToDB open the DB connection and return the DB pointer
func connectToDB() *sql.DB {
	connectionString := getConectionString()

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Opening the database failed: %v", err.Error())
	}
	return db
}

// pingDB tests the connection to the DB via db.Ping()
func pingDB() {
	// test db connection
	time.Sleep(5 * time.Second)

	err := db.Ping()
	if err != nil {
		_ = db.Close()
		log.Fatalf("Pinging the database failed: %v; Closing connection.", err.Error())
	}

	log.Print("Database could be pinged.")
}

// createTable executes a query for table creation
func createTable(tableQuery string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, tableQuery)

	if err != nil {
		return err
	}

	return nil
}

// tableExists checks if a table with the given name exists in the database; returns false on error
func tableExists(tableName string) (bool, error) {
	var exists bool
	dbCreds, err := getDBCredentials()

	dbName := dbCreds[2]

	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", dbName, tableName).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// getCreateTableQuery provides database queries for table creation according to given table name
func getCreateTableQuery(tableName string) (string, error) {
	var tableQuery string

	if tableName == "days" {
		tableQuery = `CREATE TABLE IF NOT EXISTS days (
			id INT AUTO_INCREMENT PRIMARY KEY,
			date DATE,
			weekday VARCHAR(50)
		)`
	} else if tableName == "substitutes" { // TODO: get these from the substitute plan or config file or sth
		tableQuery = `CREATE TABLE IF NOT EXISTS substitutes (
			id INT AUTO_INCREMENT PRIMARY KEY,
			class VARCHAR(10),
    		substitutetype VARCHAR(10),
			newsubject VARCHAR(10),
    		room VARCHAR(10),
    		oldsubject VARCHAR(10),
    		movedfrom VARCHAR(50),
    		notice VARCHAR(50)
		)`
	} else if tableName == "d_s_relation" {
		tableQuery = `CREATE TABLE IF NOT EXISTS d_s_relation (
    		d_id INT,
			s_id INT,
			CONSTRAINT fk_days FOREIGN KEY (d_id) REFERENCES days(id),
			CONSTRAINT fk_substitutes FOREIGN KEY (s_id) REFERENCES substitutes(id)
		)`
	} else {
		return "", errors.New("Table '" + tableName + "' not in configured table queries")
	}

	return tableQuery, nil
}

// createMissingTables checks if multiple tables exist in the DB
func createMissingTables(tablesToCheck []string) error {
	var err error
	var exists bool
	var tableQuery string

	for i := range tablesToCheck {
		exists, err = tableExists(tablesToCheck[i])
		if err != nil {
			return err
		}
		if !exists {
			tableQuery, err = getCreateTableQuery(tablesToCheck[i])
			if err != nil {
				return err
			}
			err = createTable(tableQuery)
			if err != nil {
				return err
			}
			log.Printf("Created table %s because it could not be found in DB", tablesToCheck[i])
		}
	}
	log.Print("Database includes all necessary tables.")
	return nil
}

// InitializeDB allows for the initialization of the DB connection from outside the package
func InitializeDB() {
	pingDB()

	tablesToCheck := []string{"days", "substitutes", "d_s_relation"}
	err := createMissingTables(tablesToCheck)

	if err != nil {
		log.Fatal(err)
	}

}

// CloseDB allows the closing of the DB connection from outside the package
func CloseDB() {
	err := db.Close()
	if err != nil {
		return
	}
}
