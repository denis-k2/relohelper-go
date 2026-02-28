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

	err := validateAllowedQueryParams(qs, newIncludeSet("country_code", "include", "ids"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	_, err = parseInclude(qs, newIncludeSet("country"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"include": err.Error()})
		return
	}

	_, idsPresent, err := parseIDsInt64(qs, "ids", app.config.batch.maxIDs)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"ids": err.Error()})
		return
	}
	if idsPresent {
		app.errorResponse(w, r, http.StatusNotImplemented, "batch retrieval for cities by ids is not implemented yet")
		return
	}

	if qs.Has("country_code") {
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

	qs := r.URL.Query()
	err = validateAllowedQueryParams(qs, newIncludeSet("include"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	include, err := parseInclude(qs, newIncludeSet("country", "numbeo_cost", "numbeo_indices", "avg_climate"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"include": err.Error()})
		return
	}

	costEnabled := include.Has("numbeo_cost")
	indicesEnabled := include.Has("numbeo_indices")
	climateEnabled := include.Has("avg_climate")

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
