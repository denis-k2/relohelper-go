package data

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/denis-k2/relohelper-go/internal/assert"
)

func TestListCities(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	cities, err := models.Cities.ListCities("", NewIncludeSet())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(cities), 534)
	assert.Equal(t, cities[0].CityID, int64(1))
	assert.Equal(t, cities[0].Country, "")
}

func TestListCitiesWithCountryInclude(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	cities, err := models.Cities.ListCities("USA", NewIncludeSet("country"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(cities) > 0, true)
	assert.Equal(t, cities[0].Country != "", true)
}

func TestGetCity(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	city, err := models.Cities.GetCity(273, NewIncludeSet("country", "numbeo_cost", "numbeo_indices", "avg_climate"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, city.CityID, int64(273))
	assert.Equal(t, city.Country, "Japan")
	assert.Equal(t, city.NumbeoCost != nil, true)
	assert.Equal(t, city.NumbeoIndices != nil, true)
}

func TestGetCitiesByIDs(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	cities, err := models.Cities.GetCitiesByIDs([]int64{11, 94, 11}, NewIncludeSet("country"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(cities), 2)
	assert.Equal(t, cities[0].CityID, int64(11))
	assert.Equal(t, cities[0].Country != "", true)
	assert.Equal(t, cities[1].CityID, int64(94))
}
