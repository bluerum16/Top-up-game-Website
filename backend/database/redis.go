package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var Ctx = context.Background()