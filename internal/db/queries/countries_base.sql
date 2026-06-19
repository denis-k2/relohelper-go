-- name: ListCountries :many
SELECT country_code, country, population, area, last_update::text AS last_update
FROM countries
ORDER BY country_code;

-- name: GetCountryDetailed :one
-- Detail endpoint query. Optional include blocks are guarded by boolean
-- parameters so unrequested JSON payloads are not built.
SELECT
    ctr.country_code,
    ctr.country,
    ctr.population,
    ctr.area,
    ctr.last_update::text AS last_update,
    CASE
        WHEN @include_numbeo_indices::boolean THEN (
            SELECT jsonb_build_object(
                'cost_of_living', nic.cost_of_living,
                'rent', nic.rent,
                'cost_of_living_plus_rent', nic.cost_of_living_plus_rent,
                'groceries', nic.groceries,
                'restaurant_price', nic.restaurant_price,
                'local_purchasing_power', nic.local_purchasing_power,
                'quality_of_life', nic.quality_of_life,
                'property_price_to_income_ratio', nic.property_price_to_income_ratio,
                'traffic_commute_time', nic.traffic_commute_time,
                'climate', nic.climate,
                'safety', nic.safety,
                'health_care', nic.health_care,
                'pollution', nic.pollution,
                'avg_salary_usd', nic.avg_salary_usd,
                'last_update', to_char(nic.updated_date, 'YYYY-MM-DD')
            )
            FROM numbeo_country_indices nic
            WHERE nic.country_code = ctr.country_code
        )
        ELSE NULL
    END AS numbeo_indices,
    CASE
        WHEN @include_legatum_indices::boolean THEN (
            SELECT jsonb_object_agg(l.key, l.value)
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
            WHERE li.country_code = ctr.country_code AND l.key IS NOT NULL
        )
        ELSE NULL
    END AS legatum_indices
FROM countries ctr
WHERE ctr.country_code = @country_code::text;

-- name: GetCountriesByCodes :many
SELECT ctr.country_code, ctr.country, ctr.population, ctr.area, ctr.last_update::text AS last_update
FROM countries ctr
WHERE ctr.country_code = ANY(@country_codes::text[])
ORDER BY ctr.country_code;
