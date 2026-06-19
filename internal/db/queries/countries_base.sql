-- name: ListCountries :many
SELECT country_code, country, population, area, last_update::text AS last_update
FROM countries
ORDER BY country_code;

-- name: GetCountry :one
SELECT ctr.country_code, ctr.country, ctr.population, ctr.area, ctr.last_update::text AS last_update
FROM countries ctr
WHERE ctr.country_code = @country_code::text;

-- name: GetCountriesByCodes :many
SELECT ctr.country_code, ctr.country, ctr.population, ctr.area, ctr.last_update::text AS last_update
FROM countries ctr
WHERE ctr.country_code = ANY(@country_codes::text[])
ORDER BY ctr.country_code;
