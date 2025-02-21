package main

import (
	"net/http"
	"testing"

	"github.com/denis-k2/relohelper-go/internal/assert"
	"github.com/denis-k2/relohelper-go/internal/data"
)

// TestHealthcheck tests the "/healthcheck" endpoint.
func TestHealthcheck(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/healthcheck")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got envelope
	unmarshalJSON(t, body, &got)
	assert.Equal(t, got["status"], "available")
}

// TestCities tests the “/cities” endpoint.
func TestCities(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/cities")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")

	type citiesResponse struct {
		Status string      `json:"status"`
		Cities []data.City `json:"cities"`
	}

	var got citiesResponse
	unmarshalJSON(t, body, &got)
	assert.Equal(t, got.Status, "available")
	assert.Equal(t, 534, len(got.Cities))

	wantCities := []data.City{
		{
			CityID:      11,
			City:        "New York",
			StateCode:   ptrString("US-NY"),
			CountryCode: "USA",
		},
		{
			CityID:      94,
			City:        "Toronto",
			StateCode:   nil,
			CountryCode: "CAN",
		},
		{
			CityID:      464,
			City:        "Moscow",
			StateCode:   nil,
			CountryCode: "RUS",
		},
	}
	for _, city := range wantCities {
		assert.DeepEqual(t, got.Cities[city.CityID-1], city)
	}
}

// TestCity tests the “/cities/:id” endpoint.
func TestCityID(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody data.City
	}{
		{
			name:     "Valid ID 1",
			urlPath:  "/cities/15",
			wantCode: http.StatusOK,
			wantBody: data.City{
				CityID:      15,
				City:        "Seattle",
				StateCode:   ptrString("US-WA"),
				CountryCode: "USA",
			},
		},
		{
			name:     "Valid ID 2",
			urlPath:  "/cities/273",
			wantCode: http.StatusOK,
			wantBody: data.City{
				CityID:      273,
				City:        "Tokyo",
				StateCode:   nil,
				CountryCode: "JPN",
			},
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/cities/777",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/cities/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/cities/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/cities/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/cities/",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)

			assert.Equal(t, statusCode, tt.wantCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			if tt.wantBody != (data.City{}) {
				type cityResponse struct {
					City data.City `json:"city"`
				}
				var got cityResponse
				unmarshalJSON(t, body, &got)
				assert.DeepEqual(t, got.City, tt.wantBody)
			}
		})
	}
}
