package main

import (
	"context"
	"net/http"
	"time"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	err := validateAllowedQueryParams(r.URL.Query(), newIncludeSet())
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) readinessHandler(w http.ResponseWriter, r *http.Request) {
	err := validateAllowedQueryParams(r.URL.Query(), newIncludeSet())
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{"query": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := app.db.PingContext(ctx); err != nil {
		env := envelope{
			"status": "not ready",
			"checks": map[string]string{
				"database": "unavailable",
			},
		}

		if writeErr := app.writeJSON(w, http.StatusServiceUnavailable, env, nil); writeErr != nil {
			app.serverErrorResponse(w, r, writeErr)
		}

		return
	}

	env := envelope{
		"status": "ready",
		"checks": map[string]string{
			"database": "available",
		},
	}

	if err := app.writeJSON(w, http.StatusOK, env, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
