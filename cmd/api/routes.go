package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.Use(app.recoverPanic)
	router.Use(app.enableCORS)
	router.Use(app.rateLimit)
	router.Use(app.authenticate)

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	router.Get("/healthcheck", app.healthcheckHandler)

	router.Get("/cities", app.listCitiesHandler)
	router.With(app.requireActivatedUser).Get("/cities/{id}", app.showCityHandler)
	router.Get("/countries", app.listCountriesHandler)
	router.With(app.requireActivatedUser).Get("/countries/{alpha3}", app.showCountryHandler)

	router.Post("/users", app.registerUserHandler)
	router.Put("/users/activated", app.activateUserHandler)
	router.Post("/tokens/authentication", app.createAuthenticationTokenHandler)

	if app.config.env != "production" {
		router.Method(http.MethodGet, "/debug/vars", expvar.Handler())
	}

	return router
}
