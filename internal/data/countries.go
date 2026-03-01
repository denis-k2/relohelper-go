package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
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

func (c CountryModel) ListCountries() (countries []*Country, retErr error) {
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
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	countries = []*Country{}
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.Code, &country.Name); err != nil {
			return nil, err
		}
		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

func (c CountryModel) GetCountry(countryCode string, include IncludeSet) (*Country, error) {
	query := `
		SELECT
			ctr.country_code,
			ctr.country,
			CASE
				WHEN $2 THEN (
					SELECT row_to_json(n)
					FROM (
						SELECT
							nic.cost_of_living,
							nic.rent,
							nic.cost_of_living_plus_rent,
							nic.groceries,
							nic.restaurant_price,
							nic.local_purchasing_power,
							nic.quality_of_life,
							nic.property_price_to_income_ratio,
							nic.traffic_commute_time,
							nic.climate,
							nic.safety,
							nic.health_care,
							nic.pollution,
							nic.avg_salary_usd,
							to_char(nic.sys_updated_date, 'YYYY-MM-DD') AS last_update
						FROM numbeo_index_by_country nic
						WHERE LOWER(nic.country_code) = LOWER(ctr.country_code)
					) AS n
				)
				ELSE NULL
			END AS numbeo_indices,
			CASE
				WHEN $3 THEN (
					SELECT jsonb_agg(
						jsonb_build_object(
							'pillar_name', li.pillar_name,
							'rank_2007', li.rank_2007,
							'rank_2008', li.rank_2008,
							'rank_2009', li.rank_2009,
							'rank_2010', li.rank_2010,
							'rank_2011', li.rank_2011,
							'rank_2012', li.rank_2012,
							'rank_2013', li.rank_2013,
							'rank_2014', li.rank_2014,
							'rank_2015', li.rank_2015,
							'rank_2016', li.rank_2016,
							'rank_2017', li.rank_2017,
							'rank_2018', li.rank_2018,
							'rank_2019', li.rank_2019,
							'rank_2020', li.rank_2020,
							'rank_2021', li.rank_2021,
							'rank_2022', li.rank_2022,
							'rank_2023', li.rank_2023,
							'score_2007', li.score_2007,
							'score_2008', li.score_2008,
							'score_2009', li.score_2009,
							'score_2010', li.score_2010,
							'score_2011', li.score_2011,
							'score_2012', li.score_2012,
							'score_2013', li.score_2013,
							'score_2014', li.score_2014,
							'score_2015', li.score_2015,
							'score_2016', li.score_2016,
							'score_2017', li.score_2017,
							'score_2018', li.score_2018,
							'score_2019', li.score_2019,
							'score_2020', li.score_2020,
							'score_2021', li.score_2021,
							'score_2022', li.score_2022,
							'score_2023', li.score_2023
						)
					)
					FROM legatum_index li
					WHERE LOWER(li.country_code) = LOWER(ctr.country_code)
				)
				ELSE NULL
			END AS legatum_indices
		FROM country ctr
		WHERE LOWER(ctr.country_code) = LOWER($1);`

	var (
		country     Country
		numbeoJSON  []byte
		legatumJSON []byte
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(
		ctx,
		query,
		countryCode,
		include.Has("numbeo_indices"),
		include.Has("legatum_indices"),
	).Scan(
		&country.Code,
		&country.Name,
		&numbeoJSON,
		&legatumJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if len(numbeoJSON) > 0 {
		var indices NumbeoCountryIndicies
		if err := json.Unmarshal(numbeoJSON, &indices); err != nil {
			return nil, err
		}
		country.NumbeoIndices = &indices
	}

	if len(legatumJSON) > 0 && string(legatumJSON) != "null" {
		indices, err := buildLegatumIndicesFromJSON(legatumJSON)
		if err != nil {
			return nil, err
		}
		country.LegatumIndices = indices
	}

	return &country, nil
}

func (c CountryModel) GetCountriesByCodes(codes []string) (countries []*Country, retErr error) {
	if len(codes) == 0 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT ctr.country_code, ctr.country
		FROM country ctr
		WHERE UPPER(ctr.country_code) = ANY($1)
		ORDER BY ctr.country_code;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(codes))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	countries = []*Country{}
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.Code, &country.Name); err != nil {
			return nil, err
		}
		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(countries) == 0 {
		return nil, ErrRecordNotFound
	}

	return countries, nil
}

type legatumRow struct {
	PillarName string `json:"pillar_name"`
	RankAndScore
}

func buildLegatumIndicesFromJSON(raw []byte) (*LegatumCountryIndices, error) {
	var rows []legatumRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	legatum := &LegatumCountryIndices{}
	for _, row := range rows {
		switch row.PillarName {
		case "Safety and Security":
			legatum.SafetyAndSecurity = row.RankAndScore
		case "Personal Freedom":
			legatum.PersonalFreedom = row.RankAndScore
		case "Governance":
			legatum.Governance = row.RankAndScore
		case "Social Capital":
			legatum.SocialCapital = row.RankAndScore
		case "Investment Environment":
			legatum.InvestmentEnvironment = row.RankAndScore
		case "Enterprise Conditions":
			legatum.EnterpriseConditions = row.RankAndScore
		case "Infrastructure and Market Access":
			legatum.InfrastructureAndMarketAccess = row.RankAndScore
		case "Economic Quality":
			legatum.EconomicQuality = row.RankAndScore
		case "Living Conditions":
			legatum.LivingConditions = row.RankAndScore
		case "Health":
			legatum.Health = row.RankAndScore
		case "Education":
			legatum.Education = row.RankAndScore
		case "Natural Environment":
			legatum.NaturalEnvironment = row.RankAndScore
		}
	}

	return legatum, nil
}
