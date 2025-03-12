package main

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/validator"
)

func (app *application) showCountriesHandler(w http.ResponseWriter, r *http.Request) {
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
	var input data.Filters
	v := validator.New()
	input.CountryCode = httprouter.ParamsFromContext(r.Context()).ByName("alpha3")
	if data.ValidateFilters(v, input); !v.Valid() {
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

	env := envelope{"country": country}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
