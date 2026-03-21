package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/denis-k2/relohelper-go/internal/assert"
	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/mocks"
	"github.com/prometheus/client_golang/prometheus/testutil"
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

func TestHealthcheckSetsRequestID(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, _ := ts.get(t, "/healthcheck")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("X-Request-ID") != "", true)
}

func TestHealthcheckEchoesIncomingRequestID(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, _ := ts.request(t, http.MethodGet, "/healthcheck", http.Header{
		"X-Request-ID": []string{"req-12345"},
	})
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("X-Request-ID"), "req-12345")
}

func TestReadyz(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/readyz")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got envelope
	unmarshalJSON(t, body, &got)
	assert.Equal(t, got["status"], "ready")

	checks, ok := got["checks"].(map[string]any)
	assert.Equal(t, ok, true)
	assert.Equal(t, checks["database"], "available")
}

func TestMetrics(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	_, _, _ = ts.get(t, "/healthcheck")
	_, _, _ = ts.get(t, "/readyz")

	rs, err := ts.Client().Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rs.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	metrics := string(body)
	assert.Equal(t, rs.StatusCode, http.StatusOK)
	assert.Equal(t, strings.Contains(metrics, "relohelper_http_requests_total"), true)
	assert.Equal(t, strings.Contains(metrics, "relohelper_http_requests_by_status_class_total"), true)
	assert.Equal(t, strings.Contains(metrics, "relohelper_http_request_duration_seconds"), true)
	assert.Equal(t, strings.Contains(metrics, `route="/healthcheck"`), true)
	assert.Equal(t, strings.Contains(metrics, `route="/readyz"`), true)
}

func TestMetricsHiddenOnMainRouterWhenDedicatedMetricsPortConfigured(t *testing.T) {
	appWithDedicatedMetrics := &application{
		config: testApp.config,
		logger: testApp.logger,
		db:     testApp.db,
		models: testApp.models,
		mailer: testApp.mailer,
	}
	appWithDedicatedMetrics.config.metrics.port = 4001

	ts := newTestServer(appWithDedicatedMetrics.routes())
	defer ts.Close()

	statusCode, _, _ := ts.get(t, "/metrics")
	assert.Equal(t, statusCode, http.StatusNotFound)
}

func TestRateLimiterRejectedMetric(t *testing.T) {
	cfg := testApp.config
	cfg.limiter.enabled = true
	cfg.limiter.rps = 0.0001
	cfg.limiter.burst = 1

	limitedApp := &application{
		config: cfg,
		logger: testApp.logger,
		db:     testApp.db,
		models: testApp.models,
		mailer: testApp.mailer,
	}

	ts := newTestServer(limitedApp.routes())
	defer ts.Close()

	beforeRejected := testutil.ToFloat64(rateLimiterRejectedMetric)
	beforeAllowed := testutil.ToFloat64(rateLimiterAllowedMetric)

	statusCode, _, _ := ts.get(t, "/healthcheck")
	assert.Equal(t, statusCode, http.StatusOK)

	statusCode, _, _ = ts.get(t, "/healthcheck")
	assert.Equal(t, statusCode, http.StatusTooManyRequests)

	afterRejected := testutil.ToFloat64(rateLimiterRejectedMetric)
	afterAllowed := testutil.ToFloat64(rateLimiterAllowedMetric)

	assert.Equal(t, afterRejected-beforeRejected, float64(1))
	assert.Equal(t, afterAllowed-beforeAllowed, float64(1))
}

func TestDBStatsProviderMetrics(t *testing.T) {
	assert.Equal(t, testutil.ToFloat64(dbMaxOpenConnectionsMetric) > 0, true)
	assert.Equal(t, testutil.ToFloat64(dbOpenConnectionsMetric) >= 0, true)
	assert.Equal(t, testutil.ToFloat64(dbInUseConnectionsMetric) >= 0, true)
	assert.Equal(t, testutil.ToFloat64(dbIdleConnectionsMetric) >= 0, true)
	assert.Equal(t, testutil.ToFloat64(dbWaitCountMetric) >= 0, true)
	assert.Equal(t, testutil.ToFloat64(dbWaitDurationMetric) >= 0, true)
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
	assert.Equal(t, len(got.Cities) > 0, true)

	for _, geonameID := range []int64{5128581, 6167865, 524901} {
		gotCity := findCityByID(t, got.Cities, geonameID)
		wantCity := fetchExpectedCity(t, geonameID)
		assert.DeepEqual(t, gotCity, wantCity)
	}
}

func findCityByID(t *testing.T, cities []data.City, id int64) data.City {
	t.Helper()

	for _, city := range cities {
		if city.GeonameID == id {
			return city
		}
	}

	t.Fatalf("city with id=%d not found in response", id)
	return data.City{}
}

func fetchExpectedCity(t *testing.T, geonameID int64) data.City {
	t.Helper()

	var city data.City
	err := testDB.QueryRow(`
		SELECT c.geoname_id, c.city, c.state_code, c.country_code, ctr.country,
		       c.population, c.latitude, c.longitude, c.timezone,
		       to_char(c.updated_date, 'YYYY-MM-DD') AS last_update
		FROM cities c
		LEFT JOIN countries ctr ON ctr.country_code = c.country_code
		WHERE c.geoname_id = $1;`, geonameID).Scan(
		&city.GeonameID,
		&city.Name,
		&city.StateCode,
		&city.CountryCode,
		&city.CountryName,
		&city.Population,
		&city.Latitude,
		&city.Longitude,
		&city.Timezone,
		&city.LastUpdate,
	)
	if err != nil {
		t.Fatalf("failed to fetch expected city %d: %v", geonameID, err)
	}

	return city
}

func fetchExpectedCountry(t *testing.T, code string) data.Country {
	t.Helper()

	var country data.Country
	err := testDB.QueryRow(`
		SELECT country_code, country, population, area, last_update::text
		FROM countries
		WHERE country_code = UPPER($1);`, code).Scan(
		&country.Code,
		&country.Name,
		&country.Population,
		&country.Area,
		&country.LastUpdate,
	)
	if err != nil {
		t.Fatalf("failed to fetch expected country %s: %v", code, err)
	}

	return country
}

// TestCities tests the “/cities” endpoint with query parameter.
func TestCitiesByCountry(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name        string
		countryCode string
		statusCode  int
	}{
		{
			name:        "Valid uppercase code (USA)",
			countryCode: "USA",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Valid mixed case code (cAn)",
			countryCode: "cAn",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Valid lowercase code (rus)",
			countryCode: "rus",
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
			countryCode: url.QueryEscape("usa'; DROP TABLE cities;"),
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

			switch tt.statusCode {
			case http.StatusOK:
				var expectedCount int
				err := testDB.QueryRow(`
					SELECT COUNT(*)
					FROM cities
					WHERE LOWER(country_code) = LOWER($1);`, tt.countryCode).Scan(&expectedCount)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, len(got.Cities), expectedCount)
				for _, city := range got.Cities {
					assert.Equal(t, strings.EqualFold(city.CountryCode, tt.countryCode), true)
				}
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

// TestCitiesBatchByIDs tests batch retrieval for "/cities?ids=...".
func TestCitiesBatchByIDs(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/cities?ids=5128581,6167865,5128581")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got gotResponse
	unmarshalJSON(t, body, &got)
	assert.Equal(t, len(got.Cities), 2)
	assert.Equal(t, got.Cities[0].GeonameID, int64(5128581))
	assert.Equal(t, got.Cities[0].CountryName, "United States of America")
	assert.Equal(t, got.Cities[1].GeonameID, int64(6167865))
	assert.Equal(t, got.Cities[1].CountryName, "Canada")
}

// TestCitiesBatchByIDsWithInclude tests include behavior for "/cities?ids=...&include=...".
func TestCitiesBatchByIDsWithInclude(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/cities?ids=3069011,5128581&include=avg_climate")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")
	assert.Equal(t, jsonArrayObjectHasKeyByID(body, "cities", "geoname_id", int64(3069011), "avg_climate"), true)
	assert.Equal(t, jsonArrayObjectIsNullByID(body, "cities", "geoname_id", int64(3069011), "avg_climate"), true)
}

// TestCitiesBatchByIDsRejectsCountryCode tests that batch ids mode rejects country_code.
func TestCitiesBatchByIDsRejectsCountryCode(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/cities?country_code=JPN&ids=5128581,2643743,2147714&include=avg_climate,numbeo_cost,numbeo_indices")
	assert.Equal(t, statusCode, http.StatusUnprocessableEntity)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got gotResponse
	unmarshalJSON(t, body, &got)
	assert.DeepEqual(t, got.Error, map[string]any{
		"query": "country_code cannot be used together with ids",
	})
}

// TestCitiesBatchByIDsDetailedIncludeLimit tests include batch limit for "/cities?ids=...&include=...".
func TestCitiesBatchByIDsDetailedIncludeLimit(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	rawIDs := make([]string, 0, 21)
	for i := 1; i <= 21; i++ {
		rawIDs = append(rawIDs, fmt.Sprintf("%d", i))
	}

	urlPath := fmt.Sprintf("/cities?ids=%s&include=numbeo_indices", strings.Join(rawIDs, ","))
	statusCode, header, body := ts.get(t, urlPath)
	assert.Equal(t, statusCode, http.StatusUnprocessableEntity)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got gotResponse
	unmarshalJSON(t, body, &got)
	assert.DeepEqual(t, got.Error, map[string]any{
		"ids": fmt.Sprintf("ids cannot contain more than %d unique values when detailed include blocks are requested", testApp.config.batch.maxDetailedIDs),
	})
}

// TestCity tests the “/cities/:id” endpoint.
func TestCityID(t *testing.T) {
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		urlPath    string
		statusCode int
		id         int64
	}{
		{
			name:       "Valid ID (No query params)",
			urlPath:    "/cities/5809844",
			statusCode: http.StatusOK,
			id:         5809844,
		},
		{
			name:       "Valid ID with include query",
			urlPath:    "/cities/1850147?include=country",
			statusCode: http.StatusOK,
			id:         1850147,
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
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, "GET", tt.urlPath, mocks.Headers, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			if tt.statusCode == http.StatusOK {
				assert.DeepEqual(t, got.City, fetchExpectedCity(t, tt.id))
			}

			if tt.statusCode == http.StatusNotFound {
				assert.Equal(t, got.Error, "the requested resource could not be found")
			}
		})
	}
}

// TestCities tests the “/cities/:id” endpoint with query parameters.
func TestCityIDandQuery(t *testing.T) {
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name        string
		urlPath     string
		statusCode  int
		queryParams queryParamsCity
	}{
		{
			name:       "One param enabled",
			urlPath:    "/cities/5378538?include=numbeo_cost",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: false,
				climateEnabled: false,
			},
		},
		{
			name:       "Two params enabled",
			urlPath:    "/cities/2562305?include=numbeo_cost,numbeo_indices",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: false,
			},
		},
		{
			name:       "All params enabled",
			urlPath:    "/cities/2542997?include=numbeo_cost,numbeo_indices,avg_climate",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: true,
				climateEnabled: true,
			},
		},
		{
			name:       "Include country only",
			urlPath:    "/cities/3871336?include=country",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    false,
				indicesEnabled: false,
				climateEnabled: false,
			},
		},
		{
			name:       "Unknown params (mixed cases)",
			urlPath:    "/cities/2562305?NUMBEO_COST=1&Numbeo_Indices=true&InvalidParam=TRUE",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Duplicate includes",
			urlPath:    "/cities/1850147?include=numbeo_cost,numbeo_cost,avg_climate",
			statusCode: http.StatusOK,
			queryParams: queryParamsCity{
				costEnabled:    true,
				indicesEnabled: false,
				climateEnabled: true,
			},
		},
		{
			name:       "Unprocessable include value (123)",
			urlPath:    "/cities/1850147?include=123",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Unprocessable include value (abc)",
			urlPath:    "/cities/1850147?include=abc",
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, "GET", tt.urlPath, mocks.Headers, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, cityFildsToBool(got.City), tt.queryParams)

			if tt.statusCode == http.StatusUnprocessableEntity {
				wantError := map[string]any{
					"include": "include contains unsupported value \"123\"",
				}
				if tt.name == "Unprocessable include value (abc)" {
					wantError = map[string]any{
						"include": "include contains unsupported value \"abc\"",
					}
				}
				if tt.name == "Unknown params (mixed cases)" {
					wantError = map[string]any{
						"query": "unknown query parameter \"InvalidParam\"",
					}
				}
				assert.DeepEqual(t, got.Error, wantError)
			}
		})
	}
}

// TestCityIncludeFieldPresence tests include-driven field presence/omission for "/cities/:id".
func TestCityIncludeFieldPresence(t *testing.T) {
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	findGeonameIDWithoutData := func(t *testing.T, table string, idColumn string) (int64, bool) {
		t.Helper()

		query := fmt.Sprintf(`
			SELECT c.geoname_id
			FROM cities c
			LEFT JOIN %s x ON x.%s = c.geoname_id
			WHERE x.%s IS NULL
			LIMIT 1;`, table, idColumn, idColumn)

		var geonameID int64
		if err := testDB.QueryRow(query).Scan(&geonameID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, false
			}
			t.Fatalf("failed to find city without data in %s: %v", table, err)
		}

		return geonameID, true
	}

	t.Run("avg_climate requested and absent => explicit null", func(t *testing.T) {
		statusCode, _, body := ts.sendRequest(t, "GET", "/cities/3069011?include=avg_climate,numbeo_indices", mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "city", "avg_climate"), true)
		assert.Equal(t, jsonIsNull(body, "city", "avg_climate"), true)
		assert.Equal(t, jsonHasKey(body, "city", "numbeo_indices"), true)
		assert.Equal(t, jsonIsNull(body, "city", "numbeo_indices"), false)
	})

	t.Run("avg_climate not requested => field omitted", func(t *testing.T) {
		statusCode, _, body := ts.sendRequest(t, "GET", "/cities/3069011", mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "city", "avg_climate"), false)
	})

	t.Run("numbeo_cost requested and absent => explicit null", func(t *testing.T) {
		geonameID, ok := findGeonameIDWithoutData(t, "numbeo_city_costs", "geoname_id")
		if !ok {
			t.Skip("no city without numbeo_city_costs in test dataset")
		}
		statusCode, _, body := ts.sendRequest(t, "GET", fmt.Sprintf("/cities/%d?include=numbeo_cost", geonameID), mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "city", "numbeo_cost"), true)
		assert.Equal(t, jsonIsNull(body, "city", "numbeo_cost"), true)
	})

	t.Run("numbeo_indices requested and absent => explicit null", func(t *testing.T) {
		geonameID, ok := findGeonameIDWithoutData(t, "numbeo_city_indices", "geoname_id")
		if !ok {
			t.Skip("no city without numbeo_city_indices in test dataset")
		}
		statusCode, _, body := ts.sendRequest(t, "GET", fmt.Sprintf("/cities/%d?include=numbeo_indices", geonameID), mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "city", "numbeo_indices"), true)
		assert.Equal(t, jsonIsNull(body, "city", "numbeo_indices"), true)
	})
}

// TestCountryIncludeFieldPresence tests include-driven field presence/omission for "/countries/:alpha3".
func TestCountryIncludeFieldPresence(t *testing.T) {
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	findCountryCodeWithoutData := func(t *testing.T, table string, codeColumn string) (string, bool) {
		t.Helper()

		query := fmt.Sprintf(`
			SELECT ctr.country_code
			FROM countries ctr
			LEFT JOIN %s x ON x.%s = ctr.country_code
			WHERE x.%s IS NULL
			LIMIT 1;`, table, codeColumn, codeColumn)

		var countryCode string
		if err := testDB.QueryRow(query).Scan(&countryCode); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", false
			}
			t.Fatalf("failed to find country without data in %s: %v", table, err)
		}

		return countryCode, true
	}

	findCountryCodeWithoutNumbeoWithLegatum := func(t *testing.T) (string, bool) {
		t.Helper()

		query := `
			SELECT ctr.country_code
			FROM countries ctr
			LEFT JOIN numbeo_country_indices ni ON ni.country_code = ctr.country_code
			JOIN legatum_country_indices li ON li.country_code = ctr.country_code
			WHERE ni.country_code IS NULL
			LIMIT 1;`

		var countryCode string
		if err := testDB.QueryRow(query).Scan(&countryCode); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", false
			}
			t.Fatalf("failed to find country without numbeo and with legatum: %v", err)
		}

		return countryCode, true
	}

	t.Run("numbeo_indices requested and absent => explicit null", func(t *testing.T) {
		countryCode, ok := findCountryCodeWithoutNumbeoWithLegatum(t)
		if !ok {
			t.Skip("no country without numbeo and with legatum in test dataset")
		}
		statusCode, _, body := ts.sendRequest(t, "GET", fmt.Sprintf("/countries/%s?include=numbeo_indices,legatum_indices", countryCode), mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "country", "numbeo_indices"), true)
		assert.Equal(t, jsonIsNull(body, "country", "numbeo_indices"), true)
		assert.Equal(t, jsonHasKey(body, "country", "legatum_indices"), true)
		assert.Equal(t, jsonIsNull(body, "country", "legatum_indices"), false)
	})

	t.Run("include not requested => fields omitted", func(t *testing.T) {
		statusCode, _, body := ts.sendRequest(t, "GET", "/countries/usa", mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "country", "numbeo_indices"), false)
		assert.Equal(t, jsonHasKey(body, "country", "legatum_indices"), false)
	})

	t.Run("legatum_indices requested and absent => explicit null", func(t *testing.T) {
		countryCode, ok := findCountryCodeWithoutData(t, "legatum_country_indices", "country_code")
		if !ok {
			t.Skip("no country without legatum_country_indices in test dataset")
		}
		statusCode, _, body := ts.sendRequest(t, "GET", fmt.Sprintf("/countries/%s?include=legatum_indices", countryCode), mocks.Headers, nil)
		assert.Equal(t, statusCode, http.StatusOK)
		assert.Equal(t, jsonHasKey(body, "country", "legatum_indices"), true)
		assert.Equal(t, jsonIsNull(body, "country", "legatum_indices"), true)
	})
}

func jsonHasKey(body []byte, rootKey, key string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}

	rootRaw, ok := payload[rootKey]
	if !ok {
		return false
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(rootRaw, &root); err != nil {
		return false
	}

	_, ok = root[key]
	return ok
}

func jsonIsNull(body []byte, rootKey, key string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}

	rootRaw, ok := payload[rootKey]
	if !ok {
		return false
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(rootRaw, &root); err != nil {
		return false
	}

	v, ok := root[key]
	if !ok {
		return false
	}

	return string(v) == "null"
}

func jsonArrayObjectHasKeyByID(body []byte, rootKey, idKey string, id int64, key string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}

	rootRaw, ok := payload[rootKey]
	if !ok {
		return false
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(rootRaw, &items); err != nil {
		return false
	}

	for _, item := range items {
		idRaw, ok := item[idKey]
		if !ok {
			continue
		}

		var gotID int64
		if err := json.Unmarshal(idRaw, &gotID); err != nil || gotID != id {
			continue
		}

		_, exists := item[key]
		return exists
	}

	return false
}

func jsonArrayObjectIsNullByID(body []byte, rootKey, idKey string, id int64, key string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}

	rootRaw, ok := payload[rootKey]
	if !ok {
		return false
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(rootRaw, &items); err != nil {
		return false
	}

	for _, item := range items {
		idRaw, ok := item[idKey]
		if !ok {
			continue
		}

		var gotID int64
		if err := json.Unmarshal(idRaw, &gotID); err != nil || gotID != id {
			continue
		}

		raw, exists := item[key]
		if !exists {
			return false
		}
		return string(raw) == "null"
	}

	return false
}

func jsonArrayObjectHasKeyByStringID(body []byte, rootKey, idKey, id, key string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}

	rootRaw, ok := payload[rootKey]
	if !ok {
		return false
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(rootRaw, &items); err != nil {
		return false
	}

	for _, item := range items {
		idRaw, ok := item[idKey]
		if !ok {
			continue
		}

		var gotID string
		if err := json.Unmarshal(idRaw, &gotID); err != nil || gotID != id {
			continue
		}

		_, exists := item[key]
		return exists
	}

	return false
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

	for _, code := range []string{"AUS", "ITA", "THA"} {
		t.Run(fmt.Sprintf("Check country code=%s", code), func(t *testing.T) {
			assert.DeepEqual(t, findCountryByCode(t, got.Countries, code), fetchExpectedCountry(t, code))
		})
	}
}

func findCountryByCode(t *testing.T, countries []data.Country, code string) data.Country {
	t.Helper()

	for _, country := range countries {
		if country.Code == code {
			return country
		}
	}

	t.Fatalf("country with code=%s not found in response", code)
	return data.Country{}
}

// TestCountriesBatchByCodes tests batch retrieval for "/countries?country_codes=...".
func TestCountriesBatchByCodes(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/countries?country_codes=usa,rus,usa")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got gotResponse
	unmarshalJSON(t, body, &got)
	assert.Equal(t, len(got.Countries), 2)
	assert.Equal(t, got.Countries[0].Code, "RUS")
	assert.Equal(t, got.Countries[1].Code, "USA")
}

// TestCountriesBatchByCodesWithInclude tests include behavior for "/countries?country_codes=...&include=...".
func TestCountriesBatchByCodesWithInclude(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	statusCode, header, body := ts.get(t, "/countries?country_codes=USA,CAN&include=numbeo_indices")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")
	assert.Equal(t, jsonArrayObjectHasKeyByStringID(body, "countries", "country_code", "USA", "numbeo_indices"), true)
	assert.Equal(t, jsonArrayObjectHasKeyByStringID(body, "countries", "country_code", "CAN", "numbeo_indices"), true)

	statusCode, header, body = ts.get(t, "/countries?country_codes=USA,CAN&include=legatum_indices")
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, header.Get("content-type"), "application/json")
	assert.Equal(t, jsonArrayObjectHasKeyByStringID(body, "countries", "country_code", "USA", "legatum_indices"), true)
	assert.Equal(t, jsonArrayObjectHasKeyByStringID(body, "countries", "country_code", "CAN", "legatum_indices"), true)
}

// TestCountriesBatchByCodesLimit tests batch country_codes limit.
func TestCountriesBatchByCodesLimit(t *testing.T) {
	ts := newTestServer(testApp.routes())
	defer ts.Close()

	rawCodes := make([]string, 0, 21)
	for i := 0; i < 21; i++ {
		rawCodes = append(rawCodes, fmt.Sprintf("C%02d", i))
	}

	urlPath := fmt.Sprintf("/countries?country_codes=%s&include=numbeo_indices", strings.Join(rawCodes, ","))
	statusCode, header, body := ts.get(t, urlPath)
	assert.Equal(t, statusCode, http.StatusUnprocessableEntity)
	assert.Equal(t, header.Get("content-type"), "application/json")

	var got gotResponse
	unmarshalJSON(t, body, &got)
	assert.DeepEqual(t, got.Error, map[string]any{
		"country_codes": "country_codes cannot contain more than 20 unique values",
	})
}

// TestCountry tests the “/countries/:alpha3" endpoint.
func TestCountry(t *testing.T) {
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		urlPath    string
		statusCode int
		code       string
	}{
		{
			name:       "Valid uppercase code (AUS)",
			urlPath:    "/countries/AUS",
			statusCode: http.StatusOK,
			code:       "AUS",
		},
		{
			name:       "Valid mixed case code (iTa)",
			urlPath:    "/countries/iTa",
			statusCode: http.StatusOK,
			code:       "ITA",
		},
		{
			name:       "Valid lowercase code (tha)",
			urlPath:    "/countries/tha",
			statusCode: http.StatusOK,
			code:       "THA",
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
			statusCode: http.StatusNotFound,
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
			urlPath:    "/countries/" + url.QueryEscape("usa'; DROP TABLE countries;"),
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, "GET", tt.urlPath, mocks.Headers, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			if tt.statusCode == http.StatusOK {
				assert.DeepEqual(t, got.Country, fetchExpectedCountry(t, tt.code))
			}

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
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	findCountryCodeWithoutNumbeoWithLegatum := func(t *testing.T) (string, bool) {
		t.Helper()

		query := `
			SELECT ctr.country_code
			FROM countries ctr
			LEFT JOIN numbeo_country_indices ni ON ni.country_code = ctr.country_code
			JOIN legatum_country_indices li ON li.country_code = ctr.country_code
			WHERE ni.country_code IS NULL
			LIMIT 1;`

		var countryCode string
		if err := testDB.QueryRow(query).Scan(&countryCode); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", false
			}
			t.Fatalf("failed to find country without numbeo and with legatum: %v", err)
		}

		return countryCode, true
	}

	tests := []struct {
		name        string
		urlPath     string
		statusCode  int
		queryParams queryParamsCountry
	}{
		{
			name:       "Enable only Numbeo indices",
			urlPath:    "/countries/rus?include=numbeo_indices",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:       "Enable only Legatum indices",
			urlPath:    "/countries/usa?include=legatum_indices",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Enable all params (Numbeo and Legatum)",
			urlPath:    "/countries/bra?include=numbeo_indices,legatum_indices",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Enable both params with missing Numbeo data",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Enable both params with missing both data",
			urlPath:    "/countries/wlf?include=numbeo_indices,legatum_indices",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:       "Include empty",
			urlPath:    "/countries/can?include=",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  false,
				legatumIndicesEnabled: false,
			},
		},
		{
			name:       "Unknown params (mixed cases)",
			urlPath:    "/countries/arg?Numbeo_Indices=true&LEGATHUM_INDICES=1&InvalidParam=TRUE",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Duplicate includes",
			urlPath:    "/countries/chn?include=numbeo_indices,numbeo_indices,legatum_indices",
			statusCode: http.StatusOK,
			queryParams: queryParamsCountry{
				numbeoIndicesEnabled:  true,
				legatumIndicesEnabled: true,
			},
		},
		{
			name:       "Unprocessable include value (123)",
			urlPath:    "/countries/deu?include=123",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Unprocessable include value (abc)",
			urlPath:    "/countries/nld?include=abc",
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlPath := tt.urlPath
			if tt.name == "Enable both params with missing Numbeo data" {
				countryCode, ok := findCountryCodeWithoutNumbeoWithLegatum(t)
				if !ok {
					t.Skip("no country without numbeo and with legatum in test dataset")
				}
				urlPath = fmt.Sprintf("/countries/%s?include=numbeo_indices,legatum_indices", strings.ToLower(countryCode))
			}

			statusCode, header, body := ts.sendRequest(t, "GET", urlPath, mocks.Headers, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			assert.DeepEqual(t, countryFildsToBool(got.Country), tt.queryParams)

			if tt.statusCode == http.StatusUnprocessableEntity {
				wantError := map[string]any{
					"include": "include contains unsupported value \"123\"",
				}
				if tt.name == "Unprocessable include value (abc)" {
					wantError = map[string]any{
						"include": "include contains unsupported value \"abc\"",
					}
				}
				if tt.name == "Unknown params (mixed cases)" {
					wantError = map[string]any{
						"query": "unknown query parameter \"InvalidParam\"",
					}
				}
				assert.DeepEqual(t, got.Error, wantError)
			}
		})
	}
}

// TestUnknownQueryParams tests that all endpoints reject unknown query parameters.
func TestUnknownQueryParams(t *testing.T) {
	ts := newTestServerWithMockUser(testApp.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		method     string
		urlPath    string
		headers    http.Header
		statusCode int
	}{
		{name: "healthcheck", method: http.MethodGet, urlPath: "/healthcheck?foo=bar", headers: nil, statusCode: http.StatusUnprocessableEntity},
		{name: "cities list", method: http.MethodGet, urlPath: "/cities?foo=bar", headers: nil, statusCode: http.StatusUnprocessableEntity},
		{name: "cities detail", method: http.MethodGet, urlPath: "/cities/5809844?foo=bar", headers: mocks.Headers, statusCode: http.StatusUnprocessableEntity},
		{name: "countries list", method: http.MethodGet, urlPath: "/countries?foo=bar", headers: nil, statusCode: http.StatusUnprocessableEntity},
		{name: "countries detail", method: http.MethodGet, urlPath: "/countries/AUS?foo=bar", headers: mocks.Headers, statusCode: http.StatusUnprocessableEntity},
		{name: "register user", method: http.MethodPost, urlPath: "/users?foo=bar", headers: nil, statusCode: http.StatusUnprocessableEntity},
		{name: "activate user", method: http.MethodPut, urlPath: "/users/activated?foo=bar", headers: nil, statusCode: http.StatusUnprocessableEntity},
		{name: "authentication token", method: http.MethodPost, urlPath: "/tokens/authentication?foo=bar", headers: nil, statusCode: http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, tt.method, tt.urlPath, tt.headers, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			var got gotResponse
			unmarshalJSON(t, body, &got)
			wantError := map[string]any{
				"query": "unknown query parameter \"foo\"",
			}
			assert.DeepEqual(t, got.Error, wantError)
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
			statusCode, header, body := ts.sendRequest(t, tt.method, tt.urlPath, nil, nil)
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
			statusCode, header, body := ts.sendRequest(t, "POST", "/users", nil, tt.payload)
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

	// User registration
	statusCode, _, _ := ts.sendRequest(t, "POST", "/users", nil, inputUser)
	assert.Equal(t, statusCode, http.StatusAccepted)

	// User activation via e-mail
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
			statusCode, header, body := ts.sendRequest(t, "PUT", "/users/activated", nil, tt.tokenMessage)
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
				assert.DeepEqual(t, got.Error, tt.errorMessage) // User registration
			}
		})
	}
}

func TestAuthorizationUser(t *testing.T) {
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

	// User registration
	statusCode, _, _ := ts.sendRequest(t, "POST", "/users", nil, inputUser)
	assert.Equal(t, statusCode, http.StatusAccepted)

	// User activation via e-mail
	mockMailer := testApp.mailer.(*mocks.MockMailer)
	plainBody := mockMailer.Email.PlainBody
	bodyMap, ok := plainBody.(map[string]any)
	if !ok {
		t.Fatal("plainBody is not a map[string]any")
	}
	activationToken, exists := bodyMap["activationToken"]
	if !exists {
		t.Fatal("activationToken not found in plainBody")
	}
	statusCode, _, _ = ts.sendRequest(t, "PUT", "/users/activated", nil, map[string]any{"token": activationToken})
	assert.Equal(t, statusCode, http.StatusOK)

	// Unregistered user authentication
	notExistUser := map[string]string{"email": "alice@example.com", "password": "invalidPa55word"}
	statusCode, _, body := ts.sendRequest(t, "POST", "/tokens/authentication", nil, notExistUser)
	assert.Equal(t, statusCode, http.StatusUnauthorized)
	var gotError map[string]any
	unmarshalJSON(t, body, &gotError)
	wantError := map[string]any{"error": "invalid authentication credentials"}
	assert.DeepEqual(t, gotError, wantError)

	// User authentication
	user := map[string]string{"email": "john@example.com", "password": "validPa55word"}
	statusCode, _, body = ts.sendRequest(t, "POST", "/tokens/authentication", nil, user)
	assert.Equal(t, statusCode, http.StatusCreated)
	var got gotResponse
	unmarshalJSON(t, body, &got)
	authToken := got.AuthToken.Token

	tests := []struct {
		name         string
		header       http.Header
		urlPath      string
		statusCode   int
		errorMessage string
	}{
		{
			name:       "Valid authentication header",
			header:     http.Header{"Authorization": []string{"Bearer " + authToken}},
			urlPath:    "/cities/2562305",
			statusCode: http.StatusOK,
		},
		{
			name:         "Authentication with invalid token",
			header:       http.Header{"Authorization": []string{"Bearer " + "XXXXXXXXXXXXXXX"}},
			urlPath:      "/cities/2562305",
			statusCode:   http.StatusUnauthorized,
			errorMessage: "invalid or missing authentication token",
		},
		{
			name:         "Malformed authorization header",
			header:       http.Header{"Authorization": []string{"INVALID"}},
			urlPath:      "/cities/2562305",
			statusCode:   http.StatusUnauthorized,
			errorMessage: "invalid or missing authentication token",
		},
		{
			name:         "Missing required authorization header",
			urlPath:      "/cities/2562305",
			statusCode:   http.StatusUnauthorized,
			errorMessage: "you must be authenticated to access this resource",
		},
		{
			name:       "No authorization header provided (optional)",
			urlPath:    "/countries",
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, header, body := ts.sendRequest(t, "GET", tt.urlPath, tt.header, nil)
			assert.Equal(t, statusCode, tt.statusCode)
			assert.Equal(t, header.Get("content-type"), "application/json")

			if tt.statusCode != http.StatusOK {
				var got gotResponse
				unmarshalJSON(t, body, &got)
				assert.DeepEqual(t, got.Error, tt.errorMessage)
			}
		})
	}
}
