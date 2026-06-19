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
