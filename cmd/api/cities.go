package main

import (
	"errors"
	"net/http"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/validator"
)

func (app *application) GetCities(w http.ResponseWriter, r *http.Request) {
	var input data.Filters
	v := validator.New()
	qs := r.URL.Query()

	if len(qs) > 0 {
		input.CountryCode = app.readString(qs, "country_code", "")
		if data.ValidateFilters(v, input); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
	}

	cities, err := app.models.Cities.GetCityList(input.CountryCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"cities": cities}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetCity(w http.ResponseWriter, r *http.Request) {
	var input struct {
		numbeoCost    string
		numbeoIndices string
		avgClimate    string
	}
	v := validator.New()
	qs := r.URL.Query()

	input.numbeoCost = app.readString(qs, "numbeo_cost", "")
	input.numbeoIndices = app.readString(qs, "numbeo_indices", "")
	input.avgClimate = app.readString(qs, "avg_climate", "")
	costEnabled := data.ValidateBoolQuery(v, input.numbeoCost)
	indicesEnabled := data.ValidateBoolQuery(v, input.numbeoIndices)
	climateEnabled := data.ValidateBoolQuery(v, input.avgClimate)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	city, err := app.models.Cities.GetCityID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if costEnabled {
		numbeoCost, err := app.models.Cities.GetNumbeoCost(id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		city.NumbeoCost = *numbeoCost
	}

	if indicesEnabled {
		numbeoIndicies, err := app.models.Cities.GetNumbeoIndicies(id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		city.NumbeoIndices = *numbeoIndicies
	}

	if climateEnabled {
		// TODO: Implement handling for the 'avg_climate' query parameter
		app.logger.Warn("Handling 'avg_climate' query parameter is incomplete.")
	}

	env := envelope{"city": city}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
