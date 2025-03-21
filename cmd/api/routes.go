package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/cities", app.listCitiesHandler)
	router.HandlerFunc(http.MethodGet, "/cities/:id", app.showCityHandler)
	router.HandlerFunc(http.MethodGet, "/countries", app.listCountriesHandler)
	router.HandlerFunc(http.MethodGet, "/countries/:alpha3", app.showCountryHandler)

	router.HandlerFunc(http.MethodPost, "/users", app.registerUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}
