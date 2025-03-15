package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	AppHost         string
	AppPort         string
	AppSecret       string
	MongoStringConn string
	LogLevel        string
}

func Load() Config {

	if err := godotenv.Load(); err != nil {
		log.Printf("Cannot find .env file, using default env. Err: %s", err.Error())
	}
	env := Config{
		AppEnv:          os.Getenv("APP_ENV"),
		AppHost:         os.Getenv("APP_HOST"),
		AppPort:         os.Getenv("APP_PORT"),
		AppSecret:       os.Getenv("GITHUB_SECRET"),
		MongoStringConn: os.Getenv("MONGO_STRINGCONN"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
	}

	if env.AppEnv == "development" {
		log.Println("The App is running in development env")
	}

	return env
}
