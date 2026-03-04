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

	include, err := parseInclude(qs, newIncludeSet("country", "numbeo_cost", "numbeo_indices", "avg_climate"))
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"include": err.Error()})
		return
	}

	ids, idsPresent, err := parseIDsInt64(qs, "ids", app.config.batch.maxIDs)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"ids": err.Error()})
		return
	}
	if idsPresent {
		if hasDetailedCityInclude(include) && len(ids) > 20 {
			app.failedValidationResponse(w, r, map[string]string{
				"ids": "ids cannot contain more than 20 unique values when detailed include blocks are requested",
			})
			return
		}

		cities, err := app.models.Cities.GetCitiesByIDs(ids, include)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		resp := make([]cityResponse, 0, len(cities))
		for _, city := range cities {
			resp = append(resp, newCityResponse(city, include))
		}

		err = app.writeJSON(w, http.StatusOK, envelope{"cities": resp}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if hasDetailedCityInclude(include) {
		app.failedValidationResponse(w, r, map[string]string{
			"include": "detailed include blocks are supported only for batch ids requests",
		})
		return
	}

	if qs.Has("country_code") {
		input.CountryCode = app.readString(qs, "country_code", "")
		if data.ValidateFilters(v, input); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
	}

	cities, err := app.models.Cities.ListCities(input.CountryCode, include)
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

func hasDetailedCityInclude(include data.IncludeSet) bool {
	return include.Has("numbeo_cost") || include.Has("numbeo_indices") || include.Has("avg_climate")
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
	include["country"] = struct{}{}

	city, err := app.models.Cities.GetCity(id, include)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	env := envelope{"city": newCityResponse(city, include)}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
