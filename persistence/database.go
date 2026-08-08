package persistence

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewConnection() *sqlx.DB {
	connectionString := ConnectionStringBuilder{}.Build()

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	return db
}
