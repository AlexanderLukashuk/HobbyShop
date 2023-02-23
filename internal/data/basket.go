package data

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"time"
)

type Basket struct {
	ID      int64   `json:"id"`
	Items   []int64 `json:"items"`
	User_id int64   `json:"user_id"`
}
type BasketModel struct {
	DB *sql.DB
}

func (b BasketModel) BasketInsert(basket *Basket) error {
	query := `
INSERT INTO baskets (items, user_id)
VALUES ($1, $2)
RETURNING id, items, user_id`
	args := []any{pq.Array(basket.Items), basket.User_id}
	// Create a context with a 3-second timeout.
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Use QueryRowContext() and pass the context as the first argument.
	return b.DB.QueryRowContext(ctx, query, args...).Scan(&basket.ID, pq.Array(&basket.Items), &basket.User_id)
}

func (b BasketModel) GetBasket(id int64) (*Basket, error) {
	query := `
SELECT user_id, items, id
FROM baskets
WHERE user_id = $1`
	var basket Basket
	b.DB.QueryRow(query, id).Scan(&basket.User_id, pq.Array(&basket.Items), &basket.ID)
	return &basket, b.DB.QueryRow(query, id).Err()
}

func (m BasketModel) UpdateBasket(basket *Basket) error {
	query := `
	UPDATE baskets
	SET items = $1
	WHERE user_id = $2`
	return m.DB.QueryRow(query, pq.Array(&basket.Items)).Scan(&basket.ID)
}

func (m BasketModel) DeleteBasket(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
DELETE FROM baskets
WHERE user_id = $1`

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil

}
