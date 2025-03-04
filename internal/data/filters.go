package data

import (
	"regexp"
	"strconv"

	"github.com/denis-k2/relohelper-go/internal/validator"
)

type Filters struct {
	CountryCode string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	countryCodeRegex := regexp.MustCompile(`^[A-Za-z]{3}$`)
	v.Check(countryCodeRegex.MatchString(f.CountryCode), "country_code", "must be exactly three English letters")
}

func ValidateBoolQuery(v *validator.Validator, s string) bool {
	if s != "" {
		boolQueryValue, err := strconv.ParseBool(s)
		if err != nil {
			v.Check(false, "query parameter", "must be a boolean value")
		}
		return boolQueryValue
	}
	return false
}
