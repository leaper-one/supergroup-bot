package durable

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(ctx context.Context) *gorm.DB {
	return NewDB()
}

func NewDB() *gorm.DB {
	var err error
	cfg := config.Config.Database
	connStr := ""
	if cfg.Port == "" {
		connStr = fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", cfg.Host, cfg.User, cfg.Name, cfg.Password)
	} else {
		connStr = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Password)
	}
	DB, err := gorm.Open(postgres.Open(connStr),
		&gorm.Config{
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			}),
		},
	)
	if err != nil {
		panic(err)
	}
	return DB
}
func CheckNotEmptyError(err error) error {
	if err == nil || IsEmpty(err) {
		return nil
	}
	return err
}

func CheckIsPKRepeatError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func IsEmpty(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
