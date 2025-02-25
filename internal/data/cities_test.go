package data

import (
	"fmt"
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
		},
		{
			CityID:      319,
			City:        "Valencia",
			StateCode:   nil,
			CountryCode: "ESP",
		},
		{
			CityID:      487,
			City:        "Kaliningrad",
			StateCode:   nil,
			CountryCode: "RUS",
		},
	}

	for _, want := range wantCity {
		t.Run(fmt.Sprintf("Get City id=%d", want.CityID), func(t *testing.T) {
			city, err := models.Cities.GetCityID(want.CityID)
			if err != nil {
				t.Fatal(err)
			}
			assert.DeepEqual(t, *city, want)
		})
	}
}
