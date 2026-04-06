package config

import "os"

type Config struct {
	DBhost string
	DBport string
	DBuser string
	DBpass string
	DBname string
	REDIS_HOST string
	REDIS_PORT string
	REDIS_PASS string
}

func Setupconf() Config {
	return Config{
		DBhost: os.Getenv("DB_MAIN_HOST"),
		DBport: os.Getenv("DB_MAIN_PORT"),
		DBuser: os.Getenv("DB_MAIN_USER"),
		DBpass: os.Getenv("DB_MAIN_PASS"),
		DBname: os.Getenv("DB_MAIN_NAME"),
		REDIS_HOST: os.Getenv("REDIS_HOST"),
		REDIS_PORT: os.Getenv("REDIS_PORT"),
		REDIS_PASS: os.Getenv("REDIS_PASS"),
	}
}
