package data

import (
	"context"
	"database/sql"
	"time"
)

type Country struct {
	Code string `json:"country_code"`
	Name string `json:"country"`
}

type CountryModel struct {
	DB *sql.DB
}

func (c CountryModel) GetCountryList() ([]*Country, error) {
	query := `
        SELECT country_code, country
		FROM country
		ORDER BY country_code;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	countries := []*Country{}

	for rows.Next() {
		var country Country

		err := rows.Scan(
			&country.Code,
			&country.Name,
		)

		if err != nil {
			return nil, err
		}

		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}
