package data

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-k2/relohelper-go/internal/db"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Cities    CityModel
	Countries CountryModel
	Tokens    TokenModel
	Users     UserModelInterface
}

func NewModels(pool *pgxpool.Pool) Models {
	queries := db.New(pool)
	return Models{
		Cities:    CityModel{Queries: queries},
		Countries: CountryModel{Queries: queries},
		Tokens:    TokenModel{Queries: queries},
		Users:     UserModel{Queries: queries},
	}
}
