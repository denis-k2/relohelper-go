package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	router.Get("/healthcheck", app.healthcheckHandler)

	router.Get("/cities", app.listCitiesHandler)
	router.Get("/cities/{id}", app.requireActivatedUser(app.showCityHandler))
	router.Get("/countries", app.listCountriesHandler)
	router.Get("/countries/{alpha3}", app.requireActivatedUser(app.showCountryHandler))

	router.Post("/users", app.registerUserHandler)
	router.Put("/users/activated", app.activateUserHandler)
	router.Post("/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Get("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	})

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
