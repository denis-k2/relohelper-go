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

	cities, err := models.Cities.GetCityList()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 534, len(cities))

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
