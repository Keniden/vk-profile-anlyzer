package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	host     = "localhost"
	port     = 5432
	username = "postgres"
	password = "postgres"
	dbname   = "cruddb"
)

func NewDB(ctx context.Context) (*DataBase, error) {
	pool, err := pgxpool.New(ctx, generateDSN())
	if err != nil {
		return nil, err
	}
	return NewDatabase(pool), nil
}

func generateDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname)
}
