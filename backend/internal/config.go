package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func LoadEnv() {
	// .env file is optional — environment variables may be set via Docker env_file or system env
	err := godotenv.Load(".env")
	if err != nil {
		log.Err(err)
	}
}

type Config struct {
	Debug  bool
	SECRET string

	MEDIA_PATH string
	ORIGINS    string

	// Frontend base URL
	BASE_URL string
	// Frontend base PATH: e.g.: /app
	FE_DASHBOARD_PATH string

	DB_URL string

	REDIS_HOST     string
	REDIS_PORT     int
	REDIS_DB_ASYNQ int
	REDIS_PASS     string

	SMTP_HOST     string
	SMTP_PORT     int
	SMTP_STARTTLS bool
	SMTP_USERNAME string
	SMTP_PASSWORD string

	OAuthGoogleClientID     string
	OAuthGoogleClientSecret string
	OAuthGitHubClientID     string
	OAuthGitHubClientSecret string
}

func NewConfig() *Config {
	// Load env
	LoadEnv()

	debug := false
	if os.Getenv("DEBUG") == "1" || strings.ToLower(os.Getenv("DEBUG")) == "true" {
		debug = true
	}

	mediaPath := os.Getenv("MEDIA_PATH")
	if mediaPath == "" {
		mediaPath = "./media"
	}

	origins := os.Getenv("ORIGINS")
	if origins == "" {
		origins = "*"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}

	redisPort := 6379
	redisPortStr, found := os.LookupEnv("REDIS_PORT")
	if found && redisPortStr != "" {
		redisPort, _ = strconv.Atoi(redisPortStr)
	}

	redisDBAsynq := 0
	redisDBAsyncStr, found := os.LookupEnv("REDIS_DB_ASYNQ")
	if found && redisDBAsyncStr != "" {
		redisDBAsynq, _ = strconv.Atoi(redisDBAsyncStr)
	}

	smtpPort := 587
	smtpPortStr, found := os.LookupEnv("SMTP_PORT")
	if found && smtpPortStr != "" {
		smtpPort, _ = strconv.Atoi(smtpPortStr)
	}

	smtpSsl := os.Getenv("SMTP_STARTTLS") == "1" || strings.ToLower(os.Getenv("SMTP_STARTTLS")) == "true"

	return &Config{
		Debug:      debug,
		SECRET:     os.Getenv("SECRET"),
		BASE_URL:   os.Getenv("BASE_URL"),
		MEDIA_PATH: mediaPath,
		ORIGINS:    origins,

		FE_DASHBOARD_PATH: os.Getenv("FE_DASHBOARD_PATH"),

		DB_URL: os.Getenv("DB_URL"),

		REDIS_HOST:     redisHost,
		REDIS_PORT:     redisPort,
		REDIS_DB_ASYNQ: redisDBAsynq,
		REDIS_PASS:     os.Getenv("REDIS_PASS"),

		SMTP_HOST:     os.Getenv("SMTP_HOST"),
		SMTP_PORT:     smtpPort,
		SMTP_STARTTLS: smtpSsl,
		SMTP_USERNAME: os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD: os.Getenv("SMTP_PASSWORD"),

		OAuthGoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		OAuthGoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		OAuthGitHubClientID:     os.Getenv("OAUTH_GITHUB_CLIENT_ID"),
		OAuthGitHubClientSecret: os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"),
	}
}

func NewTestConfig() *Config {
	// Load env
	LoadEnv()
	return &Config{
		DB_URL:         "test.db",
		Debug:          true,
		SECRET:         "secret",
		REDIS_DB_ASYNQ: 14,
	}
}

func (c *Config) RedisAddress() string {
	return fmt.Sprintf("%s:%d", c.REDIS_HOST, c.REDIS_PORT)
}

func (c *Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTP_HOST, c.SMTP_PORT)
}

func (c *Config) Check() error {
	if c.DB_URL == "" {
		return errors.New("Config: Invalid DB_URL")
	}
	if c.SECRET == "" {
		return errors.New("Config: Invalid SECRET")
	}
	if c.BASE_URL == "" {
		return errors.New("Config: Invalid BASE_URL (Web FE/UI)")
	}
	if c.SMTP_HOST == "" {
		return errors.New("Config: Invalid SMTP_HOST")
	}
	if c.SMTP_USERNAME == "" {
		return errors.New("Config: Invalid SMTP_USERNAME")
	}
	return nil
}

func (c *Config) ToJSON() []byte {
	data, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return data
}
