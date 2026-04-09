package main

import (
	"net/http"

	uiassets "github.com/denis-k2/relohelper-go/ui"
)

func (app *application) serveUIAsset(w http.ResponseWriter, r *http.Request, path string, contentType string) {
	asset, err := uiassets.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(asset)
}

func (app *application) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	app.serveUIAsset(w, r, "index.html", "text/html; charset=utf-8")
}

func (app *application) dashboardAppJSHandler(w http.ResponseWriter, r *http.Request) {
	app.serveUIAsset(w, r, "app.js", "application/javascript; charset=utf-8")
}

func (app *application) dashboardStylesHandler(w http.ResponseWriter, r *http.Request) {
	app.serveUIAsset(w, r, "styles.css", "text/css; charset=utf-8")
}

func (app *application) dashboardFaviconHandler(w http.ResponseWriter, r *http.Request) {
	app.serveUIAsset(w, r, "favicon.svg", "image/svg+xml")
}
