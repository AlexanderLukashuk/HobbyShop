package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record (row, entry) not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Items   ItemModel
	Users   UserModel
	Tokens  TokenModel
	Baskets BasketModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Items:   ItemModel{DB: db},
		Users:   UserModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Baskets: BasketModel{DB: db},
	}
}
