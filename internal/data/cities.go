package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type City struct {
	CityID        int64        `json:"city_id"`
	City          string       `json:"city"`
	StateCode     *string      `json:"state_code"`
	CountryCode   string       `json:"country_code"`
	Country       string       `json:"country,omitzero"`
	NumbeoCost    *CostDetails `json:"numbeo_cost,omitzero"`
	NumbeoIndices *Indices     `json:"numbeo_indices,omitzero"`
	AvgClimate    *AvgClimate  `json:"avg_climate,omitzero"`
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

type AvgClimate struct {
	HighTemp     MonthlyValue      `json:"high_temp"`
	LowTemp      MonthlyValue      `json:"low_temp"`
	Pressure     MonthlyValue      `json:"pressure"`
	WindSpeed    MonthlyValue      `json:"wind_speed"`
	Humidity     MonthlyValue      `json:"humidity"`
	Rainfall     MonthlyValue      `json:"rainfall"`
	RainfallDays MonthlyValue      `json:"rainfall_days"`
	Snowfall     MonthlyValue      `json:"snowfall"`
	SnowfallDays MonthlyValue      `json:"snowfall_days"`
	SeaTemp      MonthlyValue      `json:"sea_temp"`
	Daylight     MonthlyValue      `json:"daylight"`
	Sunshine     MonthlyValue      `json:"sunshine"`
	SunshineDays MonthlyValue      `json:"sunshine_days"`
	UVIndex      MonthlyValue      `json:"uv_index"`
	CloudCover   MonthlyValue      `json:"cloud_cover"`
	Visibility   MonthlyValue      `json:"visibility"`
	Measures     map[string]string `json:"measures"`
}

type MonthlyValue struct {
	January   *float64 `json:"january"`
	February  *float64 `json:"february"`
	March     *float64 `json:"march"`
	April     *float64 `json:"april"`
	May       *float64 `json:"may"`
	June      *float64 `json:"june"`
	July      *float64 `json:"july"`
	August    *float64 `json:"august"`
	September *float64 `json:"september"`
	October   *float64 `json:"october"`
	November  *float64 `json:"november"`
	December  *float64 `json:"december"`
}

var measures = map[string]string{
	"high_temp":     "Average high temperature, °C",
	"low_temp":      "Average low temperature, °C",
	"pressure":      "Average pressure, mbar",
	"wind_speed":    "Average wind speed, km/h",
	"humidity":      "Average humidity, %",
	"rainfall":      "Average rainfall, mm",
	"rainfall_days": "Average rainfall days, days",
	"snowfall":      "Average snowfall, mm",
	"snowfall_days": "Average snowfall days, days",
	"sea_temp":      "Average sea temperature, °C",
	"daylight":      "Average daylight, hours",
	"sunshine":      "Average sunshine, hours",
	"sunshine_days": "Average sunshine days, days",
	"uv_index":      "Average UV index",
	"cloud_cover":   "Average cloud cover, %",
	"visibility":    "Average visibility, km",
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

	dataFound := false
	for rows.Next() {
		dataFound = true
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

	if !dataFound {
		return nil, ErrRecordNotFound
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
	defer rows.Close()

	dataFound := false
	for rows.Next() {
		dataFound = true
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if !dataFound {
		return nil, ErrRecordNotFound
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

func (c CityModel) GetAvgClimatePivot(id int64) (*AvgClimate, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT
			climate_param,
			january,
			february,
			march,
			april,
			may,
			june,
			july,
			august,
			september,
			october,
			november,
			december
		FROM pivot_avg_climate
		WHERE city_id = $1;`

	var climate AvgClimate
	climate.Measures = measures

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dataFound := false
	for rows.Next() {
		dataFound = true
		var param string
		var monthly MonthlyValue

		err := rows.Scan(
			&param,
			&monthly.January,
			&monthly.February,
			&monthly.March,
			&monthly.April,
			&monthly.May,
			&monthly.June,
			&monthly.July,
			&monthly.August,
			&monthly.September,
			&monthly.October,
			&monthly.November,
			&monthly.December,
		)
		if err != nil {
			return nil, err
		}

		switch param {
		case "high_temp":
			climate.HighTemp = monthly
		case "low_temp":
			climate.LowTemp = monthly
		case "pressure":
			climate.Pressure = monthly
		case "wind_speed":
			climate.WindSpeed = monthly
		case "humidity":
			climate.Humidity = monthly
		case "rainfall":
			climate.Rainfall = monthly
		case "rainfall_days":
			climate.RainfallDays = monthly
		case "snowfall":
			climate.Snowfall = monthly
		case "snowfall_days":
			climate.SnowfallDays = monthly
		case "sea_temp":
			climate.SeaTemp = monthly
		case "daylight":
			climate.Daylight = monthly
		case "sunshine":
			climate.Sunshine = monthly
		case "sunshine_days":
			climate.SunshineDays = monthly
		case "uv_index":
			climate.UVIndex = monthly
		case "cloud_cover":
			climate.CloudCover = monthly
		case "visibility":
			climate.Visibility = monthly
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if !dataFound {
		return nil, ErrRecordNotFound
	}

	return &climate, nil
}
