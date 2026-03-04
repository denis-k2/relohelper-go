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

	err := validateAllowedQueryParams(qs, newIncludeSet("country_codes", "include"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	codes, countryCodesPresent, err := parseIDsString(qs, "country_codes", 20)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"country_codes": err.Error()})
		return
	}

	include, err := parseInclude(qs, newIncludeSet("numbeo_indices", "legatum_indices"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"include": err.Error()})
		return
	}

	if countryCodesPresent {
		countries, err := app.models.Countries.GetCountriesByCodes(codes, include)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		resp := make([]countryResponse, 0, len(countries))
		for _, country := range countries {
			resp = append(resp, newCountryResponse(country, include))
		}

		err = app.writeJSON(w, http.StatusOK, envelope{"countries": resp}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if qs.Has("include") {
		app.failedValidationResponse(w, r, map[string]string{"include": "include is supported only for countries batch endpoint with country_codes"})
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
