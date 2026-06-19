package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/denis-k2/relohelper-go/internal/db"
)

type City struct {
	GeonameID         int64              `json:"geoname_id"`
	Name              string             `json:"city"`
	StateCode         *string            `json:"state_code,omitzero"`
	CountryCode       string             `json:"country_code"`
	CountryName       string             `json:"country"`
	Population        *int64             `json:"population"`
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
	Queries *db.Queries
}

func (c CityModel) ListCities(countryCode string, include IncludeSet) ([]*City, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.Queries.ListCities(ctx, countryCode)
	if err != nil {
		return nil, err
	}

	cities := make([]*City, 0, len(rows))
	for _, row := range rows {
		cities = append(cities, newCityFromListRow(row))
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row, err := c.Queries.GetCityDetailed(ctx, db.GetCityDetailedParams{
		IncludeNumbeoCost:    include.Has("numbeo_cost"),
		IncludeNumbeoIndices: include.Has("numbeo_indices"),
		IncludeAvgClimate:    include.Has("avg_climate"),
		GeonameID:            id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return newCityFromDetailedRow(row)
}

func (c CityModel) GetCitiesByIDs(ids []int64, include IncludeSet) ([]*City, error) {
	if len(ids) == 0 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.Queries.GetCitiesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	cities := make([]*City, 0, len(rows))
	cityByID := make(map[int64]*City, len(ids))
	for _, row := range rows {
		city := newCityFromIDsRow(row)
		cities = append(cities, city)
		cityByID[city.GeonameID] = city
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

func (c CityModel) attachNumbeoCostByCityIDs(ctx context.Context, ids []int64, cityByID map[int64]*City) error {
	rows, err := c.Queries.GetNumbeoCostByCityIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row.NumbeoCost) == 0 || string(row.NumbeoCost) == "null" {
			continue
		}

		var details NumbeoCost
		if err := json.Unmarshal(row.NumbeoCost, &details); err != nil {
			return err
		}

		city, ok := cityByID[row.GeonameID]
		if !ok {
			continue
		}
		city.NumbeoCost = &details
	}

	return nil
}

func (c CityModel) attachNumbeoCityIndicesByCityIDs(ctx context.Context, ids []int64, cityByID map[int64]*City) error {
	rows, err := c.Queries.GetNumbeoCityIndicesByCityIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row.NumbeoIndices) == 0 || string(row.NumbeoIndices) == "null" {
			continue
		}

		var details NumbeoCityIndices
		if err := json.Unmarshal(row.NumbeoIndices, &details); err != nil {
			return err
		}

		city, ok := cityByID[row.GeonameID]
		if !ok {
			continue
		}
		city.NumbeoCityIndices = &details
	}

	return nil
}

func (c CityModel) attachAvgClimateByCityIDs(ctx context.Context, ids []int64, cityByID map[int64]*City) error {
	rows, err := c.Queries.GetAvgClimateByCityIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		rawJSON, err := jsonRawBytes(row.AvgClimate)
		if err != nil {
			return err
		}
		if len(rawJSON) == 0 || string(rawJSON) == "null" {
			continue
		}

		avgClimate, err := decodeAvgClimateSeries(row.GeonameID, rawJSON)
		if err != nil {
			return err
		}

		city, ok := cityByID[row.GeonameID]
		if !ok {
			continue
		}
		city.AvgClimate = avgClimate
	}

	return nil
}

func newCityFromListRow(row db.ListCitiesRow) *City {
	return &City{
		GeonameID:   row.GeonameID,
		Name:        row.City,
		StateCode:   row.StateCode,
		CountryCode: row.CountryCode,
		CountryName: stringValue(row.Country),
		Population:  row.Population,
		Latitude:    float64Value(row.Latitude),
		Longitude:   float64Value(row.Longitude),
		Timezone:    stringValue(row.Timezone),
		LastUpdate:  row.LastUpdate,
	}
}

func newCityFromDetailedRow(row db.GetCityDetailedRow) (*City, error) {
	city := &City{
		GeonameID:   row.GeonameID,
		Name:        row.City,
		StateCode:   row.StateCode,
		CountryCode: row.CountryCode,
		CountryName: stringValue(row.Country),
		Population:  row.Population,
		Latitude:    float64Value(row.Latitude),
		Longitude:   float64Value(row.Longitude),
		Timezone:    stringValue(row.Timezone),
		LastUpdate:  row.LastUpdate,
	}

	costJSON, err := jsonRawBytes(row.NumbeoCost)
	if err != nil {
		return nil, err
	}
	if len(costJSON) > 0 && string(costJSON) != "null" {
		var details NumbeoCost
		if err := json.Unmarshal(costJSON, &details); err != nil {
			return nil, err
		}
		city.NumbeoCost = &details
	}

	indicesJSON, err := jsonRawBytes(row.NumbeoIndices)
	if err != nil {
		return nil, err
	}
	if len(indicesJSON) > 0 && string(indicesJSON) != "null" {
		var details NumbeoCityIndices
		if err := json.Unmarshal(indicesJSON, &details); err != nil {
			return nil, err
		}
		city.NumbeoCityIndices = &details
	}

	climateJSON, err := jsonRawBytes(row.AvgClimate)
	if err != nil {
		return nil, err
	}
	if len(climateJSON) > 0 && string(climateJSON) != "null" {
		avgClimate, err := decodeAvgClimateSeries(row.GeonameID, climateJSON)
		if err != nil {
			return nil, err
		}
		city.AvgClimate = avgClimate
	}

	return city, nil
}

func newCityFromIDsRow(row db.GetCitiesByIDsRow) *City {
	return &City{
		GeonameID:   row.GeonameID,
		Name:        row.City,
		StateCode:   row.StateCode,
		CountryCode: row.CountryCode,
		CountryName: stringValue(row.Country),
		Population:  row.Population,
		Latitude:    float64Value(row.Latitude),
		Longitude:   float64Value(row.Longitude),
		Timezone:    stringValue(row.Timezone),
		LastUpdate:  row.LastUpdate,
	}
}

func jsonRawBytes(value any) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func float64Value(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
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
