package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/irham/topup-backend/config"
)

var DB *sql.DB

func ConnectDB(conf config.Config) {
	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DBuser,
		conf.DBpass,
		conf.DBhost,
		conf.DBport,
		conf.DBname,
	)
	var err error

	DB, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal("Failed connect: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("DB error: ", err)
	}

	log.Println("Connected to DB")
}