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

	env := envelope{"city": city}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
