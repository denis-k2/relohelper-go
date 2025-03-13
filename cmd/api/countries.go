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
	var input struct {
		data.Filters
		numbeoCountryIndices string
		legatumIndices       string
	}
	v := validator.New()
	qs := r.URL.Query()

	input.CountryCode = httprouter.ParamsFromContext(r.Context()).ByName("alpha3")
	input.numbeoCountryIndices = app.readString(qs, "numbeo_indices", "")
	input.legatumIndices = app.readString(qs, "legatum_indices", "")
	numbeoIndicesEnabled := data.ValidateBoolQuery(v, input.numbeoCountryIndices)
	legatumIndicesEnabled := data.ValidateBoolQuery(v, input.legatumIndices)
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
		numbeoIndicies, err := app.models.Countries.GetNumbeoCountryIndicies(input.CountryCode)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		country.NumbeoIndices = *numbeoIndicies
	}

	if legatumIndicesEnabled {
		// TODO: Implement handling for the 'numbeo_indices' query parameter
		app.logger.Warn("Handling 'legatum_indices' query parameter is incomplete.")
	}

	env := envelope{"country": country}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
