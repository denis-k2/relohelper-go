-- name: GetAvgClimateByCityIDs :many
-- Builds one 12-month JSON series per climate metric. The explicit aggregates
-- avoid expanding rows through jsonb_each and keep the query to one grouped
-- pass over avg_climate.
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
