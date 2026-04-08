package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.Use(app.requestID)
	router.Use(app.collectPrometheusMetrics)
	router.Use(app.logRequest)
	router.Use(app.recoverPanic)
	router.Use(app.enableCORS)
	router.Use(app.rateLimit)
	if app.config.auth.enabled {
		router.Use(app.authenticate)
	}

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	router.Get("/", app.dashboardHandler)
	router.Get("/app.js", app.dashboardAppJSHandler)
	router.Get("/styles.css", app.dashboardStylesHandler)
	router.Get("/healthcheck", app.healthcheckHandler)
	router.Get("/readyz", app.readinessHandler)
	if app.config.metrics.port == 0 {
		router.Method(http.MethodGet, "/metrics", promhttp.Handler())
	}

	router.Get("/cities", app.listCitiesHandler)
	router.Get("/countries", app.listCountriesHandler)
	if app.config.auth.enabled {
		router.With(app.requireActivatedUser).Get("/cities/{id}", app.showCityHandler)
		router.With(app.requireActivatedUser).Get("/countries/{alpha3}", app.showCountryHandler)
	} else {
		router.Get("/cities/{id}", app.showCityHandler)
		router.Get("/countries/{alpha3}", app.showCountryHandler)
	}

	router.Post("/users", app.registerUserHandler)
	router.Put("/users/activated", app.activateUserHandler)
	router.Post("/tokens/authentication", app.createAuthenticationTokenHandler)

	if app.config.env != "production" {
		router.Method(http.MethodGet, "/debug/vars", expvar.Handler())
	}
	router.Get("/swagger", app.swaggerUIHandler)
	router.Get("/swagger/openapi.yaml", app.openAPISpecHandler)

	return router
}

func (app *application) metricsRoutes() http.Handler {
	router := chi.NewRouter()
	router.Method(http.MethodGet, "/metrics", promhttp.Handler())
	return router
}
