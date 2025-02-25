package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type City struct {
	CityID      int64   `json:"city_id"`
	City        string  `json:"city"`
	StateCode   *string `json:"state_code"`
	CountryCode string  `json:"country_code"`
}

type CityModel struct {
	DB *sql.DB
}

func (c CityModel) GetCityList(countryСode string) ([]*City, error) {
	query := `
        SELECT city_id, city, state_code, country_code
		FROM city
		WHERE (LOWER(country_code) = LOWER($1) OR $1 = '')
		ORDER BY city_id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, countryСode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cities := []*City{}

	for rows.Next() {
		var city City

		err := rows.Scan(
			&city.CityID,
			&city.City,
			&city.StateCode,
			&city.CountryCode,
		)

		if err != nil {
			return nil, err
		}

		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cities, nil
}

func (c CityModel) GetCityID(id int64) (*City, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT city_id, city, state_code, country_code
		FROM city
		WHERE city_id = $1;`

	var city City
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&city.CityID,
		&city.City,
		&city.StateCode,
		&city.CountryCode,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &city, nil
}
