package main

import "github.com/denis-k2/relohelper-go/internal/data"

type cityResponse struct {
	CityID        int64   `json:"geoname_id"`
	City          string  `json:"city"`
	StateCode     *string `json:"state_code,omitzero"`
	CountryCode   string  `json:"country_code"`
	Country       string  `json:"country"`
	Population    *int64  `json:"population"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timezone      string  `json:"timezone"`
	LastUpdate    string  `json:"last_update"`
	NumbeoCost    any     `json:"numbeo_cost,omitzero"`
	NumbeoIndices any     `json:"numbeo_indices,omitzero"`
	AvgClimate    any     `json:"avg_climate,omitzero"`
}

func newCityResponse(city *data.City, include data.IncludeSet) cityResponse {
	res := cityResponse{
		CityID:      city.GeonameID,
		City:        city.Name,
		StateCode:   city.StateCode,
		CountryCode: city.CountryCode,
		Country:     city.CountryName,
		Population:  city.Population,
		Latitude:    city.Latitude,
		Longitude:   city.Longitude,
		Timezone:    city.Timezone,
		LastUpdate:  city.LastUpdate,
	}

	if include.Has("numbeo_cost") {
		res.NumbeoCost = city.NumbeoCost
	}
	if include.Has("numbeo_indices") {
		res.NumbeoIndices = city.NumbeoCityIndices
	}
	if include.Has("avg_climate") {
		res.AvgClimate = city.AvgClimate
	}

	return res
}

type countryResponse struct {
	Code           string `json:"country_code"`
	Name           string `json:"country"`
	Population     *int64 `json:"population"`
	Area           *int64 `json:"area"`
	LastUpdate     string `json:"last_update"`
	NumbeoIndices  any    `json:"numbeo_indices,omitzero"`
	LegatumIndices any    `json:"legatum_indices,omitzero"`
}

func newCountryResponse(country *data.Country, include data.IncludeSet) countryResponse {
	res := countryResponse{
		Code:       country.Code,
		Name:       country.Name,
		Population: country.Population,
		Area:       country.Area,
		LastUpdate: country.LastUpdate,
	}

	if include.Has("numbeo_indices") {
		res.NumbeoIndices = country.NumbeoCountryIndices
	}
	if include.Has("legatum_indices") {
		res.LegatumIndices = country.LegatumCountryIndices
	}

	return res
}
