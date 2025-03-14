package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Country struct {
	Code           string                 `json:"country_code"`
	Name           string                 `json:"country"`
	NumbeoIndices  *NumbeoCountryIndicies `json:"numbeo_indices,omitempty"`
	LegatumIndices *LegatumCountryIndices `json:"legatum_indices,omitempty"`
}

type NumbeoCountryIndicies struct {
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
	DB *sql.DB
}

func (c CountryModel) GetCountryList() ([]*Country, error) {
	query := `
        SELECT country_code, country
		FROM country
		ORDER BY country_code;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	countries := []*Country{}

	for rows.Next() {
		var country Country

		err := rows.Scan(
			&country.Code,
			&country.Name,
		)

		if err != nil {
			return nil, err
		}

		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

func (c CountryModel) GetCountry(countryCode string) (*Country, error) {
	query := `
		SELECT country_code, country
		FROM country
		WHERE LOWER(country_code) = LOWER($1);`

	var country Country
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, countryCode).Scan(
		&country.Code,
		&country.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &country, nil
}

func (c CountryModel) GetNumbeoCountryIndicies(country_code string) (*NumbeoCountryIndicies, error) {
	query := `
		SELECT
			cost_of_living,
			rent,
			cost_of_living_plus_rent,
			groceries,
			restaurant_price,
			local_purchasing_power,
			quality_of_life,
			property_price_to_income_ratio,
			traffic_commute_time,
			climate,
			safety,
			health_care,
			pollution,
			avg_salary_usd,
			to_char(sys_updated_date, 'YYYY-MM-DD')
		FROM numbeo_index_by_country
		WHERE LOWER(country_code) = LOWER($1);`

	var index NumbeoCountryIndicies

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, country_code).Scan(
		&index.CostOfLiving,
		&index.Rent,
		&index.CostOfLivingPlusRent,
		&index.Groceries,
		&index.RestaurantPrice,
		&index.LocalPurchasingPower,
		&index.QualityOfLife,
		&index.PropertyPriceToIncomeRatio,
		&index.TrafficCommuteTime,
		&index.Climate,
		&index.Safety,
		&index.HealthCare,
		&index.Pollution,
		&index.AvgSalaryUSD,
		&index.LastUpdate,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &index, nil
}

func (c CountryModel) GetLegatumIndicies(country_code string) (*LegatumCountryIndices, error) {
	query := `
	SELECT 
		pillar_name,
		rank_2007, rank_2008, rank_2009, rank_2010, rank_2011, rank_2012, rank_2013, rank_2014,
		rank_2015, rank_2016, rank_2017, rank_2018, rank_2019, rank_2020, rank_2021, rank_2022, rank_2023,
		score_2007, score_2008, score_2009, score_2010, score_2011, score_2012, score_2013, score_2014,
		score_2015, score_2016, score_2017, score_2018, score_2019, score_2020, score_2021, score_2022, score_2023
	FROM legatum_index
	WHERE LOWER(country_code) = LOWER($1);`

	var legatum LegatumCountryIndices

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, country_code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dataFound := false
	for rows.Next() {
		dataFound = true
		var pillarName string
		var rankScore RankAndScore

		err := rows.Scan(
			&pillarName,
			&rankScore.Rank2007,
			&rankScore.Rank2008,
			&rankScore.Rank2009,
			&rankScore.Rank2010,
			&rankScore.Rank2011,
			&rankScore.Rank2012,
			&rankScore.Rank2013,
			&rankScore.Rank2014,
			&rankScore.Rank2015,
			&rankScore.Rank2016,
			&rankScore.Rank2017,
			&rankScore.Rank2018,
			&rankScore.Rank2019,
			&rankScore.Rank2020,
			&rankScore.Rank2021,
			&rankScore.Rank2022,
			&rankScore.Rank2023,
			&rankScore.Score2007,
			&rankScore.Score2008,
			&rankScore.Score2009,
			&rankScore.Score2010,
			&rankScore.Score2011,
			&rankScore.Score2012,
			&rankScore.Score2013,
			&rankScore.Score2014,
			&rankScore.Score2015,
			&rankScore.Score2016,
			&rankScore.Score2017,
			&rankScore.Score2018,
			&rankScore.Score2019,
			&rankScore.Score2020,
			&rankScore.Score2021,
			&rankScore.Score2022,
			&rankScore.Score2023,
		)
		if err != nil {
			return nil, err
		}

		switch pillarName {
		case "Safety and Security":
			legatum.SafetyAndSecurity = rankScore
		case "Personal Freedom":
			legatum.PersonalFreedom = rankScore
		case "Governance":
			legatum.Governance = rankScore
		case "Social Capital":
			legatum.SocialCapital = rankScore
		case "Investment Environment":
			legatum.InvestmentEnvironment = rankScore
		case "Enterprise Conditions":
			legatum.EnterpriseConditions = rankScore
		case "Infrastructure and Market Access":
			legatum.InfrastructureAndMarketAccess = rankScore
		case "Economic Quality":
			legatum.EconomicQuality = rankScore
		case "Living Conditions":
			legatum.LivingConditions = rankScore
		case "Health":
			legatum.Health = rankScore
		case "Education":
			legatum.Education = rankScore
		case "Natural Environment":
			legatum.NaturalEnvironment = rankScore
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if !dataFound {
		return nil, ErrRecordNotFound
	}

	return &legatum, nil
}
