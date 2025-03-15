package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv        string `mapstructure:"APP_ENV"`
	AppHost       string `mapstructure:"APP_HOST"`
	AppPort       int    `mapstructure:"APP_PORT"`
	AppSecret     string `mapstructure:"GITHUB_SECRET"`
	MongoHost     string `mapstructure:"MONGO_HOST"`
	MongoPort     string `mapstructure:"MONGO_PORT"`
	MongoUsername string `mapstructure:"MOGNO_USERNAME"`
	MongoPassword string `mapstructure:"MONGO_PASS"`
	LogLevel      string `mapstructure:"LOG_LEVEL"`
}

func Load() Config {
	env := Config{}
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Cannot find .env file. Err: %s", err.Error())
		panic(err)
	}

	err := viper.Unmarshal(&env)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	if env.AppEnv == "development" {
		log.Println("The App is running in development env")
	}

	return env
}
