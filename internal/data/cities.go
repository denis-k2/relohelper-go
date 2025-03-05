package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type City struct {
	CityID        int64       `json:"city_id"`
	City          string      `json:"city"`
	StateCode     *string     `json:"state_code"`
	CountryCode   string      `json:"country_code"`
	Country       string      `json:"country,omitzero"`
	NumbeoCost    CostDetails `json:"numbeo_cost,omitzero"`
	NumbeoIndices Indices     `json:"numbeo_indices,omitzero"`
}

type CostDetails struct {
	Currency   string  `json:"currency"`
	LastUpdate string  `json:"last_update"`
	Prices     []Price `json:"prices"`
}

type Price struct {
	Category   string   `json:"category"`
	Param      string   `json:"param"`
	Cost       *float64 `json:"cost"`
	RangeLower *float64 `json:"range_lower"`
	RangeUpper *float64 `json:"range_upper"`
}

type Indices struct {
	CostOfLiving               *float64 `json:"cost_of_living"`
	Rent                       *float64 `json:"rent"`
	CostOfLivingPlusRent       *float64 `json:"cost_of_living_plus_rent"`
	Groceries                  *float64 `json:"groceries"`
	LocalPurchasingPower       *float64 `json:"local_purchasing_power"`
	QualityOfLife              *float64 `json:"quality_of_life"`
	PropertyPriceToIncomeRatio *float64 `json:"property_price_to_income_ratio"`
	TrafficCommuteTime         *float64 `json:"traffic_commute_time"`
	Climate                    *float64 `json:"climate"`
	Safety                     *float64 `json:"safety"`
	HealthCare                 *float64 `json:"health_care"`
	Pollution                  *float64 `json:"pollution"`
	LastUpdate                 string   `json:"last_update"`
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
        SELECT c.city_id , c.city, c.state_code , c.country_code, ctr.country
		FROM city c
		JOIN country ctr on c.country_code = ctr.country_code
		WHERE c.city_id = $1;`

	var city City
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&city.CityID,
		&city.City,
		&city.StateCode,
		&city.CountryCode,
		&city.Country,
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

func (c CityModel) GetNumbeoCost(id int64) (*CostDetails, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT 
        	nc.category,
        	np.param,
        	ns.cost,
        	lower(ns.range) AS lower_bound,
			upper(ns.range) AS upper_bound,
			ns.currency,
        	to_char(ns.updated_date, 'YYYY-MM-DD')
		FROM numbeo_stat ns 
		JOIN numbeo_param np ON ns.param_id = np.param_id
		JOIN numbeo_category nc ON np.category_id = nc.category_id
		WHERE ns.city_id = $1;`

	var cost CostDetails

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var price Price
		err := rows.Scan(
			&price.Category,
			&price.Param,
			&price.Cost,
			&price.RangeLower,
			&price.RangeUpper,
			&cost.Currency,
			&cost.LastUpdate,
		)
		if err != nil {
			return nil, err
		}
		cost.Prices = append(cost.Prices, price)
	}

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &cost, nil
}

func (c CityModel) GetNumbeoIndicies(id int64) (*Indices, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT
			cost_of_living,
			rent,
			cost_of_living_plus_rent,
			groceries,
			local_purchasing_power,
			quality_of_life,
			property_price_to_income_ratio,
			traffic_commute_time,
			climate,
			safety,
			health_care,
			pollution,
			to_char(sys_updated_date, 'YYYY-MM-DD')
		FROM public.numbeo_index_by_city
		WHERE city_id = $1;`

	var index Indices

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&index.CostOfLiving,
		&index.Rent,
		&index.CostOfLivingPlusRent,
		&index.Groceries,
		&index.LocalPurchasingPower,
		&index.QualityOfLife,
		&index.PropertyPriceToIncomeRatio,
		&index.TrafficCommuteTime,
		&index.Climate,
		&index.Safety,
		&index.HealthCare,
		&index.Pollution,
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
