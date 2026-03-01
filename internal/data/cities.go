package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
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

func (c CityModel) ListCities(countryCode string, include IncludeSet) (cities []*City, retErr error) {
	query := `
		SELECT c.city_id, c.city, c.state_code, c.country_code,
		       CASE WHEN $2 THEN ctr.country ELSE '' END AS country
		FROM city c
		LEFT JOIN country ctr ON ctr.country_code = c.country_code
		WHERE (LOWER(c.country_code) = LOWER($1) OR $1 = '')
		ORDER BY c.city_id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, countryCode, include.Has("country"))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	cities = []*City{}
	for rows.Next() {
		var city City
		if err := rows.Scan(
			&city.CityID,
			&city.City,
			&city.StateCode,
			&city.CountryCode,
			&city.Country,
		); err != nil {
			return nil, err
		}
		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(cities) == 0 {
		return nil, ErrRecordNotFound
	}

	return cities, nil
}

func (c CityModel) GetCity(id int64, include IncludeSet) (*City, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT
			c.city_id,
			c.city,
			c.state_code,
			c.country_code,
			CASE WHEN $2 THEN ctr.country ELSE '' END AS country,
			CASE
				WHEN $3 THEN (
					SELECT CASE
						WHEN COUNT(*) = 0 THEN NULL
						ELSE jsonb_build_object(
							'currency', MAX(ns.currency),
							'last_update', to_char(MAX(ns.updated_date), 'YYYY-MM-DD'),
							'prices', jsonb_agg(
								jsonb_build_object(
									'category', nc.category,
									'param', np.param,
									'cost', ns.cost,
									'range_lower', lower(ns.range),
									'range_upper', upper(ns.range)
								)
								ORDER BY nc.category, np.param
							)
						)
					END
					FROM numbeo_stat ns
					JOIN numbeo_param np ON np.param_id = ns.param_id
					JOIN numbeo_category nc ON nc.category_id = np.category_id
					WHERE ns.city_id = c.city_id
				)
				ELSE NULL
			END AS numbeo_cost,
			CASE
				WHEN $4 THEN (
					SELECT row_to_json(n)
					FROM (
						SELECT
							nic.cost_of_living,
							nic.rent,
							nic.cost_of_living_plus_rent,
							nic.groceries,
							nic.local_purchasing_power,
							nic.quality_of_life,
							nic.property_price_to_income_ratio,
							nic.traffic_commute_time,
							nic.climate,
							nic.safety,
							nic.health_care,
							nic.pollution,
							to_char(nic.sys_updated_date, 'YYYY-MM-DD') AS last_update
						FROM numbeo_index_by_city nic
						WHERE nic.city_id = c.city_id
					) AS n
				)
				ELSE NULL
			END AS numbeo_indices,
			CASE
				WHEN $5 THEN (
					SELECT jsonb_agg(
						jsonb_build_object(
							'climate_param', ac.climate_param,
							'january', ac.january,
							'february', ac.february,
							'march', ac.march,
							'april', ac.april,
							'may', ac.may,
							'june', ac.june,
							'july', ac.july,
							'august', ac.august,
							'september', ac.september,
							'october', ac.october,
							'november', ac.november,
							'december', ac.december
						)
					)
					FROM pivot_avg_climate ac
					WHERE ac.city_id = c.city_id
				)
				ELSE NULL
			END AS avg_climate
		FROM city c
		LEFT JOIN country ctr ON ctr.country_code = c.country_code
		WHERE c.city_id = $1;`

	var (
		city        City
		costJSON    []byte
		indicesJSON []byte
		climateJSON []byte
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(
		ctx,
		query,
		id,
		include.Has("country"),
		include.Has("numbeo_cost"),
		include.Has("numbeo_indices"),
		include.Has("avg_climate"),
	).Scan(
		&city.CityID,
		&city.City,
		&city.StateCode,
		&city.CountryCode,
		&city.Country,
		&costJSON,
		&indicesJSON,
		&climateJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if len(costJSON) > 0 {
		var details CostDetails
		if err := json.Unmarshal(costJSON, &details); err != nil {
			return nil, err
		}
		city.NumbeoCost = &details
	}

	if len(indicesJSON) > 0 {
		var details Indices
		if err := json.Unmarshal(indicesJSON, &details); err != nil {
			return nil, err
		}
		city.NumbeoIndices = &details
	}

	if len(climateJSON) > 0 && string(climateJSON) != "null" {
		climate, err := buildAvgClimateFromJSON(climateJSON)
		if err != nil {
			return nil, err
		}
		city.AvgClimate = climate
	}

	return &city, nil
}

func (c CityModel) GetCitiesByIDs(ids []int64, include IncludeSet) (cities []*City, retErr error) {
	if len(ids) == 0 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT c.city_id, c.city, c.state_code, c.country_code,
		       CASE WHEN $2 THEN ctr.country ELSE '' END AS country
		FROM city c
		LEFT JOIN country ctr ON ctr.country_code = c.country_code
		WHERE c.city_id = ANY($1)
		ORDER BY c.city_id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(ids), include.Has("country"))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	cities = []*City{}
	for rows.Next() {
		var city City
		if err := rows.Scan(
			&city.CityID,
			&city.City,
			&city.StateCode,
			&city.CountryCode,
			&city.Country,
		); err != nil {
			return nil, err
		}
		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(cities) == 0 {
		return nil, ErrRecordNotFound
	}

	return cities, nil
}

type climateRow struct {
	ClimateParam string   `json:"climate_param"`
	January      *float64 `json:"january"`
	February     *float64 `json:"february"`
	March        *float64 `json:"march"`
	April        *float64 `json:"april"`
	May          *float64 `json:"may"`
	June         *float64 `json:"june"`
	July         *float64 `json:"july"`
	August       *float64 `json:"august"`
	September    *float64 `json:"september"`
	October      *float64 `json:"october"`
	November     *float64 `json:"november"`
	December     *float64 `json:"december"`
}

func buildAvgClimateFromJSON(raw []byte) (*AvgClimate, error) {
	var rows []climateRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	climate := &AvgClimate{Measures: measures}
	for _, row := range rows {
		monthly := MonthlyValue{
			January:   row.January,
			February:  row.February,
			March:     row.March,
			April:     row.April,
			May:       row.May,
			June:      row.June,
			July:      row.July,
			August:    row.August,
			September: row.September,
			October:   row.October,
			November:  row.November,
			December:  row.December,
		}

		switch row.ClimateParam {
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

	return climate, nil
}
