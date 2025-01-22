package main

import (
	"net/http"
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
