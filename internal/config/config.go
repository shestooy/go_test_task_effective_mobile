package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DataBaseEndPoint string
	ServerEndPoint   string
	LogLevel         string
	DefaultLimit     int
	DefaultPage      int
	DefaultVerse     int
}

func NewConfig() *Config {
	dbEndPoint := os.Getenv("DATABASE_ENDPOINT")
	if dbEndPoint == "" {
		dbEndPoint = "postgresql://postgres:root@localhost:5432/db"
	}

	serverEndPoint := os.Getenv("SERVER_ENDPOINT")
	if serverEndPoint == "" {
		serverEndPoint = "http://127.0.0.1:8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	defaultLimitStr := os.Getenv("DEFAULT_LIMIT")
	defaultLimit, err := strconv.Atoi(defaultLimitStr)
	if err != nil {
		defaultLimit = 5
	}

	DefaultPageStr := os.Getenv("DEFAULT_PAGE")
	defaultPage, err := strconv.Atoi(DefaultPageStr)
	if err != nil {
		defaultPage = 1
	}
	defaultVerseStr := os.Getenv("DEFAULT_VERSE")
	defaultVerse, err := strconv.Atoi(defaultVerseStr)
	if err != nil {
		defaultVerse = 1
	}
	return &Config{
		DataBaseEndPoint: dbEndPoint,
		ServerEndPoint:   serverEndPoint,
		LogLevel:         logLevel,
		DefaultLimit:     defaultLimit,
		DefaultPage:      defaultPage,
		DefaultVerse:     defaultVerse,
	}
}

func init() {
	err := godotenv.Load("./config/config.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
