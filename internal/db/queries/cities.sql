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
WITH climate_rows AS (
    SELECT *
    FROM avg_climate ac
    WHERE ac.geoname_id = ANY(@geoname_ids::bigint[])
),
climate_stats AS (
    SELECT
        geoname_id,
        COUNT(*) AS row_count,
        COUNT(DISTINCT month) AS unique_month_count,
        MIN(month) AS min_month,
        MAX(month) AS max_month
    FROM climate_rows
    GROUP BY geoname_id
),
climate_data AS (
    SELECT
        cr.geoname_id,
        m.metric_key,
        jsonb_agg(m.metric_value ORDER BY cr.month) AS month_values
    FROM climate_rows cr
    CROSS JOIN LATERAL jsonb_each(
        to_jsonb(cr)
            - 'geoname_id'
            - 'month'
            - 'updated_date'
            - 'updated_by'
    ) AS m(metric_key, metric_value)
    GROUP BY cr.geoname_id, m.metric_key
)
SELECT
    s.geoname_id,
    CASE
        WHEN s.row_count = 12
            AND s.unique_month_count = 12
            AND s.min_month = 1
            AND s.max_month = 12
        THEN (
            SELECT jsonb_object_agg(cd.metric_key, cd.month_values)
            FROM climate_data cd
            WHERE cd.geoname_id = s.geoname_id
        )
        ELSE jsonb_build_object('__invalid_structure__', true)
    END AS avg_climate
FROM climate_stats s;
