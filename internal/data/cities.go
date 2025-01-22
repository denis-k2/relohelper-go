package data

import (
	"context"
	"database/sql"
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

func (c CityModel) GetCityList() ([]*City, error) {
	query := `
        SELECT city_id, city, state_code, country_code
		FROM city
		ORDER BY city_id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
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
