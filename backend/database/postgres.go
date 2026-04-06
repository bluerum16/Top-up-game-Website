package database

import (
	"context"
	"fmt"
	"log"

	"github.com/irham/topup-backend/config"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *pgxpool.Pool

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

	DB, err = pgxpool.New(context.Background(), conn)
	if err != nil {
		log.Fatal("Failed connect: ", err)
	}

	log.Println("Connected to DB")
}