-- name: GetLegatumCountryIndicesByCodes :many
-- Converts Legatum pillar names into API field names and aggregates them into
-- a single JSON object per country.
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
WHERE li.country_code = ANY(@country_codes::text[]) AND l.key IS NOT NULL
GROUP BY li.country_code;
