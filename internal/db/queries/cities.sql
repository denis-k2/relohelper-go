-- name: ListCities :many
SELECT c.geoname_id, c.city, c.state_code, c.country_code,
       ctr.country AS country, c.population, c.latitude, c.longitude, c.timezone,
       to_char(c.updated_date, 'YYYY-MM-DD') AS last_update
FROM cities c
LEFT JOIN countries ctr ON ctr.country_code = c.country_code
WHERE (LOWER(c.country_code) = LOWER(@country_code::text) OR @country_code::text = '')
ORDER BY c.geoname_id;

-- name: GetCity :one
SELECT c.geoname_id, c.city, c.state_code, c.country_code,
       ctr.country AS country, c.population, c.latitude, c.longitude, c.timezone,
       to_char(c.updated_date, 'YYYY-MM-DD') AS last_update
FROM cities c
LEFT JOIN countries ctr ON ctr.country_code = c.country_code
WHERE c.geoname_id = @geoname_id::bigint;

-- name: GetCitiesByIDs :many
SELECT c.geoname_id, c.city, c.state_code, c.country_code,
       ctr.country AS country, c.population, c.latitude, c.longitude, c.timezone,
       to_char(c.updated_date, 'YYYY-MM-DD') AS last_update
FROM cities c
LEFT JOIN countries ctr ON ctr.country_code = c.country_code
WHERE c.geoname_id = ANY(@geoname_ids::bigint[])
ORDER BY c.geoname_id;

-- name: GetNumbeoCostByCityIDs :many
SELECT
    ns.geoname_id,
    jsonb_build_object(
        'currency', 'USD',
        'last_update', MAX(ns.last_update)::text,
        'prices', jsonb_agg(
            jsonb_build_object(
                'category', nc.category,
                'param', np.param,
                'cost', ns.cost,
                'range_lower', lower(ns.range),
                'range_upper', upper(ns.range)
            )
            ORDER BY nc.category, np.param
        )
    ) AS numbeo_cost
FROM numbeo_city_costs ns
JOIN numbeo_cost_params np ON np.param_id = ns.param_id
JOIN numbeo_cost_categories nc ON nc.category_id = np.category_id
WHERE ns.geoname_id = ANY(@geoname_ids::bigint[])
GROUP BY ns.geoname_id;

-- name: GetNumbeoCityIndicesByCityIDs :many
SELECT
    nic.geoname_id,
    jsonb_build_object(
        'cost_of_living', nic.cost_of_living,
        'rent', nic.rent,
        'cost_of_living_plus_rent', nic.cost_of_living_plus_rent,
        'groceries', nic.groceries,
        'local_purchasing_power', nic.local_purchasing_power,
        'quality_of_life', nic.quality_of_life,
        'property_price_to_income_ratio', nic.property_price_to_income_ratio,
        'traffic_commute_time', nic.traffic_commute_time,
        'climate', nic.climate,
        'safety', nic.safety,
        'health_care', nic.health_care,
        'pollution', nic.pollution,
        'last_update', to_char(nic.updated_date, 'YYYY-MM-DD')
    ) AS numbeo_indices
FROM numbeo_city_indices nic
WHERE nic.geoname_id = ANY(@geoname_ids::bigint[]);

-- name: GetAvgClimateByCityIDs :many
SELECT
    ac.geoname_id,
    CASE
        WHEN COUNT(*) = 12
            AND COUNT(DISTINCT ac.month) = 12
            AND MIN(ac.month) = 1
            AND MAX(ac.month) = 12
        THEN jsonb_build_object(
            'high_temp', jsonb_agg(ac.high_temp ORDER BY ac.month),
            'low_temp', jsonb_agg(ac.low_temp ORDER BY ac.month),
            'pressure', jsonb_agg(ac.pressure ORDER BY ac.month),
            'wind_speed', jsonb_agg(ac.wind_speed ORDER BY ac.month),
            'humidity', jsonb_agg(ac.humidity ORDER BY ac.month),
            'rainfall', jsonb_agg(ac.rainfall ORDER BY ac.month),
            'rainfall_days', jsonb_agg(ac.rainfall_days ORDER BY ac.month),
            'snowfall', jsonb_agg(ac.snowfall ORDER BY ac.month),
            'snowfall_days', jsonb_agg(ac.snowfall_days ORDER BY ac.month),
            'sea_temp', jsonb_agg(ac.sea_temp ORDER BY ac.month),
            'daylight', jsonb_agg(ac.daylight ORDER BY ac.month),
            'sunshine', jsonb_agg(ac.sunshine ORDER BY ac.month),
            'sunshine_days', jsonb_agg(ac.sunshine_days ORDER BY ac.month),
            'uv_index', jsonb_agg(ac.uv_index ORDER BY ac.month),
            'cloud_cover', jsonb_agg(ac.cloud_cover ORDER BY ac.month),
            'visibility', jsonb_agg(ac.visibility ORDER BY ac.month)
        )
        ELSE jsonb_build_object('__invalid_structure__', true)
    END AS avg_climate
FROM avg_climate ac
WHERE ac.geoname_id = ANY(@geoname_ids::bigint[])
GROUP BY ac.geoname_id;
