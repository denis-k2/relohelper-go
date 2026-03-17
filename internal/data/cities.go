package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type City struct {
	ID                int64              `json:"geoname_id"`
	Name              string             `json:"city"`
	StateCode         *string            `json:"state_code"`
	CountryCode       string             `json:"country_code"`
	CountryName       string             `json:"country,omitzero"`
	Population        *int64             `json:"population,omitzero"`
	Latitude          float64            `json:"latitude"`
	Longitude         float64            `json:"longitude"`
	Timezone          string             `json:"timezone"`
	LastUpdate        string             `json:"last_update"`
	NumbeoCost        *NumbeoCost        `json:"numbeo_cost,omitzero"`
	NumbeoCityIndices *NumbeoCityIndices `json:"numbeo_indices,omitzero"`
	AvgClimate        *AvgClimate        `json:"avg_climate,omitzero"`
}

type NumbeoCost struct {
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

type NumbeoCityIndices struct {
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
	HighTemp     [12]*float64 `json:"high_temp"`
	LowTemp      [12]*float64 `json:"low_temp"`
	Pressure     [12]*float64 `json:"pressure"`
	WindSpeed    [12]*float64 `json:"wind_speed"`
	Humidity     [12]*float64 `json:"humidity"`
	Rainfall     [12]*float64 `json:"rainfall"`
	RainfallDays [12]*float64 `json:"rainfall_days"`
	Snowfall     [12]*float64 `json:"snowfall"`
	SnowfallDays [12]*float64 `json:"snowfall_days"`
	SeaTemp      [12]*float64 `json:"sea_temp"`
	Daylight     [12]*float64 `json:"daylight"`
	Sunshine     [12]*float64 `json:"sunshine"`
	SunshineDays [12]*float64 `json:"sunshine_days"`
	UVIndex      [12]*float64 `json:"uv_index"`
	CloudCover   [12]*float64 `json:"cloud_cover"`
	Visibility   [12]*float64 `json:"visibility"`
}

type avgClimateRaw struct {
	HighTemp     []*float64 `json:"high_temp"`
	LowTemp      []*float64 `json:"low_temp"`
	Pressure     []*float64 `json:"pressure"`
	WindSpeed    []*float64 `json:"wind_speed"`
	Humidity     []*float64 `json:"humidity"`
	Rainfall     []*float64 `json:"rainfall"`
	RainfallDays []*float64 `json:"rainfall_days"`
	Snowfall     []*float64 `json:"snowfall"`
	SnowfallDays []*float64 `json:"snowfall_days"`
	SeaTemp      []*float64 `json:"sea_temp"`
	Daylight     []*float64 `json:"daylight"`
	Sunshine     []*float64 `json:"sunshine"`
	SunshineDays []*float64 `json:"sunshine_days"`
	UVIndex      []*float64 `json:"uv_index"`
	CloudCover   []*float64 `json:"cloud_cover"`
	Visibility   []*float64 `json:"visibility"`
}

type CityModel struct {
	DB *sql.DB
}

func (c CityModel) ListCities(countryCode string, include IncludeSet) (cities []*City, retErr error) {
	query := `
		SELECT c.geoname_id, c.city, c.state_code, c.country_code,
		       ctr.country AS country, c.population, c.latitude, c.longitude, c.timezone,
		       to_char(c.updated_date, 'YYYY-MM-DD') AS last_update
		FROM cities c
		LEFT JOIN countries ctr ON ctr.country_code = c.country_code
		WHERE (LOWER(c.country_code) = LOWER($1) OR $1 = '')
		ORDER BY c.geoname_id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, countryCode)
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
			&city.ID,
			&city.Name,
			&city.StateCode,
			&city.CountryCode,
			&city.CountryName,
			&city.Population,
			&city.Latitude,
			&city.Longitude,
			&city.Timezone,
			&city.LastUpdate,
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
			c.geoname_id,
			c.city,
			c.state_code,
			c.country_code,
			ctr.country AS country,
			c.population,
			c.latitude,
			c.longitude,
			c.timezone,
			to_char(c.updated_date, 'YYYY-MM-DD') AS last_update,
			CASE
				WHEN $2 THEN (
					SELECT CASE
						WHEN COUNT(*) = 0 THEN NULL
						ELSE jsonb_build_object(
							'currency', 'USD',
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
					FROM numbeo_city_costs ns
					JOIN numbeo_cost_params np ON np.param_id = ns.param_id
					JOIN numbeo_cost_categories nc ON nc.category_id = np.category_id
					WHERE ns.geoname_id = c.geoname_id
				)
				ELSE NULL
			END AS numbeo_cost,
			CASE
				WHEN $3 THEN (
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
							to_char(nic.updated_date, 'YYYY-MM-DD') AS last_update
						FROM numbeo_city_indices nic
						WHERE nic.geoname_id = c.geoname_id
					) AS n
				)
				ELSE NULL
			END AS numbeo_indices,
			CASE
				WHEN $4 THEN (
					WITH climate_rows AS (
						SELECT *
						FROM avg_climate ac
						WHERE ac.geoname_id = c.geoname_id
					),
					climate_stats AS (
						SELECT
							COUNT(*) AS row_count,
							COUNT(DISTINCT month) AS unique_month_count,
							MIN(month) AS min_month,
							MAX(month) AS max_month
						FROM climate_rows
					),
					climate_data AS (
						SELECT
							m.metric_key,
							jsonb_agg(m.metric_value ORDER BY cr.month) AS month_values
						FROM climate_rows cr
						CROSS JOIN LATERAL jsonb_each(
							to_jsonb(cr)
								- 'geoname_id'
								- 'month'
								- 'updated_date'
								- 'updated_by'
						) AS m(metric_key, metric_value)
						GROUP BY m.metric_key
					)
					SELECT
						CASE
							WHEN s.row_count = 0
							THEN NULL
							WHEN s.row_count = 12
								AND s.unique_month_count = 12
								AND s.min_month = 1
								AND s.max_month = 12
							THEN (SELECT jsonb_object_agg(metric_key, month_values) FROM climate_data)
							ELSE jsonb_build_object('__invalid_structure__', true)
						END
					FROM climate_stats s
				)
				ELSE NULL
			END AS avg_climate
		FROM cities c
		LEFT JOIN countries ctr ON ctr.country_code = c.country_code
		WHERE c.geoname_id = $1;`

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
		include.Has("numbeo_cost"),
		include.Has("numbeo_indices"),
		include.Has("avg_climate"),
	).Scan(
		&city.ID,
		&city.Name,
		&city.StateCode,
		&city.CountryCode,
		&city.CountryName,
		&city.Population,
		&city.Latitude,
		&city.Longitude,
		&city.Timezone,
		&city.LastUpdate,
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
		var details NumbeoCost
		if err := json.Unmarshal(costJSON, &details); err != nil {
			return nil, err
		}
		city.NumbeoCost = &details
	}

	if len(indicesJSON) > 0 {
		var details NumbeoCityIndices
		if err := json.Unmarshal(indicesJSON, &details); err != nil {
			return nil, err
		}
		city.NumbeoCityIndices = &details
	}

	if len(climateJSON) > 0 && string(climateJSON) != "null" {
		climate, err := decodeAvgClimateSeries(id, climateJSON)
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
		SELECT c.geoname_id, c.city, c.state_code, c.country_code,
		       ctr.country AS country, c.population, c.latitude, c.longitude, c.timezone,
		       to_char(c.updated_date, 'YYYY-MM-DD') AS last_update
		FROM cities c
		LEFT JOIN countries ctr ON ctr.country_code = c.country_code
		WHERE c.geoname_id = ANY($1)
		ORDER BY c.geoname_id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	cities = []*City{}
	cityByID := make(map[int64]*City, len(ids))
	for rows.Next() {
		var city City
		if err := rows.Scan(
			&city.ID,
			&city.Name,
			&city.StateCode,
			&city.CountryCode,
			&city.CountryName,
			&city.Population,
			&city.Latitude,
			&city.Longitude,
			&city.Timezone,
			&city.LastUpdate,
		); err != nil {
			return nil, err
		}
		cityPtr := &city
		cities = append(cities, cityPtr)
		cityByID[city.ID] = cityPtr
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(cities) == 0 {
		return nil, ErrRecordNotFound
	}

	if include.Has("numbeo_cost") {
		err = c.attachNumbeoCostByCityIDs(ctx, ids, cityByID)
		if err != nil {
			return nil, err
		}
	}

	if include.Has("numbeo_indices") {
		err = c.attachNumbeoCityIndicesByCityIDs(ctx, ids, cityByID)
		if err != nil {
			return nil, err
		}
	}

	if include.Has("avg_climate") {
		err = c.attachAvgClimateByCityIDs(ctx, ids, cityByID)
		if err != nil {
			return nil, err
		}
	}

	return cities, nil
}

func (c CityModel) attachNumbeoCostByCityIDs(ctx context.Context, ids []int64, cityByID map[int64]*City) (retErr error) {
	query := `
		SELECT
			ns.geoname_id,
			jsonb_build_object(
				'currency', 'USD',
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
			) AS numbeo_cost
		FROM numbeo_city_costs ns
		JOIN numbeo_cost_params np ON np.param_id = ns.param_id
		JOIN numbeo_cost_categories nc ON nc.category_id = np.category_id
		WHERE ns.geoname_id = ANY($1)
		GROUP BY ns.geoname_id;`

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	for rows.Next() {
		var (
			cityID  int64
			rawJSON []byte
		)

		if err := rows.Scan(&cityID, &rawJSON); err != nil {
			return err
		}
		if len(rawJSON) == 0 || string(rawJSON) == "null" {
			continue
		}

		var details NumbeoCost
		if err := json.Unmarshal(rawJSON, &details); err != nil {
			return err
		}

		city, ok := cityByID[cityID]
		if !ok {
			continue
		}
		city.NumbeoCost = &details
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (c CityModel) attachNumbeoCityIndicesByCityIDs(ctx context.Context, ids []int64, cityByID map[int64]*City) (retErr error) {
	query := `
		SELECT
			nic.geoname_id,
			row_to_json(n) AS numbeo_indices
		FROM numbeo_city_indices nic
		CROSS JOIN LATERAL (
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
				to_char(nic.updated_date, 'YYYY-MM-DD') AS last_update
		) AS n
		WHERE nic.geoname_id = ANY($1);`

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	for rows.Next() {
		var (
			cityID  int64
			rawJSON []byte
		)

		if err := rows.Scan(&cityID, &rawJSON); err != nil {
			return err
		}
		if len(rawJSON) == 0 || string(rawJSON) == "null" {
			continue
		}

		var details NumbeoCityIndices
		if err := json.Unmarshal(rawJSON, &details); err != nil {
			return err
		}

		city, ok := cityByID[cityID]
		if !ok {
			continue
		}
		city.NumbeoCityIndices = &details
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (c CityModel) attachAvgClimateByCityIDs(ctx context.Context, ids []int64, cityByID map[int64]*City) (retErr error) {
	query := `
		WITH climate_rows AS (
			SELECT *
			FROM avg_climate ac
			WHERE ac.geoname_id = ANY($1)
		),
		climate_stats AS (
			SELECT
				geoname_id,
				COUNT(*) AS row_count,
				COUNT(DISTINCT month) AS unique_month_count,
				MIN(month) AS min_month,
				MAX(month) AS max_month
			FROM climate_rows
			GROUP BY geoname_id
		),
		climate_data AS (
			SELECT
				cr.geoname_id,
				m.metric_key,
				jsonb_agg(m.metric_value ORDER BY cr.month) AS month_values
			FROM climate_rows cr
			CROSS JOIN LATERAL jsonb_each(
				to_jsonb(cr)
					- 'geoname_id'
					- 'month'
					- 'updated_date'
					- 'updated_by'
			) AS m(metric_key, metric_value)
			GROUP BY cr.geoname_id, m.metric_key
		)
		SELECT
			s.geoname_id,
			CASE
				WHEN s.row_count = 12
					AND s.unique_month_count = 12
					AND s.min_month = 1
					AND s.max_month = 12
				THEN (
					SELECT jsonb_object_agg(cd.metric_key, cd.month_values)
					FROM climate_data cd
					WHERE cd.geoname_id = s.geoname_id
				)
				ELSE jsonb_build_object('__invalid_structure__', true)
			END AS avg_climate
		FROM climate_stats s;`

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	for rows.Next() {
		var (
			cityID  int64
			rawJSON []byte
		)

		if err := rows.Scan(&cityID, &rawJSON); err != nil {
			return err
		}
		if len(rawJSON) == 0 || string(rawJSON) == "null" {
			continue
		}

		avgClimate, err := decodeAvgClimateSeries(cityID, rawJSON)
		if err != nil {
			return err
		}

		city, ok := cityByID[cityID]
		if !ok {
			continue
		}
		city.AvgClimate = avgClimate
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func decodeAvgClimateSeries(cityID int64, rawJSON []byte) (*AvgClimate, error) {
	var raw avgClimateRaw
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		return nil, err
	}

	series := map[string][]*float64{
		"high_temp":     raw.HighTemp,
		"low_temp":      raw.LowTemp,
		"pressure":      raw.Pressure,
		"wind_speed":    raw.WindSpeed,
		"humidity":      raw.Humidity,
		"rainfall":      raw.Rainfall,
		"rainfall_days": raw.RainfallDays,
		"snowfall":      raw.Snowfall,
		"snowfall_days": raw.SnowfallDays,
		"sea_temp":      raw.SeaTemp,
		"daylight":      raw.Daylight,
		"sunshine":      raw.Sunshine,
		"sunshine_days": raw.SunshineDays,
		"uv_index":      raw.UVIndex,
		"cloud_cover":   raw.CloudCover,
		"visibility":    raw.Visibility,
	}

	for metric, values := range series {
		if len(values) != 12 {
			return nil, fmt.Errorf("invalid avg_climate series for city_id=%d metric=%s: got %d values, expected 12", cityID, metric, len(values))
		}
	}

	return &AvgClimate{
		HighTemp:     toMonthArray(raw.HighTemp),
		LowTemp:      toMonthArray(raw.LowTemp),
		Pressure:     toMonthArray(raw.Pressure),
		WindSpeed:    toMonthArray(raw.WindSpeed),
		Humidity:     toMonthArray(raw.Humidity),
		Rainfall:     toMonthArray(raw.Rainfall),
		RainfallDays: toMonthArray(raw.RainfallDays),
		Snowfall:     toMonthArray(raw.Snowfall),
		SnowfallDays: toMonthArray(raw.SnowfallDays),
		SeaTemp:      toMonthArray(raw.SeaTemp),
		Daylight:     toMonthArray(raw.Daylight),
		Sunshine:     toMonthArray(raw.Sunshine),
		SunshineDays: toMonthArray(raw.SunshineDays),
		UVIndex:      toMonthArray(raw.UVIndex),
		CloudCover:   toMonthArray(raw.CloudCover),
		Visibility:   toMonthArray(raw.Visibility),
	}, nil
}

func toMonthArray(values []*float64) [12]*float64 {
	var result [12]*float64
	copy(result[:], values)
	return result
}
