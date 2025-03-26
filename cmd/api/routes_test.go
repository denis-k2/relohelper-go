package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/denis-k2/relohelper-go/internal/assert"
	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/mailer/mocks"
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
		name        string
		countryCode string
		citiesCount int
		statusCode  int
	}{
		{
			name:        "Valid uppercase code (USA)",
			countryCode: "USA",
			citiesCount: 58,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Valid mixed case code (cAn)",
			countryCode: "cAn",
			citiesCount: 29,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Valid lowercase code (rus)",
			countryCode: "rus",
			citiesCount: 8,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Nonexistent country code (XXX)",
			countryCode: "xxx",
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Non-alphabetic code (123)",
			countryCode: "123",
			statusCode:  http.StatusUnprocessableEntity,
		},
		{
			name:        "Empty country code",
			countryCode: "",
			statusCode:  http.StatusUnprocessableEntity,
		},
		{
			name:        "Code with 1 letter (a)",
			countryCode: "a",
			statusCode:  http.StatusUnprocessableEntity,
		},
		{
			name:        "Code with 4 letters (usaa)",
			countryCode: "usaa",
			statusCode:  http.StatusUnprocessableEntity,
		},
		{
			name:        "Code with whitespace",
			countryCode: url.QueryEscape(" us "),
			statusCode:  http.StatusUnprocessableEntity,
		},
		{
			name:        "Code with special characters (#$%)",
			countryCode: url.QueryEscape("#$%"),
			statusCode:  http.StatusUnprocessableEntity,
		},
		{
			name:        "SQL injection attempt",
			countryCode: url.QueryEscape("usa'; DROP TABLE city;"),
			statusCode:  http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/cities?country_code=%s", tt.countryCode)
			statusCode, header, body := ts.get(t, url)
			assert.Equal(t, statusCode, tt.statusCode)

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.Equal(t, header.Get("content-type"), "application/json")
			assert.Equal(t, len(got.Cities), tt.citiesCount)

			switch tt.statusCode {
			case http.StatusUnprocessableEntity:
				wantError := map[string]any{
					"country_code": "must be exactly three English letters",
				}
				assert.DeepEqual(t, got.Error, wantError)
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
		name       string
		urlPath    string
		statusCode int
		city       data.City
	}{
		{
			name:       "Valid ID (No query params)",
			urlPath:    "/cities/15",
			statusCode: http.StatusOK,
			city: data.City{
				CityID:      15,
				City:        "Seattle",
				StateCode:   ptrString("US-WA"),
				CountryCode: "USA",
				Country:     "United States of America",
			},
		},
		{
			name:       "Valid ID with False & extra query params",
			urlPath:    "/cities/273?numbeo_cost=false&numbeo_indices=0&avg_climate=&extra_param=true",
			statusCode: http.StatusOK,
			city: data.City{
				CityID:      273,
				City:        "Tokyo",
				StateCode:   nil,
				CountryCode: "JPN",
				Country:     "Japan",
			},
		},
		{
			name:       "Non-existent ID",
			urlPath:    "/cities/777",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Negative ID",
			urlPath:    "/cities/-1",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Decimal ID",
			urlPath:    "/cities/1.23",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "String ID",
			urlPath:    "/cities/foo",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Empty ID",
			urlPath:    "/cities/",
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, got.City, tt.city)

			if tt.statusCode == http.StatusNotFound {
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
		statusCode  int
		queryParams queryParamsCity
	}{
		{
			name:       "One param enabled",
			urlPath:    "/cities/12?numbeo_cost=true",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: false,
				climateEnabled: false,
			},
		},
		{
			name:       "Two params enabled",
			urlPath:    "/cities/123?numbeo_cost=1&numbeo_indices=TRUE&avg_climate=f",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: false,
			},
		},
		{
			name:       "All params enabled",
			urlPath:    "/cities/456?numbeo_cost=t&numbeo_indices=1&avg_climate=True",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: true,
			},
		},
		{
			name:       "Enable both params with missing Avg Climate data",
			urlPath:    "/cities/329?numbeo_cost=t&numbeo_indices=1&avg_climate=t",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: false,
			},
		},
		{
			name:       "One param with false value",
			urlPath:    "/cities/321?numbeo_indices=0",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    false,
				indicesEnabled: false,
				climateEnabled: false,
			},
		},
		{
			name:       "Unknown params (mixed cases)",
			urlPath:    "/cities/123?NUMBEO_COST=1&Numbeo_Indices=true&InvalidParam=TRUE",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    false, // Upper case parameter not recognized
				indicesEnabled: false, // CamelCase parameter not recognized
				climateEnabled: false,
			},
		},
		{
			name:       "Duplicate params",
			urlPath:    "/cities/234?numbeo_cost=false&numbeo_cost=true&avg_climate=1",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    false,
				indicesEnabled: false,
				climateEnabled: true,
			},
		},
		{
			name:       "Unprocessable query value (123)",
			urlPath:    "/cities/100?numbeo_cost=123",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Unprocessable query value (abc)",
			urlPath:    "/cities/100?numbeo_cost=abc",
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, cityFildsToBool(got.City), tt.queryParams)

			if tt.statusCode == http.StatusUnprocessableEntity {
				wantError := map[string]any{
					"query parameter": "must be a boolean value",
				}
				assert.DeepEqual(t, got.Error, wantError)
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
		name       string
		urlPath    string
		statusCode int
		country    data.Country
	}{
		{
			name:       "Valid uppercase code (AUS)",
			urlPath:    "/countries/AUS",
			statusCode: http.StatusOK,
			country: data.Country{
				Code: "AUS",
				Name: "Australia",
			},
		},
		{
			name:       "Valid mixed case code (iTa)",
			urlPath:    "/countries/iTa",
			statusCode: http.StatusOK,
			country: data.Country{
				Code: "ITA",
				Name: "Italy",
			},
		},
		{
			name:       "Valid lowercase code (tha)",
			urlPath:    "/countries/tha",
			statusCode: http.StatusOK,
			country: data.Country{
				Code: "THA",
				Name: "Thailand",
			},
		},
		{
			name:       "Nonexistent country code (XXX)",
			urlPath:    "/countries/XXX",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Non-alphabetic code (123)",
			urlPath:    "/countries/123",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Empty country code",
			urlPath:    "/countries/",
			statusCode: http.StatusOK,
		},
		{
			name:       "Code with 1 letter (a)",
			urlPath:    "/countries/a",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Code with 4 letters (usaa)",
			urlPath:    "/countries/usaa",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Code with whitespace",
			urlPath:    "/countries/ us ",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Code with special characters (#$%)",
			urlPath:    "/countries/" + url.QueryEscape("#$%"),
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "SQL injection attempt",
			urlPath:    "/countries/" + url.QueryEscape("usa'; DROP TABLE country;"),
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, got.Country, tt.country)

			switch tt.statusCode {
			case http.StatusUnprocessableEntity:
				wantError := map[string]any{
					"country_code": "must be exactly three English letters",
				}
				assert.DeepEqual(t, got.Error, wantError)
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
		statusCode  int
		queryParams queryParamsCountry
	}{
		{
			name:       "Enable only Numbeo indices",
			urlPath:    "/countries/rus?numbeo_indices=true",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:       "Enable only Legatum indices",
			urlPath:    "/countries/usa?legatum_indices=TRUE",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Enable all params (Numbeo and Legatum)",
			urlPath:    "/countries/bra?numbeo_indices=1&legatum_indices=True",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Enable both params with missing Numbeo data",
			urlPath:    "/countries/afg?numbeo_indices=t&legatum_indices=true",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Enable both params with missing both data",
			urlPath:    "/countries/wlf?numbeo_indices=t&legatum_indices=t",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:       "Disable only Numbeo indices",
			urlPath:    "/countries/can?numbeo_indices=0",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:       "Unknown params (mixed cases)",
			urlPath:    "/countries/arg?Numbeo_Indices=true&LEGATHUM_INDICES=1&InvalidParam=TRUE",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false, // CamelCase parameter not recognized
				legatumIndicesEnabled: false, // Upper case parameter not recognized
			},
		},
		{
			name:       "Duplicate params",
			urlPath:    "/countries/chn?numbeo_indices=false&numbeo_indices=true&legatum_indices=1",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Unprocessable query value (123)",
			urlPath:    "/countries/deu?numbeo_indices=123",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Unprocessable query value (abc)",
			urlPath:    "/countries/nld?numbeo_indices=abc",
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.get(t, tt.urlPath)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, countryFildsToBool(got.Country), tt.queryParams)

			if tt.statusCode == http.StatusUnprocessableEntity {
				wantError := map[string]any{
					"query parameter": "must be a boolean value",
				}
				assert.DeepEqual(t, got.Error, wantError)
			}
		})
	}
}

// TestErrorHandling tests handle 404 Not Found and 405 Method Not Allowed errors.
func TestErrorHandling(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		method     string
		urlPath    string
		statusCode int
		errMessage string
	}{
		{
			name:       "GET to non-existent endpoint",
			method:     "GET",
			urlPath:    "/invalidpath",
			statusCode: http.StatusNotFound,
			errMessage: "the requested resource could not be found",
		},
		{
			name:       "POST to non-existent endpoint",
			method:     "POST",
			urlPath:    "/cities/123/invalidpath",
			statusCode: http.StatusNotFound,
			errMessage: "the requested resource could not be found",
		},
		{
			name:       "Unsupported method PUT to existent endpoint",
			method:     "PUT",
			urlPath:    "/cities",
			statusCode: http.StatusMethodNotAllowed,
			errMessage: "the PUT method is not supported for this resource",
		},
		{
			name:       "Unsupported method PATCH to existent endpoint",
			method:     "PATCH",
			urlPath:    "/countries/rus",
			statusCode: http.StatusMethodNotAllowed,
			errMessage: "the PATCH method is not supported for this resource",
		},
		{
			name:       "Unsupported method DELETE to existent endpoint",
			method:     "DELETE",
			urlPath:    "/cities/123",
			statusCode: http.StatusMethodNotAllowed,
			errMessage: "the DELETE method is not supported for this resource",
		},
		{
			name:       "Unsupported method GET for users registratoin endpoint",
			method:     "GET",
			urlPath:    "/users",
			statusCode: http.StatusMethodNotAllowed,
			errMessage: "the GET method is not supported for this resource",
		},
		{
			name:       "Unsupported method POST for users activation endpoint",
			method:     "POST",
			urlPath:    "/users/activated",
			statusCode: http.StatusMethodNotAllowed,
			errMessage: "the POST method is not supported for this resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, tt.method, tt.urlPath, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("Content-Type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, got.Error, tt.errMessage)
		})
	}
}

func TestRegisterUser(t *testing.T) {
	setupUsersTable(t)
	defer teardownUsersTable(t)
	setupTokensTable(t)
	defer teardownTokensTable(t)

	ts := newTestServer(testApp.routes())
	defer ts.Close()

	var (
		validName = "Bob"
		emptyName = ""
		longName  = strings.Repeat("a", 501)

		validPassword = "validPa$$word"
		emptyPassword = ""
		shortPassword = "pa$$"
		longPassword  = strings.Repeat("a", 73)

		validEmail   = "bob@example.com"
		invalidEmail = "bob@invalid."
		emptyEmail   = ""
	)

	tests := []struct {
		name       string
		payload    data.InputUser
		statusCode int
		errMessage map[string]any
	}{
		{
			name: "Valid submission",
			payload: data.InputUser{
				Name:          validName,
				Email:         validEmail,
				PlainPassword: validPassword,
			},
			statusCode: http.StatusAccepted,
			errMessage: nil,
		},
		{
			name: "User already exists (duplicate email)",
			payload: data.InputUser{
				Name:          validName,
				Email:         validEmail,
				PlainPassword: validPassword,
			},
			statusCode: http.StatusUnprocessableEntity,
			errMessage: map[string]any{
				"email": "a user with this email address already exists",
			},
		},
		{
			name: "Empty name, empty email, empty password",
			payload: data.InputUser{
				Name:          emptyName,
				Email:         emptyEmail,
				PlainPassword: emptyPassword,
			},
			statusCode: http.StatusUnprocessableEntity,
			errMessage: map[string]any{
				"email":    "must be provided",
				"name":     "must be provided",
				"password": "must be provided",
			},
		},
		{
			name: "Invalid email, short password",
			payload: data.InputUser{
				Name:          validName,
				Email:         invalidEmail,
				PlainPassword: shortPassword,
			},
			statusCode: http.StatusUnprocessableEntity,
			errMessage: map[string]any{
				"email":    "must be a valid email address",
				"password": "must be at least 8 bytes long",
			},
		},
		{
			name: "Long name, long password",
			payload: data.InputUser{
				Name:          longName,
				Email:         validEmail,
				PlainPassword: longPassword,
			},
			statusCode: http.StatusUnprocessableEntity,
			errMessage: map[string]any{
				"name":     "must not be more than 500 bytes long",
				"password": "must not be more than 72 bytes long",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, "POST", "/users", tt.payload)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)

			if tt.statusCode == http.StatusAccepted {
				assert.NotEmpty(t, got.User.ID)
				assert.NotEmpty(t, got.User.CreatedAt)
				assert.Equal(t, got.User.Name, tt.payload.Name)
				assert.Equal(t, got.User.Email, tt.payload.Email)
				assert.Equal(t, got.User.Activated, false)
			} else {
				assert.DeepEqual(t, got.Error, tt.errMessage)
			}
		})
	}
}

func TestActivateUser(t *testing.T) {
	setupUsersTable(t)
	defer teardownUsersTable(t)
	setupTokensTable(t)
	defer teardownTokensTable(t)

	ts := newTestServer(testApp.routes())
	defer ts.Close()

	inputUser := data.InputUser{
		Name:          "John Smith",
		Email:         "john@example.com",
		PlainPassword: "validPa55word",
	}

	statusCode, _, _ := ts.sendRequest(t, "POST", "/users", inputUser)
	assert.Equal(t, statusCode, http.StatusAccepted)

	mockMailer := testApp.mailer.(*mocks.MockMailer)
	plainBody := mockMailer.Email.PlainBody
	bodyMap, ok := plainBody.(map[string]any)
	if !ok {
		t.Fatal("plainBody is not a map[string]any")
	}
	token, exists := bodyMap["activationToken"]
	if !exists {
		t.Fatal("activationToken not found in plainBody")
	}

	tests := []struct {
		name         string
		tokenMessage map[string]any
		statusCode   int
		errorMessage map[string]any
		payload      data.InputUser
	}{
		{
			name:         "Valid activation",
			tokenMessage: map[string]any{"token": token},
			statusCode:   http.StatusOK,
			errorMessage: nil,
			payload:      inputUser,
		},
		{
			name:         "User already activated (duplicate activation)",
			tokenMessage: map[string]any{"token": token},
			statusCode:   http.StatusUnprocessableEntity,
			errorMessage: map[string]any{
				"token": "invalid or expired activation token",
			},
		},
		{
			name:         "Mismatched or expired activation token",
			tokenMessage: map[string]any{"token": "P4B3URJZJ2NW5UPZC2OHN4H2NM"},
			statusCode:   http.StatusUnprocessableEntity,
			errorMessage: map[string]any{
				"token": "invalid or expired activation token",
			},
		},
		{
			name:         "Short activation token",
			tokenMessage: map[string]any{"token": "P4B3URJZJ"},
			statusCode:   http.StatusUnprocessableEntity,
			errorMessage: map[string]any{
				"token": "must be 26 bytes long",
			},
		},
		{
			name:         "Long activation token",
			tokenMessage: map[string]any{"token": "P4B3URJZJ2NW5UPZC2OHN4H2NM11111"},
			statusCode:   http.StatusUnprocessableEntity,
			errorMessage: map[string]any{
				"token": "must be 26 bytes long",
			},
		},
		{
			name:         "Empty token messege",
			tokenMessage: map[string]any{},
			statusCode:   http.StatusUnprocessableEntity,
			errorMessage: map[string]any{
				"token": "must be provided",
			},
		},
		{
			name:         "Empty token value",
			tokenMessage: map[string]any{"token": ""},
			statusCode:   http.StatusUnprocessableEntity,
			errorMessage: map[string]any{
				"token": "must be provided",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, "PUT", "/users/activated", tt.tokenMessage)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)

			if tt.statusCode == http.StatusOK {
				assert.Equal(t, got.User.ID, 1)
				assert.Equal(t, got.User.Name, tt.payload.Name)
				assert.Equal(t, got.User.Email, tt.payload.Email)
				assert.Equal(t, got.User.Activated, true)
				assert.NotEmpty(t, got.User.CreatedAt)
			} else {
				assert.DeepEqual(t, got.Error, tt.errorMessage)
			}
		})
	}
}
