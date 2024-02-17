// init-db.go
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./data/editions.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS editions (
            id TEXT PRIMARY KEY,
            name TEXT,
            date TEXT,
            start_time DATETIME,
            end_time DATETIME,
            created DATETIME,
            sources TEXT,
            articles TEXT,
            categories TEXT,
            metadata TEXT
        )
    `)
    if err != nil {
        log.Fatal(err)
    }
}