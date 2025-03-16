package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/cities", app.listCitiesHandler)
	router.HandlerFunc(http.MethodGet, "/cities/:id", app.showCityHandler)
	router.HandlerFunc(http.MethodGet, "/countries", app.listCountriesHandler)
	router.HandlerFunc(http.MethodGet, "/countries/:alpha3", app.showCountryHandler)

	return router
}
