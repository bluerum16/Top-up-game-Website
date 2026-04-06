package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/irham/topup-backend/config"
	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var Ctx = context.Background()

func ConnectRed(conf config.Config) {
	addr := fmt.Sprintf("%s:%s", conf.REDIS_HOST, conf.REDIS_PORT)

	Redis := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: conf.REDIS_PASS,
		DB: 0,
	})

	Ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Unable connect redis: ", err)
	}


}