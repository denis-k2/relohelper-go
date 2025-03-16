package main

import (
	"errors"
	"net/http"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/validator"
)

func (app *application) listCitiesHandler(w http.ResponseWriter, r *http.Request) {
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
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	env := envelope{"cities": cities}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCityHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

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
		switch {
		case err == nil:
			city.NumbeoCost = numbeoCost
		case errors.Is(err, data.ErrRecordNotFound):
			city.NumbeoCost = nil
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if indicesEnabled {
		numbeoIndicies, err := app.models.Cities.GetNumbeoIndicies(id)
		switch {
		case err == nil:
			city.NumbeoIndices = numbeoIndicies
		case errors.Is(err, data.ErrRecordNotFound):
			city.NumbeoIndices = nil
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if climateEnabled {
		avgClimate, err := app.models.Cities.GetAvgClimatePivot(id)
		switch {
		case err == nil:
			city.AvgClimate = avgClimate
		case errors.Is(err, data.ErrRecordNotFound):
			city.AvgClimate = nil
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	env := envelope{"city": city}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
