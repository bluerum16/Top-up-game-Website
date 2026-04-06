package repository

import "github.com/jackc/pgx/v5/pgxpool"

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo (db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db:db}
}

