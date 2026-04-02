package main

import (
	"github.com/irham/topup-backend/config"
	"github.com/irham/topup-backend/database"
	"github.com/irham/topup-backend/router"
)

func main() {
    conf := config.Setupconf()
    r := router.SetupRouter()
    database.ConnectDB(conf)
    r.Run(":8080")
}