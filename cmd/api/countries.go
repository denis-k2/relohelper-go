package main

import (
	"net/http"
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
