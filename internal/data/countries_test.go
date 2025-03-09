package data

import (
	"fmt"
	"testing"

	_ "github.com/lib/pq"

	"github.com/denis-k2/relohelper-go/internal/assert"
)

func TestGetCountryList(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	countries, err := models.Countries.GetCountryList()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(countries), 249)

	tests := []struct {
		index   int
		country Country
	}{
		{
			index: 8,
			country: Country{
				Code: "ARG",
				Name: "Argentina",
			},
		},
		{
			index: 189,
			country: Country{
				Code: "RUS",
				Name: "Russian Federation",
			},
		},
		{
			index: 234,
			country: Country{
				Code: "USA",
				Name: "United States of America",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Check country code=%s", tt.country.Code), func(t *testing.T) {
			assert.DeepEqual(t, *countries[tt.index], tt.country)
		})
	}
}
