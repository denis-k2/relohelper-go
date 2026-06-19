-- name: GetNumbeoCostByCityIDs :many
-- Builds the detailed per-city cost payload expected by the API. The response
-- order is intentionally stable by category and parameter display names.
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
