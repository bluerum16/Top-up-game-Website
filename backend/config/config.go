package config

import "os"

type Config struct {
	DBhost string
	DBport string
	DBuser string
	DBpass string
	DBname string
}

func Setupconf() Config {
	return Config{
		DBhost: os.Getenv("DB_MAIN_HOST"),
		DBport: os.Getenv("DB_MAIN_PORT"),
		DBuser: os.Getenv("DB_MAIN_USER"),
		DBpass: os.Getenv("DB_MAIN_PASS"),
		DBname: os.Getenv("DB_MAIN_NAME"),

	}
}
