package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"example/internal/pkg/db"
	"example/internal/pkg/repository"
)

type ArticleRepo struct {
	db *db.DataBase
}

func NewArticles(db *db.DataBase) *ArticleRepo {
	return &ArticleRepo{db: db}
}

func (r *ArticleRepo) Add(ctx context.Context, article *repository.Article) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, "INSERT INTO articles(name,rating) VALUES($1,$2) RETURNING id;", article.Name, article.Rating).Scan(&id)
	return id, err
}

func (r *ArticleRepo) GetByID(ctx context.Context, id int64) (*repository.Article, error) {
	var a repository.Article
	err := r.db.Get(ctx, &a,
		"SELECT id, name, rating, created_at FROM articles WHERE id= any($1);",
		id, // id уже int64
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrObjNotFound
	}
	return &a, nil
}
