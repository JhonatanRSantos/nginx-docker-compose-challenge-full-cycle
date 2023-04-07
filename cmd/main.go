package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/goombaio/namegenerator"
)

func main() {
	db, err := openDBConnection()

	if err != nil {
		panic(err)
	}

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	defer db.Close()

	if err = initDB(db); err != nil {
		panic(err)
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		page := "<h1>Full Cycle Rocks!</h1>"

		conn, err := db.Conn(c.Context())

		if err == nil {
			defer conn.Close()

			if stmt, err := conn.PrepareContext(c.Context(), fmt.Sprintf("INSERT INTO people (name) VALUES ('%s')", nameGenerator.Generate())); err == nil {
				defer stmt.Close()
				stmt.ExecContext(c.Context())
			}

			if stmt, err := conn.PrepareContext(c.Context(), "SELECT name FROM people"); err == nil {
				defer stmt.Close()

				if rows, err := stmt.QueryContext(c.Context()); err == nil {
					names := ""
					for rows.Next() {
						var name string

						if err := rows.Scan(&name); err == nil && name != "" {
							names += "<li>" + name + "</li>"
						}
					}

					if names != "" {
						page += "<br>" + "<ul>" + names + "</ul>"
					}
				}
			}

		}
		return c.SendString(page)
	})

	app.Listen("0.0.0.0:80")
}

func openDBConnection() (*sql.DB, error) {
	maxAttempts := 6
	currentAttemp := 1

	host := "127.0.0.1"
	if os.Getenv("DB_HOST") != "" {
		host = os.Getenv("DB_HOST")
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:root@tcp(%s)/docker-compose-challenge", host))

	if err != nil {
		return db, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	log.Println("connecting to db. Attempt =", currentAttemp)
	err = db.Ping()

	for err != nil && currentAttemp < maxAttempts {
		currentAttemp++
		time.Sleep(time.Second * 5)
		log.Println("connecting to db. Attempt =", currentAttemp)
		err = db.Ping()
	}

	return db, err
}

func initDB(db *sql.DB) error {
	_, err := db.Query(`CREATE TABLE IF NOT EXISTS people (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR (255) NOT NULL
	)`)

	return err
}
