package main

import "github.com/denis-k2/relohelper-go/internal/data"

type cityResponse struct {
	CityID        int64   `json:"city_id"`
	City          string  `json:"city"`
	StateCode     *string `json:"state_code"`
	CountryCode   string  `json:"country_code"`
	Country       string  `json:"country,omitzero"`
	NumbeoCost    any     `json:"numbeo_cost,omitempty"`
	NumbeoIndices any     `json:"numbeo_indices,omitempty"`
	AvgClimate    any     `json:"avg_climate,omitempty"`
}

func newCityResponse(city *data.City, include data.IncludeSet) cityResponse {
	res := cityResponse{
		CityID:      city.ID,
		City:        city.Name,
		StateCode:   city.StateCode,
		CountryCode: city.CountryCode,
		Country:     city.Country,
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
	NumbeoIndices  any    `json:"numbeo_indices,omitempty"`
	LegatumIndices any    `json:"legatum_indices,omitempty"`
}

func newCountryResponse(country *data.Country, include data.IncludeSet) countryResponse {
	res := countryResponse{
		Code: country.Code,
		Name: country.Name,
	}

	if include.Has("numbeo_indices") {
		res.NumbeoIndices = country.NumbeoCountryIndices
	}
	if include.Has("legatum_indices") {
		res.LegatumIndices = country.LegatumCountryIndices
	}

	return res
}
