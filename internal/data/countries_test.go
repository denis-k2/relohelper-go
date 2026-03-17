package data

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/denis-k2/relohelper-go/internal/assert"
)

func TestListCountries(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	countries, err := models.Countries.ListCountries()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(countries), 249)
	assert.Equal(t, countries[0].Code != "", true)
	assert.Equal(t, countries[0].LastUpdate != "", true)
}

func TestGetCountry(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	country, err := models.Countries.GetCountry("USA", NewIncludeSet("numbeo_indices", "legatum_indices"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, country.Code, "USA")
	assert.Equal(t, country.LastUpdate != "", true)
	assert.Equal(t, country.NumbeoCountryIndices != nil, true)
	assert.Equal(t, country.LegatumCountryIndices != nil, true)
}

func TestGetCountriesByCodes(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	countries, err := models.Countries.GetCountriesByCodes([]string{"USA", "RUS", "USA"}, NewIncludeSet())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(countries), 2)
	assert.Equal(t, countries[0].Code, "RUS")
	assert.Equal(t, countries[0].LastUpdate != "", true)
	assert.Equal(t, countries[1].Code, "USA")
}
