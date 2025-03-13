package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Country struct {
	Code          string                `json:"country_code"`
	Name          string                `json:"country"`
	NumbeoIndices NumbeoCountryIndicies `json:"numbeo_indices,omitzero"`
}

type NumbeoCountryIndicies struct {
	CostOfLiving               *float64 `json:"cost_of_living"`
	Rent                       *float64 `json:"rent"`
	CostOfLivingPlusRent       *float64 `json:"cost_of_living_plus_rent"`
	Groceries                  *float64 `json:"groceries"`
	RestaurantPrice            *float64 `json:"restaurant_price"`
	LocalPurchasingPower       *float64 `json:"local_purchasing_power"`
	QualityOfLife              *float64 `json:"quality_of_life"`
	PropertyPriceToIncomeRatio *float64 `json:"property_price_to_income_ratio"`
	TrafficCommuteTime         *float64 `json:"traffic_commute_time"`
	Climate                    *float64 `json:"climate"`
	Safety                     *float64 `json:"safety"`
	HealthCare                 *float64 `json:"health_care"`
	Pollution                  *float64 `json:"pollution"`
	AvgSalaryUSD               *float64 `json:"avg_salary_usd"`
	LastUpdate                 string   `json:"last_update"`
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

func (c CountryModel) GetCountry(countryCode string) (*Country, error) {
	query := `
		SELECT country_code, country
		FROM country
		WHERE LOWER(country_code) = LOWER($1);`

	var country Country
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, countryCode).Scan(
		&country.Code,
		&country.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &country, nil
}

func (c CountryModel) GetNumbeoCountryIndicies(country_code string) (*NumbeoCountryIndicies, error) {
	query := `
		SELECT
			cost_of_living,
			rent,
			cost_of_living_plus_rent,
			groceries,
			restaurant_price,
			local_purchasing_power,
			quality_of_life,
			property_price_to_income_ratio,
			traffic_commute_time,
			climate,
			safety,
			health_care,
			pollution,
			avg_salary_usd,
			to_char(sys_updated_date, 'YYYY-MM-DD')
		FROM numbeo_index_by_country
		WHERE LOWER(country_code) = LOWER($1);`

	var index NumbeoCountryIndicies

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, country_code).Scan(
		&index.CostOfLiving,
		&index.Rent,
		&index.CostOfLivingPlusRent,
		&index.Groceries,
		&index.RestaurantPrice,
		&index.LocalPurchasingPower,
		&index.QualityOfLife,
		&index.PropertyPriceToIncomeRatio,
		&index.TrafficCommuteTime,
		&index.Climate,
		&index.Safety,
		&index.HealthCare,
		&index.Pollution,
		&index.AvgSalaryUSD,
		&index.LastUpdate,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &index, nil
}
