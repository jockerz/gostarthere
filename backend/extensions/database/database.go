package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(url string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Microsecond, // Slow SQL threshold
			LogLevel:      logger.Info,      // Log level
			// IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			// ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful: false, // Disable color
		},
	)
	_db, err := gorm.Open(sqlite.Open(url), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
	DB = _db
}

// For testing purposes
func ConnectTestDatabase() {
	_db, err := gorm.Open(sqlite.Open("test.db"))
	if err != nil {
		panic(err)
	}
	DB = _db
}
