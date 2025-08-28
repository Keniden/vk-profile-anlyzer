package db

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgconn"
)

type DataBase struct {
	cluster *pgxpool.Pool
}

func NewDatabase(cluster *pgxpool.Pool) *DataBase {
	return &DataBase{
		cluster: cluster,
	}
}

func (d DataBase) GetPool(_ context.Context) *pgxpool.Pool {
	return d.cluster
}

func(d DataBase) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error{
    return pgxscan.Get(ctx, d.cluster, dest, query, args)
}

func(d DataBase) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, d.cluster, dest, query, args)
}

func(d DataBase) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return d.cluster.Exec(ctx, query, args...)
}

func (d DataBase) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return d.cluster.QueryRow(ctx, query, args...)
} 
