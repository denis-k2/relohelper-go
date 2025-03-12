package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/cities", app.GetCities)
	router.HandlerFunc(http.MethodGet, "/cities/:id", app.GetCity)
	router.HandlerFunc(http.MethodGet, "/countries", app.showCountriesHandler)
	router.HandlerFunc(http.MethodGet, "/countries/:alpha3", app.showCountryHandler)

	return router
}
