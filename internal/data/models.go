package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Cities    CityModel
	Countries CountryModel
	Users     UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Cities:    CityModel{DB: db},
		Countries: CountryModel{DB: db},
		Users:     UserModel{DB: db},
	}
}
