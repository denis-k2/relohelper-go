-- name: GetNumbeoCountryIndicesByCodes :many
SELECT
    nic.country_code,
    jsonb_build_object(
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
    ) AS numbeo_indices
FROM numbeo_country_indices nic
WHERE nic.country_code = ANY(@country_codes::text[]);
