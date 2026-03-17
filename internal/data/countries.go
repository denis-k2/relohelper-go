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
	Code                  string                 `json:"country_code"`
	Name                  string                 `json:"country"`
	NumbeoCountryIndices  *NumbeoCountryIndices  `json:"numbeo_indices,omitempty"`
	LegatumCountryIndices *LegatumCountryIndices `json:"legatum_indices,omitempty"`
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
	DB *sql.DB
}

func (c CountryModel) ListCountries() (countries []*Country, retErr error) {
	query := `
		SELECT country_code, country
		FROM countries
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
							to_char(nic.updated_date, 'YYYY-MM-DD') AS last_update
						FROM numbeo_country_indices nic
						WHERE nic.country_code = ctr.country_code
					) AS n
				)
				ELSE NULL
			END AS numbeo_indices,
			CASE
				WHEN $3 THEN (
					SELECT jsonb_object_agg(l.key, l.value)
					FROM (
						SELECT
							CASE li.pillar_name
								WHEN 'Safety and Security' THEN 'safety_and_security'
								WHEN 'Personal Freedom' THEN 'personal_freedom'
								WHEN 'Governance' THEN 'governance'
								WHEN 'Social Capital' THEN 'social_capital'
								WHEN 'Investment Environment' THEN 'investment_invironment'
								WHEN 'Enterprise Conditions' THEN 'enterprise_conditions'
								WHEN 'Infrastructure and Market Access' THEN 'infrastructure_and_market_access'
								WHEN 'Economic Quality' THEN 'economic_quality'
								WHEN 'Living Conditions' THEN 'living_conditions'
								WHEN 'Health' THEN 'health'
								WHEN 'Education' THEN 'education'
								WHEN 'Natural Environment' THEN 'natural_environment'
							END AS key,
							to_jsonb(li) - 'country_code' - 'pillar_name' AS value
						FROM legatum_country_indices li
						WHERE li.country_code = ctr.country_code
					) AS l
					WHERE l.key IS NOT NULL
				)
				ELSE NULL
			END AS legatum_indices
		FROM countries ctr
		WHERE ctr.country_code = $1;`

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
		var indices NumbeoCountryIndices
		if err := json.Unmarshal(numbeoJSON, &indices); err != nil {
			return nil, err
		}
		country.NumbeoCountryIndices = &indices
	}

	if len(legatumJSON) > 0 && string(legatumJSON) != "null" {
		var indices LegatumCountryIndices
		if err := json.Unmarshal(legatumJSON, &indices); err != nil {
			return nil, err
		}
		country.LegatumCountryIndices = &indices
	}

	return &country, nil
}

func (c CountryModel) GetCountriesByCodes(codes []string, include IncludeSet) (countries []*Country, retErr error) {
	if len(codes) == 0 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT ctr.country_code, ctr.country
		FROM countries ctr
		WHERE ctr.country_code = ANY($1)
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
	countryByCode := make(map[string]*Country, len(codes))
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.Code, &country.Name); err != nil {
			return nil, err
		}
		countryPtr := &country
		countries = append(countries, countryPtr)
		countryByCode[country.Code] = countryPtr
	}

	if err = rows.Err(); err != nil {
		return nil, err
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

func (c CountryModel) attachNumbeoIndicesByCodes(ctx context.Context, codes []string, countryByCode map[string]*Country) (retErr error) {
	query := `
		SELECT
			nic.country_code,
			row_to_json(n) AS numbeo_indices
		FROM numbeo_country_indices nic
		CROSS JOIN LATERAL (
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
				to_char(nic.updated_date, 'YYYY-MM-DD') AS last_update
		) AS n
		WHERE nic.country_code = ANY($1);`

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(codes))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	for rows.Next() {
		var (
			countryCode string
			rawJSON     []byte
		)

		if err := rows.Scan(&countryCode, &rawJSON); err != nil {
			return err
		}
		if len(rawJSON) == 0 || string(rawJSON) == "null" {
			continue
		}

		var indices NumbeoCountryIndices
		if err := json.Unmarshal(rawJSON, &indices); err != nil {
			return err
		}

		country, ok := countryByCode[countryCode]
		if !ok {
			continue
		}
		country.NumbeoCountryIndices = &indices
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (c CountryModel) attachLegatumIndicesByCodes(ctx context.Context, codes []string, countryByCode map[string]*Country) (retErr error) {
	query := `
		SELECT
			li.country_code,
			jsonb_object_agg(l.key, l.value) AS legatum_indices
		FROM legatum_country_indices li
		CROSS JOIN LATERAL (
			SELECT
				CASE li.pillar_name
					WHEN 'Safety and Security' THEN 'safety_and_security'
					WHEN 'Personal Freedom' THEN 'personal_freedom'
					WHEN 'Governance' THEN 'governance'
					WHEN 'Social Capital' THEN 'social_capital'
					WHEN 'Investment Environment' THEN 'investment_invironment'
					WHEN 'Enterprise Conditions' THEN 'enterprise_conditions'
					WHEN 'Infrastructure and Market Access' THEN 'infrastructure_and_market_access'
					WHEN 'Economic Quality' THEN 'economic_quality'
					WHEN 'Living Conditions' THEN 'living_conditions'
					WHEN 'Health' THEN 'health'
					WHEN 'Education' THEN 'education'
					WHEN 'Natural Environment' THEN 'natural_environment'
				END AS key,
				to_jsonb(li) - 'country_code' - 'pillar_name' AS value
		) AS l
		WHERE li.country_code = ANY($1) AND l.key IS NOT NULL
		GROUP BY li.country_code;`

	rows, err := c.DB.QueryContext(ctx, query, pq.Array(codes))
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	for rows.Next() {
		var (
			countryCode string
			rawJSON     []byte
		)

		if err := rows.Scan(&countryCode, &rawJSON); err != nil {
			return err
		}
		if len(rawJSON) == 0 || string(rawJSON) == "null" {
			continue
		}

		var indices LegatumCountryIndices
		if err := json.Unmarshal(rawJSON, &indices); err != nil {
			return err
		}

		country, ok := countryByCode[countryCode]
		if !ok {
			continue
		}
		country.LegatumCountryIndices = &indices
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
