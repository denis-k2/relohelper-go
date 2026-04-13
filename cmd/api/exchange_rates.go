package main

import (
	"context"
	"net/http"
	"time"
)

func (app *application) exchangeRatesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
	defer cancel()

	resp, err := app.exchangeRates.Get(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"base":       resp.Base,
		"timestamp":  resp.Timestamp,
		"stale":      resp.Stale,
		"currencies": resp.Currencies,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
