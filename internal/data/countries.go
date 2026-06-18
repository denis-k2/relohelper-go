package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/denis-k2/relohelper-go/internal/db"
)

type Country struct {
	Code                  string                 `json:"country_code"`
	Name                  string                 `json:"country"`
	Population            *int64                 `json:"population"`
	Area                  *int64                 `json:"area"`
	LastUpdate            string                 `json:"last_update"`
	NumbeoCountryIndices  *NumbeoCountryIndices  `json:"numbeo_indices,omitzero"`
	LegatumCountryIndices *LegatumCountryIndices `json:"legatum_indices,omitzero"`
}

type NumbeoCountryIndices struct {
	CostOfLiving               *float64 `json:"cost_of_living"`
	Rent                       *float64 `json:"rent"`
	CostOfLivingPlusRent       *float64 `json:"cost_of_living_plus_rent"`
	Groceries                  *float64 `json:"groceries"`
	RestaurantPrice            *float64 `json:"restaurant_price"`
	LocalPurchasingPower       *float64 `json:"local_purchasing_power"`
	QualityOfLife              *float64 `json:"quality_of_life"`
	PropertyPriceToIncomeRatio *float64 `json:"property_price_to_income_ratio"`
	TrafficCommuteTime         *float64 `json:"traffic_commute_time"`
	Climate                    *float64 `json:"climate"`
	Safety                     *float64 `json:"safety"`
	HealthCare                 *float64 `json:"health_care"`
	Pollution                  *float64 `json:"pollution"`
	AvgSalaryUSD               *float64 `json:"avg_salary_usd"`
	LastUpdate                 string   `json:"last_update"`
}

type LegatumCountryIndices struct {
	SafetyAndSecurity             RankAndScore `json:"safety_and_security"`
	PersonalFreedom               RankAndScore `json:"personal_freedom"`
	Governance                    RankAndScore `json:"governance"`
	SocialCapital                 RankAndScore `json:"social_capital"`
	InvestmentEnvironment         RankAndScore `json:"investment_invironment"`
	EnterpriseConditions          RankAndScore `json:"enterprise_conditions"`
	InfrastructureAndMarketAccess RankAndScore `json:"infrastructure_and_market_access"`
	EconomicQuality               RankAndScore `json:"economic_quality"`
	LivingConditions              RankAndScore `json:"living_conditions"`
	Health                        RankAndScore `json:"health"`
	Education                     RankAndScore `json:"education"`
	NaturalEnvironment            RankAndScore `json:"natural_environment"`
}

type RankAndScore struct {
	Rank2007  int     `json:"rank_2007"`
	Rank2008  int     `json:"rank_2008"`
	Rank2009  int     `json:"rank_2009"`
	Rank2010  int     `json:"rank_2010"`
	Rank2011  int     `json:"rank_2011"`
	Rank2012  int     `json:"rank_2012"`
	Rank2013  int     `json:"rank_2013"`
	Rank2014  int     `json:"rank_2014"`
	Rank2015  int     `json:"rank_2015"`
	Rank2016  int     `json:"rank_2016"`
	Rank2017  int     `json:"rank_2017"`
	Rank2018  int     `json:"rank_2018"`
	Rank2019  int     `json:"rank_2019"`
	Rank2020  int     `json:"rank_2020"`
	Rank2021  int     `json:"rank_2021"`
	Rank2022  int     `json:"rank_2022"`
	Rank2023  int     `json:"rank_2023"`
	Score2007 float64 `json:"score_2007"`
	Score2008 float64 `json:"score_2008"`
	Score2009 float64 `json:"score_2009"`
	Score2010 float64 `json:"score_2010"`
	Score2011 float64 `json:"score_2011"`
	Score2012 float64 `json:"score_2012"`
	Score2013 float64 `json:"score_2013"`
	Score2014 float64 `json:"score_2014"`
	Score2015 float64 `json:"score_2015"`
	Score2016 float64 `json:"score_2016"`
	Score2017 float64 `json:"score_2017"`
	Score2018 float64 `json:"score_2018"`
	Score2019 float64 `json:"score_2019"`
	Score2020 float64 `json:"score_2020"`
	Score2021 float64 `json:"score_2021"`
	Score2022 float64 `json:"score_2022"`
	Score2023 float64 `json:"score_2023"`
}

type CountryModel struct {
	Queries *db.Queries
}

func (c CountryModel) ListCountries() ([]*Country, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.Queries.ListCountries(ctx)
	if err != nil {
		return nil, err
	}

	countries := make([]*Country, 0, len(rows))
	for _, row := range rows {
		countries = append(countries, newCountryFromListRow(row))
	}

	return countries, nil
}

func (c CountryModel) GetCountry(countryCode string, include IncludeSet) (*Country, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row, err := c.Queries.GetCountry(ctx, countryCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	country := newCountryFromGetRow(row)
	countryByCode := map[string]*Country{country.Code: country}

	if include.Has("numbeo_indices") {
		if err := c.attachNumbeoIndicesByCodes(ctx, []string{country.Code}, countryByCode); err != nil {
			return nil, err
		}
	}

	if include.Has("legatum_indices") {
		if err := c.attachLegatumIndicesByCodes(ctx, []string{country.Code}, countryByCode); err != nil {
			return nil, err
		}
	}

	return country, nil
}

func (c CountryModel) GetCountriesByCodes(codes []string, include IncludeSet) ([]*Country, error) {
	if len(codes) == 0 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.Queries.GetCountriesByCodes(ctx, codes)
	if err != nil {
		return nil, err
	}

	countries := make([]*Country, 0, len(rows))
	countryByCode := make(map[string]*Country, len(codes))
	for _, row := range rows {
		country := newCountryFromCodesRow(row)
		countries = append(countries, country)
		countryByCode[country.Code] = country
	}
	if len(countries) == 0 {
		return nil, ErrRecordNotFound
	}

	if include.Has("numbeo_indices") {
		err = c.attachNumbeoIndicesByCodes(ctx, codes, countryByCode)
		if err != nil {
			return nil, err
		}
	}

	if include.Has("legatum_indices") {
		err = c.attachLegatumIndicesByCodes(ctx, codes, countryByCode)
		if err != nil {
			return nil, err
		}
	}

	return countries, nil
}

func (c CountryModel) attachNumbeoIndicesByCodes(ctx context.Context, codes []string, countryByCode map[string]*Country) error {
	rows, err := c.Queries.GetNumbeoCountryIndicesByCodes(ctx, codes)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row.NumbeoIndices) == 0 || string(row.NumbeoIndices) == "null" {
			continue
		}

		var indices NumbeoCountryIndices
		if err := json.Unmarshal(row.NumbeoIndices, &indices); err != nil {
			return err
		}

		country, ok := countryByCode[row.CountryCode]
		if !ok {
			continue
		}
		country.NumbeoCountryIndices = &indices
	}

	return nil
}

func (c CountryModel) attachLegatumIndicesByCodes(ctx context.Context, codes []string, countryByCode map[string]*Country) error {
	rows, err := c.Queries.GetLegatumCountryIndicesByCodes(ctx, codes)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row.LegatumIndices) == 0 || string(row.LegatumIndices) == "null" {
			continue
		}

		var indices LegatumCountryIndices
		if err := json.Unmarshal(row.LegatumIndices, &indices); err != nil {
			return err
		}

		country, ok := countryByCode[row.CountryCode]
		if !ok {
			continue
		}
		country.LegatumCountryIndices = &indices
	}

	return nil
}

func newCountryFromListRow(row db.ListCountriesRow) *Country {
	return &Country{
		Code:       row.CountryCode,
		Name:       row.Country,
		Population: row.Population,
		Area:       row.Area,
		LastUpdate: row.LastUpdate,
	}
}

func newCountryFromGetRow(row db.GetCountryRow) *Country {
	return &Country{
		Code:       row.CountryCode,
		Name:       row.Country,
		Population: row.Population,
		Area:       row.Area,
		LastUpdate: row.LastUpdate,
	}
}

func newCountryFromCodesRow(row db.GetCountriesByCodesRow) *Country {
	return &Country{
		Code:       row.CountryCode,
		Name:       row.Country,
		Population: row.Population,
		Area:       row.Area,
		LastUpdate: row.LastUpdate,
	}
}
