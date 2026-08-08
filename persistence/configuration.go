package persistence

import (
	"os"
)

type DatabaseConfiguration struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func (DatabaseConfiguration) Get() *DatabaseConfiguration {
	return &DatabaseConfiguration{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}
}
