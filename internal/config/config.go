package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port             string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPass           string
	DBName           string
	RedisHost        string
	RedisPort        string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env") // Look for .env in the root
	viper.AutomaticEnv()        // Read environment variables that are set in the system

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("No .env file found, using system environment variables...")
	}

	config := &Config{
		Port:      viper.GetString("PORT"),
		DBHost:    viper.GetString("DB_HOST"),
		DBPort:    viper.GetString("DB_PORT"),
		DBUser:    viper.GetString("DB_USER"),
		DBPass:    viper.GetString("DB_PASS"),
		DBName:    viper.GetString("DB_NAME"),
		RedisHost: viper.GetString("REDIS_HOST"),
		RedisPort: viper.GetString("REDIS_PORT"),
	}

	return config
}
