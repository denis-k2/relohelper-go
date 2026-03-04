package data

import (
	"database/sql"
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
	assert.Equal(t, cities[0].ID, int64(1))
	assert.Equal(t, cities[0].CountryName, "")
}

func TestListCitiesWithCountryInclude(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	cities, err := models.Cities.ListCities("USA", NewIncludeSet("country"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(cities) > 0, true)
	assert.Equal(t, cities[0].CountryName != "", true)
}

func TestGetCity(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	city, err := models.Cities.GetCity(273, NewIncludeSet("country", "numbeo_cost", "numbeo_indices", "avg_climate"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, city.ID, int64(273))
	assert.Equal(t, city.CountryName, "Japan")
	assert.Equal(t, city.NumbeoCost != nil, true)
	assert.Equal(t, city.NumbeoCityIndices != nil, true)
}

func TestGetCitiesByIDs(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	cities, err := models.Cities.GetCitiesByIDs([]int64{11, 94, 11}, NewIncludeSet("country"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(cities), 2)
	assert.Equal(t, cities[0].ID, int64(11))
	assert.Equal(t, cities[0].CountryName != "", true)
	assert.Equal(t, cities[1].ID, int64(94))
}

func TestGetCityAvgClimateOrderedByMonth(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	var cityID int64
	err := db.QueryRow(`
		SELECT city_id
		FROM avg_climate
		GROUP BY city_id
		HAVING COUNT(*) = 12
		LIMIT 1;
	`).Scan(&cityID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("testing city_id=%d", cityID)

	city, err := models.Cities.GetCity(cityID, NewIncludeSet("avg_climate"))
	if err != nil {
		t.Fatal(err)
	}
	if city.AvgClimate == nil {
		t.Fatal("expected avg_climate to be present")
	}

	expected := make([]*float64, 12)
	rows, err := db.Query(`
		SELECT month, high_temp
		FROM avg_climate
		WHERE city_id = $1
		ORDER BY month;`, cityID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Fatalf("failed to close rows: %v", err)
		}
	}()

	for rows.Next() {
		var (
			month int
			value sql.NullFloat64
		)
		if err := rows.Scan(&month, &value); err != nil {
			t.Fatal(err)
		}
		if value.Valid {
			v := value.Float64
			expected[month-1] = &v
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(city.AvgClimate.HighTemp), 12)
	for i := range expected {
		if !equalFloatPtrs(expected[i], city.AvgClimate.HighTemp[i]) {
			t.Fatalf("city_id=%d high_temp[%d] mismatch: got=%v want=%v", cityID, i, city.AvgClimate.HighTemp[i], expected[i])
		}
	}
}

func TestGetCityAvgClimateSeaTempAllNull(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	var cityID int64
	err := db.QueryRow(`
		SELECT city_id
		FROM avg_climate
		GROUP BY city_id
		HAVING COUNT(*) = 12 AND COUNT(sea_temp) = 0
		LIMIT 1;
	`).Scan(&cityID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("testing city_id=%d", cityID)

	city, err := models.Cities.GetCity(cityID, NewIncludeSet("avg_climate"))
	if err != nil {
		t.Fatal(err)
	}
	if city.AvgClimate == nil {
		t.Fatal("expected avg_climate to be present")
	}

	assert.Equal(t, len(city.AvgClimate.SeaTemp), 12)
	for i, value := range city.AvgClimate.SeaTemp {
		if value != nil {
			t.Fatalf("city_id=%d expected sea_temp[%d] to be nil, got %v", cityID, i, *value)
		}
	}
}

func equalFloatPtrs(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
