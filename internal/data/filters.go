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

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateInputUser(v *validator.Validator, inputUser InputUser) {
	v.Check(inputUser.Name != "", "name", "must be provided")
	v.Check(len(inputUser.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, inputUser.Email)
	ValidatePasswordPlaintext(v, inputUser.PlainPassword)
}
