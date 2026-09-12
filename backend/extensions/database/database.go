package database

import (
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(url string) {
	// newLogger := logger.New(
	// 	// log.New(os.Stdout, "\r\n", log.LstdFlags),
	// 	nil,
	// 	logger.Config{
	// 		// SlowThreshold: time.Microsecond,
	// 		// LogLevel:      logger.Info,
	// 		// Colorful:      false,
	// 	},
	// )

	var dialector gorm.Dialector
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		dialector = postgres.Open(url)
	} else {
		dialector = sqlite.Open(url)
	}

	_db, err := gorm.Open(dialector, &gorm.Config{
		// Logger: newLogger,
		Logger: nil,
	})
	if err != nil {
		log.Panic().Err(err)
	}
	DB = _db
}

// For testing purposes
func ConnectTestDatabase() {
	_db, err := gorm.Open(sqlite.Open("test.db"))
	if err != nil {
		log.Panic().Err(err)
	}
	DB = _db
}
