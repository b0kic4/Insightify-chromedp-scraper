package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Service interface {
	Health() map[string]string
	DB() *gorm.DB
}

type service struct {
	db *gorm.DB
}

func New() Service {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PW"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return &service{db: db}
}

func (s *service) DB() *gorm.DB {
	return s.db
}

func (s *service) Health() map[string]string {
	postgresDB, err := s.db.DB()
	if err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	err = postgresDB.Ping()
	if err != nil {
		log.Fatalf("DB down: %v", err)
	}

	return map[string]string{"message": "It's healthy"}
}
