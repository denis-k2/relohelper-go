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

func TestGetCountry(t *testing.T) {
	db := newTestDB(t)
	models := NewModels(db)

	tests := []struct {
		name        string
		countryCode string
		country     Country
	}{
		{
			name:        "Valid uppercase code (AUT)",
			countryCode: "AUT",
			country: Country{
				Code: "AUT",
				Name: "Austria",
			},
		},
		{
			name:        "Valid mixed case code (Mex)",
			countryCode: "Mex",
			country: Country{
				Code: "MEX",
				Name: "Mexico",
			},
		},
		{
			name:        "Valid lowercase code (srb)",
			countryCode: "srb",
			country: Country{
				Code: "SRB",
				Name: "Serbia",
			},
		},
		{
			name:        "Nonexistent country code (XXXX)",
			countryCode: "xxxx",
		},
		{
			name:        "Non-alphabetic code (123)",
			countryCode: "123",
		},
		{
			name:        "Empty country code",
			countryCode: "",
		},
		{
			name:        "Code with 1 letter (a)",
			countryCode: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			country, err := models.Countries.GetCountry(tt.countryCode)
			if err != nil {
				assert.Equal(t, err, ErrRecordNotFound)
				return
			}
			assert.DeepEqual(t, *country, tt.country)
		})
	}
}
