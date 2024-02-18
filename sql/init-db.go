package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
    // Check if the data directory exists
    _, err := os.Stat("./data")
    if os.IsNotExist(err) {
        // Create the data directory
        errDir := os.MkdirAll("./data", 0755)
        if errDir != nil {
            log.Fatal(err)
        }
    }

    dbPath := "./data/news.db"

    // Check if the database file exists
    _, err = os.Stat(dbPath)
    if os.IsNotExist(err) {
        // Create the database file
        file, errFile := os.Create(dbPath)
        if errFile != nil {
            log.Fatal(errFile)
        }
        file.Close()
    }

    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS edition (
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS analytics (
			user_id TEXT,
			insertion_timestamp DATETIME,
			payload TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id TEXT PRIMARY KEY,
			title TEXT,
			description TEXT,
			compressed_content BLOB,
			image_url TEXT,
			link TEXT,
			author TEXT,
			source TEXT,
			timestamp DATETIME,
			ts TEXT,
			layout TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS sources (
            id TEXT PRIMARY KEY,
            owner_id TEXT,
            name TEXT,
            url TEXT,
            feed_url TEXT,
            categories TEXT,
            disable_fetch BOOLEAN
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            name TEXT,
            created DATETIME,
            password_hash BLOB,
            is_admin BOOLEAN
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

}