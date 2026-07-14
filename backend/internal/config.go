package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
}

type Config struct {
	Debug      bool
	SECRET     string
	MEDIA_PATH string
	ORIGINS    string

	DB_URL string

	REDIS_HOST string
	REDIS_PORT int
	REDIS_DB   int
	REDIS_PASS string

	SMTP_HOST     string
	SMTP_PORT     int
	SMTP_STARTTLS bool
	SMTP_USERNAME string
	SMTP_PASSWORD string

	OAuthGoogleClientID     string
	OAuthGoogleClientSecret string
	OAuthGitHubClientID     string
	OAuthGitHubClientSecret string
	OAuthRedirectBase       string
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
		origins = "http://127.0.0.1:5173, http://localhost:5173"
	}

	redisHost, found := os.LookupEnv("REDIS_HOST")
	if !found {
		redisHost = "127.0.0.1"
	}

	redisPort := 6379
	redisPortStr, found := os.LookupEnv("REDIS_HOST")
	if found {
		redisPort, _ = strconv.Atoi(redisPortStr)
	}

	redisDB := 0
	redisDBStr, found := os.LookupEnv("REDIS_DB")
	if found {
		redisDB, _ = strconv.Atoi(redisDBStr)
	}

	smtpPort := 587
	smtpPortStr, found := os.LookupEnv("SMTP_PORT")
	if found {
		smtpPort, _ = strconv.Atoi(smtpPortStr)
	}

	oauthRedirectBase := os.Getenv("OAUTH_REDIRECT_BASE")
	if oauthRedirectBase == "" {
		oauthRedirectBase = "http://localhost:5173"
	}

	return &Config{
		Debug:      debug,
		SECRET:     os.Getenv("SECRET"),
		MEDIA_PATH: mediaPath,
		ORIGINS:    origins,

		DB_URL: os.Getenv("DB_URL"),

		REDIS_HOST: redisHost,
		REDIS_PORT: redisPort,
		REDIS_DB:   redisDB,
		REDIS_PASS: os.Getenv("REDIS_PASS"),

		SMTP_HOST:     os.Getenv("SMTP_HOST"),
		SMTP_PORT:     smtpPort,
		SMTP_STARTTLS: os.Getenv("SMTP_STARTTLS") != "",
		SMTP_USERNAME: os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD: os.Getenv("SMTP_PASSWORD"),

		OAuthGoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		OAuthGoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		OAuthGitHubClientID:     os.Getenv("OAUTH_GITHUB_CLIENT_ID"),
		OAuthGitHubClientSecret: os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"),
		OAuthRedirectBase:       oauthRedirectBase,
	}
}

func NewTestConfig() *Config {
	// Load env
	LoadEnv()
	return &Config{
		DB_URL:   "test.db",
		Debug:    true,
		SECRET:   "secret",
		REDIS_DB: 14,
	}
}

func (c *Config) RedisAddress() string {
	return fmt.Sprintf("%s:%d", c.REDIS_HOST, c.REDIS_PORT)
}

func (c *Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTP_HOST, c.SMTP_PORT)
}
