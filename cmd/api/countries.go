package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/validator"
)

func (app *application) listCountriesHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	err := validateAllowedQueryParams(qs, newIncludeSet("country_codes", "include"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	if qs.Has("include") {
		app.failedValidationResponse(w, r, map[string]string{"include": "include is not supported for countries list endpoint"})
		return
	}

	_, idsPresent, err := parseIDsString(qs, "country_codes", 100)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"country_codes": err.Error()})
		return
	}
	if idsPresent {
		app.errorResponse(w, r, http.StatusNotImplemented, "batch retrieval for countries by country_codes is not implemented yet")
		return
	}

	countries, err := app.models.Countries.GetCountryList()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"countries": countries}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCountryHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	err := validateAllowedQueryParams(qs, newIncludeSet("include"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	input.CountryCode = chi.URLParam(r, "alpha3")
	include, err := parseInclude(qs, newIncludeSet("numbeo_indices", "legatum_indices"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"include": err.Error()})
		return
	}

	numbeoIndicesEnabled := include.Has("numbeo_indices")
	legatumIndicesEnabled := include.Has("legatum_indices")

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	country, err := app.models.Countries.GetCountry(input.CountryCode)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if numbeoIndicesEnabled {
		numbeoIndices, err := app.models.Countries.GetNumbeoCountryIndicies(input.CountryCode)
		switch {
		case err == nil:
			country.NumbeoIndices = numbeoIndices
		case errors.Is(err, data.ErrRecordNotFound):
			country.NumbeoIndices = nil
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if legatumIndicesEnabled {
		legatumIndices, err := app.models.Countries.GetLegatumIndicies(input.CountryCode)
		switch {
		case err == nil:
			country.LegatumIndices = legatumIndices
		case errors.Is(err, data.ErrRecordNotFound):
			country.LegatumIndices = nil
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	env := envelope{"country": country}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
