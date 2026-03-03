package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/validator"
)

func (app *application) listCountriesHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	err := validateAllowedQueryParams(qs, newIncludeSet("ids", "include"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	if qs.Has("include") {
		app.failedValidationResponse(w, r, map[string]string{"include": "include is not supported for countries list endpoint"})
		return
	}

	codes, idsPresent, err := parseIDsString(qs, "ids", app.config.batch.maxIDs)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"ids": err.Error()})
		return
	}
	if idsPresent {
		countries, err := app.models.Countries.GetCountriesByCodes(codes)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		err = app.writeJSON(w, http.StatusOK, envelope{"countries": countries}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	countries, err := app.models.Countries.ListCountries()
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

	input.CountryCode = strings.ToUpper(chi.URLParam(r, "alpha3"))
	include, err := parseInclude(qs, newIncludeSet("numbeo_indices", "legatum_indices"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"include": err.Error()})
		return
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	country, err := app.models.Countries.GetCountry(input.CountryCode, include)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	env := envelope{"country": newCountryResponse(country, include)}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
