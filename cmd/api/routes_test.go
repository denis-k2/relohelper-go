package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/denis-k2/relohelper-go/internal/assert"
	"github.com/denis-k2/relohelper-go/internal/data"
)

// TestHealthcheck tests the "/healthcheck" endpoint.
func TestHealthcheck(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/healthcheck")
	assert.Equal(t, statusCode, http.StatusOK)
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
		Status      string      `json:"status"`
		Cities      []data.City `json:"cities"`
		CountryCode string      `json:"country_code"`
	}

	var got citiesResponse
	unmarshalJSON(t, body, &got)
	assert.Equal(t, len(got.Cities), 534)

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

// TestCities tests the “/cities” endpoint with query parameter.
func TestCitiesByCountry(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name         string
		countryCode  string
		expectedCnt  int
		expectedCode int
	}{
		{
			name:         "Valid uppercase code (USA)",
			countryCode:  "USA",
			expectedCnt:  58,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid mixed case code (cAn)",
			countryCode:  "cAn",
			expectedCnt:  29,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid lowercase code (rus)",
			countryCode:  "rus",
			expectedCnt:  8,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Nonexistent country code (XXX)",
			countryCode:  "xxx",
			expectedCnt:  0,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Non-alphabetic code (123)",
			countryCode:  "123",
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Empty country code",
			countryCode:  "",
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with 1 letter (a)",
			countryCode:  "a",
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with 4 letters (usaa)",
			countryCode:  "usaa",
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with whitespace",
			countryCode:  url.QueryEscape(" us "),
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with special characters (#$%)",
			countryCode:  url.QueryEscape("#$%"),
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "SQL injection attempt",
			countryCode:  url.QueryEscape("usa'; DROP TABLE cities;--"),
			expectedCnt:  0,
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/cities?country_code=%s", tt.countryCode)
			statusCode, header, body := ts.get(t, url)
			assert.Equal(t, statusCode, tt.expectedCode)

			var got struct {
				Cities      []data.City       `json:"cities"`
				CountryCode string            `json:"country_code"`
				Error       map[string]string `json:"error"`
			}
			unmarshalJSON(t, body, &got)
			assert.Equal(t, header.Get("content-type"), "application/json")
			assert.Equal(t, len(got.Cities), tt.expectedCnt)

			if tt.expectedCode != http.StatusOK {
				errorMsg := map[string]string{
					"country_code": "must be exactly three English letters",
				}
				assert.DeepEqual(t, got.Error, errorMsg)
			}
		})
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
