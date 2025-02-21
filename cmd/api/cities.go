package main

import (
	"errors"
	"net/http"

	"github.com/denis-k2/relohelper-go/internal/data"
)

func (app *application) GetCities(w http.ResponseWriter, r *http.Request) {
	cities, err := app.models.Cities.GetCityList()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
		"cities": cities,
	}

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
