package data

import (
	"context"
	"database/sql"
	"fainal.net/internal/validator"
	"fmt"
	"github.com/lib/pq"
	"time"
)

type Item struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int32    `json:"price,omitempty"`
	Category    []string `json:"category,omitempty"`
}

func ValidateItem(v *validator.Validator, item *Item) {
	v.Check(item.Name != "", "item's name", "must be provided")
	v.Check(len(item.Name) <= 500, "item's name", "must not be more than 500 bytes long")
	v.Check(item.Description != "", "item's description", "must be provided")
	v.Check(len(item.Description) <= 2000, "item's description", "must not be more than 2000 bytes long")
	v.Check(item.Price != 0, "item's price", "must be provided")
	v.Check(item.Price >= 200, "item's price", "must be greater than 200 tenge")
	v.Check(item.Category != nil, "item's category", "must be provided")
	v.Check(len(item.Category) >= 1, "item's category", "must contain at least 1 category")
	v.Check(len(item.Category) <= 5, "item's category", "must not contain more than 5 category")
	v.Check(validator.Unique(item.Category), "item's category", "must not contain duplicate values")
}

type ItemModel struct {
	DB *sql.DB
}

func (m ItemModel) Insert(item *Item) error {
	query := `
		INSERT INTO items(name, description, price, category)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return m.DB.QueryRow(query, &item.Name, &item.Description, &item.Price, pq.Array(&item.Category)).Scan(&item.ID)
}

func (m ItemModel) Get(id int64) (*Item, error) {
	query := `
SELECT id, name, description, price, category
FROM items
WHERE id = $1`
	var item Item
	m.DB.QueryRow(query, id).Scan(&item.ID, &item.Name,
		&item.Description,
		&item.Price,
		pq.Array(&item.Category))
	return &item, m.DB.QueryRow(query, id).Err()
}

func (m ItemModel) GetAll(name string, category []string, filters Filters) ([]*Item, error) {

	query := fmt.Sprintf(`
SELECT id, name, description, price, category
FROM items
WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (category @> $2 OR $2 = '{}')
ORDER BY %s %s, id ASC
LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{name, pq.Array(category), filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []*Item{}
	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			pq.Array(&item.Category),
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m ItemModel) Update(item *Item) error {
	query := `
	UPDATE items
	SET name = $1, description = $2, price = $3, category = $4
	WHERE id = $5`

	return m.DB.QueryRow(query, &item.Name, &item.Description, &item.Price, pq.Array(&item.Category)).Scan(&item.ID)

}

func (m ItemModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	// Construct the SQL query to delete the record.
	query := `
DELETE FROM items
WHERE id = $1`

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
