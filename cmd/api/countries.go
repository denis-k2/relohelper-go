package main

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/validator"
)

func (app *application) listCountriesHandler(w http.ResponseWriter, r *http.Request) {
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
