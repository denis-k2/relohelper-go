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

	var got gotResponse
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
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Non-alphabetic code (123)",
			countryCode:  "123",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Empty country code",
			countryCode:  "",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with 1 letter (a)",
			countryCode:  "a",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with 4 letters (usaa)",
			countryCode:  "usaa",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with whitespace",
			countryCode:  url.QueryEscape(" us "),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with special characters (#$%)",
			countryCode:  url.QueryEscape("#$%"),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "SQL injection attempt",
			countryCode:  url.QueryEscape("usa'; DROP TABLE cities;--"),
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/cities?country_code=%s", tt.countryCode)
			statusCode, header, body := ts.get(t, url)
			assert.Equal(t, statusCode, tt.expectedCode)

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.Equal(t, header.Get("content-type"), "application/json")
			assert.Equal(t, len(got.Cities), tt.expectedCnt)

			switch tt.expectedCode {
			case http.StatusUnprocessableEntity:
				expectedError := map[string]any{
					"country_code": "must be exactly three English letters",
				}
				assert.DeepEqual(t, got.Error, expectedError)
			case http.StatusNotFound:
				assert.Equal(t, got.Error, "the requested resource could not be found")
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
			name:     "Valid ID (No query params)",
			urlPath:  "/cities/15",
			wantCode: http.StatusOK,
			wantBody: data.City{
				CityID:      15,
				City:        "Seattle",
				StateCode:   ptrString("US-WA"),
				CountryCode: "USA",
				Country:     "United States of America",
			},
		},
		{
			name:     "Valid ID with False & extra query params",
			urlPath:  "/cities/273?numbeo_cost=false&numbeo_indices=0&avg_climate=&extra_param=true",
			wantCode: http.StatusOK,
			wantBody: data.City{
				CityID:      273,
				City:        "Tokyo",
				StateCode:   nil,
				CountryCode: "JPN",
				Country:     "Japan",
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

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, got.City, tt.wantBody)

			if tt.wantCode == http.StatusNotFound {
				assert.Equal(t, got.Error, "the requested resource could not be found")
			}
		})
	}
}

// TestCities tests the “/cities/:id” endpoint with query parameters.
func TestCityIDandQuery(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name        string
		urlPath     string
		wantCode    int
		queryParams QueryParams
	}{
		{
			name:     "One param enabled",
			urlPath:  "/cities/12?numbeo_cost=true",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    true,
				indicesEnabled: false,
				climateEnabled: false,
			},
		},
		{
			name:     "Two params enabled",
			urlPath:  "/cities/123?numbeo_cost=1&numbeo_indices=TRUE&avg_climate=f",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: false,
			},
		},
		{
			name:     "All params enabled",
			urlPath:  "/cities/456?numbeo_cost=t&numbeo_indices=1&avg_climate=True",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: true,
			},
		},
		{
			name:     "Enable both params with missing Avg Climate data",
			urlPath:  "/cities/329?numbeo_cost=t&numbeo_indices=1&avg_climate=t",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: false,
			},
		},
		{
			name:     "One param with false value",
			urlPath:  "/cities/321?numbeo_indices=0",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    false,
				indicesEnabled: false,
				climateEnabled: false,
			},
		},
		{
			name:     "Unknown params (mixed cases)",
			urlPath:  "/cities/123?NUMBEO_COST=1&Numbeo_Indices=true&InvalidParam=TRUE",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    false, // Upper case parameter not recognized
				indicesEnabled: false, // CamelCase parameter not recognized
				climateEnabled: false,
			},
		},
		{
			name:     "Duplicate params",
			urlPath:  "/cities/234?numbeo_cost=false&numbeo_cost=true&avg_climate=1",
			wantCode: http.StatusOK,
			queryParams: QueryParams{
				costEnabled:    false,
				indicesEnabled: false,
				climateEnabled: true,
			},
		},
		{
			name:     "Unprocessable query value (123)",
			urlPath:  "/cities/100?numbeo_cost=123",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Unprocessable query value (abc)",
			urlPath:  "/cities/100?numbeo_cost=abc",
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.wantCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, cityFildsToBool(got.City), tt.queryParams)

			if tt.wantCode == http.StatusUnprocessableEntity {
				expectedError := map[string]any{
					"query parameter": "must be a boolean value",
				}
				assert.DeepEqual(t, got.Error, expectedError)
			}
		})
	}
}

// TestCountries tests the “/countries" endpoint.
func TestCountries(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/countries")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got gotResponse
	unmarshalJSON(t, body, &got)
	assert.Equal(t, len(got.Countries), 249)

	tests := []struct {
		index   int
		country data.Country
	}{
		{
			index: 14,
			country: data.Country{
				Code: "AUS",
				Name: "Australia",
			},
		},
		{
			index: 111,
			country: data.Country{
				Code: "ITA",
				Name: "Italy",
			},
		},
		{
			index: 218,
			country: data.Country{
				Code: "THA",
				Name: "Thailand",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Check country code=%s", tt.country.Code), func(t *testing.T) {
			assert.DeepEqual(t, got.Countries[tt.index], tt.country)
		})
	}
}

// TestCountry tests the “/countries/:alpha3" endpoint.
func TestCountry(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name         string
		urlPath      string
		expectedCode int
		country      data.Country
	}{
		{
			name:         "Valid uppercase code (AUS)",
			urlPath:      "/countries/AUS",
			expectedCode: http.StatusOK,
			country: data.Country{
				Code: "AUS",
				Name: "Australia",
			},
		},
		{
			name:         "Valid mixed case code (iTa)",
			urlPath:      "/countries/iTa",
			expectedCode: http.StatusOK,
			country: data.Country{
				Code: "ITA",
				Name: "Italy",
			},
		},
		{
			name:         "Valid lowercase code (tha)",
			urlPath:      "/countries/tha",
			expectedCode: http.StatusOK,
			country: data.Country{
				Code: "THA",
				Name: "Thailand",
			},
		},
		{
			name:         "Nonexistent country code (XXX)",
			urlPath:      "/countries/XXX",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Non-alphabetic code (123)",
			urlPath:      "/countries/123",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Empty country code",
			urlPath:      "/countries/",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Code with 1 letter (a)",
			urlPath:      "/countries/a",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with 4 letters (usaa)",
			urlPath:      "/countries/usaa",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with whitespace",
			urlPath:      "/countries/ us ",
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Code with special characters (#$%)",
			urlPath:      "/countries/" + url.QueryEscape("#$%"),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "SQL injection attempt",
			urlPath:      "/countries/" + url.QueryEscape("usa'; DROP TABLE cities;--"),
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.expectedCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, got.Country, tt.country)

			switch tt.expectedCode {
			case http.StatusUnprocessableEntity:
				expectedError := map[string]any{
					"country_code": "must be exactly three English letters",
				}
				assert.DeepEqual(t, got.Error, expectedError)
			case http.StatusNotFound:
				assert.Equal(t, got.Error, "the requested resource could not be found")
			}
		})
	}
}

// TestCountryandQuery tests the “/countries/:alpha3” endpoint with query parameters.
func TestCountryandQuery(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name        string
		urlPath     string
		wantCode    int
		queryParams QueryParamsCountry
	}{
		{
			name:     "Enable only Numbeo indices",
			urlPath:  "/countries/rus?numbeo_indices=true",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:     "Enable only Legatum indices",
			urlPath:  "/countries/usa?legatum_indices=TRUE",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:     "Enable all params (Numbeo and Legatum)",
			urlPath:  "/countries/bra?numbeo_indices=1&legatum_indices=True",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:     "Enable both params with missing Numbeo data",
			urlPath:  "/countries/afg?numbeo_indices=t&legatum_indices=true",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:     "Enable both params with missing both data",
			urlPath:  "/countries/wlf?numbeo_indices=t&legatum_indices=t",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:     "Disable only Numbeo indices",
			urlPath:  "/countries/can?numbeo_indices=0",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:     "Unknown params (mixed cases)",
			urlPath:  "/countries/arg?Numbeo_Indices=true&LEGATHUM_INDICES=1&InvalidParam=TRUE",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  false, // CamelCase parameter not recognized
				legatumIndicesEnabled: false, // Upper case parameter not recognized
			},
		},
		{
			name:     "Duplicate params",
			urlPath:  "/countries/chn?numbeo_indices=false&numbeo_indices=true&legatum_indices=1",
			wantCode: http.StatusOK,
			queryParams: QueryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:     "Unprocessable query value (123)",
			urlPath:  "/countries/deu?numbeo_indices=123",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Unprocessable query value (abc)",
			urlPath:  "/countries/nld?numbeo_indices=abc",
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.wantCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, countryFildsToBool(got.Country), tt.queryParams)

			if tt.wantCode == http.StatusUnprocessableEntity {
				expectedError := map[string]any{
					"query parameter": "must be a boolean value",
				}
				assert.DeepEqual(t, got.Error, expectedError)
			}
		})
	}
}
