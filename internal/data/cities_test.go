package data

import (
	"fmt"
	"reflect"
	"testing"

	_ "github.com/lib/pq"

	"github.com/denis-k2/relohelper-go/internal/assert"
)

func TestGetCityList(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	cities, err := models.Cities.GetCityList("")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(cities), 534)

	stateOH := "US-OH"
	wantCities := []City{
		{
			CityID:      1,
			City:        "Hamilton",
			StateCode:   nil,
			CountryCode: "BMU",
		},
		{
			CityID:      235,
			City:        "Cincinnati",
			StateCode:   &stateOH,
			CountryCode: "USA",
		},
		{
			CityID:      534,
			City:        "Karachi",
			StateCode:   nil,
			CountryCode: "PAK",
		},
	}
	for _, city := range wantCities {
		assert.DeepEqual(t, *cities[city.CityID-1], city)
	}
}

func TestGetCityListByCountry(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	tests := []struct {
		name        string
		countryCode string
		expectedCnt int
	}{
		{
			name:        "Valid uppercase code (GBR)",
			countryCode: "GBR",
			expectedCnt: 32,
		},
		{
			name:        "Valid mixed case code (Deu)",
			countryCode: "Deu",
			expectedCnt: 23,
		},
		{
			name:        "Valid lowercase code (rus)",
			countryCode: "arg",
			expectedCnt: 1,
		},
		{
			name:        "Nonexistent country code (XXXX)",
			countryCode: "xxxx",
			expectedCnt: 0,
		},
		{
			name:        "Non-alphabetic code (123)",
			countryCode: "123",
			expectedCnt: 0,
		},
		{
			name:        "Empty country code",
			countryCode: "",
			expectedCnt: 534,
		},
		{
			name:        "Code with 1 letter (a)",
			countryCode: "a",
			expectedCnt: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cities, err := models.Cities.GetCityList(tt.countryCode)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, len(cities), tt.expectedCnt)

			if tt.expectedCnt > 0 && tt.expectedCnt < 534 {
				// Check uniqueness of CityID.
				idSet := make(map[int64]struct{}, len(cities))
				for _, city := range cities {
					idSet[city.CityID] = struct{}{}
				}
				assert.Equal(t, len(idSet), len(cities))

				// Ensure all cities have the same CountryCode.
				countrySet := make(map[string]struct{}, len(cities))
				for _, city := range cities {
					countrySet[city.CountryCode] = struct{}{}
				}
				assert.Equal(t, len(countrySet), 1)
			}
		})
	}
}

func TestGetCityID(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	stateMA := "US-MA"
	wantCity := []City{
		{
			CityID:      19,
			City:        "Boston",
			StateCode:   &stateMA,
			CountryCode: "USA",
			Country:     "United States of America",
		},
		{
			CityID:      319,
			City:        "Valencia",
			StateCode:   nil,
			CountryCode: "ESP",
			Country:     "Spain",
		},
		{
			CityID:      487,
			City:        "Kaliningrad",
			StateCode:   nil,
			CountryCode: "RUS",
			Country:     "Russian Federation",
		},
		{
			CityID: 4000,
		},
		{
			CityID: -5000,
		},
	}

	for _, want := range wantCity {
		t.Run(fmt.Sprintf("Get City id=%d", want.CityID), func(t *testing.T) {
			city, err := models.Cities.GetCityID(want.CityID)
			if err != nil {
				assert.Equal(t, err, ErrRecordNotFound)
				return
			}
			assert.DeepEqual(t, *city, want)
		})
	}
}

func TestGetNumbeoCost(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	tests := []struct {
		name       string
		cityID     int64
		currency   string
		dateLength int
		priceItems int
	}{
		{"exist", 100, "USD", 10, 57},
		{"exist", 200, "USD", 10, 57},
		{"exist", 300, "USD", 10, 57},
		{"exist", 400, "USD", 10, 57},
		{"exist", 500, "USD", 10, 57},
		{"not exist", 600, "", 0, 0},
		{"not exist", -700, "", 0, 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Get Numbeo Cost by City id=%d (%s)", tt.cityID, tt.name), func(t *testing.T) {
			cost, err := models.Cities.GetNumbeoCost(tt.cityID)
			if err != nil {
				assert.Equal(t, err, ErrRecordNotFound)
				return
			}
			assert.Equal(t, cost.Currency, tt.currency)
			assert.Equal(t, len(cost.LastUpdate), tt.dateLength)
			assert.Equal(t, len(cost.Prices), tt.priceItems)
		})
	}
}

func TestGetNumbeoIndicies(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	tests := []struct {
		name       string
		cityID     int64
		itemsCount int
		dateLength int
		valueFloat float64
		valueNil   *float64
	}{
		{"exist", 100, 13, 10, 71.7, nil},
		{"exist", 200, 13, 10, 64.8, nil},
		{"exist", 300, 13, 10, 51.8, nil},
		{"exist", 400, 13, 10, 39.1, nil},
		{"exist", 500, 13, 10, 28.9, nil},
		{"not exist", 600, 0, 0, 0, nil},
		{"not exist", -700, 0, 0, 0, nil},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Get Numbeo Indicies by City id=%d (%s)", tt.cityID, tt.name), func(t *testing.T) {
			index, err := models.Cities.GetNumbeoIndicies(tt.cityID)
			if err != nil {
				assert.Equal(t, err, ErrRecordNotFound)
				return
			}
			assert.Equal(t, len(index.LastUpdate), tt.dateLength)
			assert.Equal(t, reflect.TypeOf(Indices{}).NumField(), tt.itemsCount)
			assert.Equal(t, *index.CostOfLiving, tt.valueFloat)
			assert.Equal(t, index.QualityOfLife, tt.valueNil)
		})
	}
}
