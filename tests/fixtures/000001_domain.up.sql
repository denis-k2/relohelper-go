TRUNCATE TABLE
    avg_climate,
    numbeo_city_costs,
    numbeo_city_indices,
    numbeo_country_indices,
    legatum_country_indices,
    numbeo_cost_params,
    numbeo_cost_categories,
    cities,
    states,
    countries
RESTART IDENTITY CASCADE;

INSERT INTO countries (
    country,
    alpha2_code,
    country_code,
    population,
    area,
    last_update
)
VALUES
    ('Argentina', 'AR', 'ARG', 45800000, 2780400, DATE '2026-01-01'),
    ('Australia', 'AU', 'AUS', 26700000, 7692024, DATE '2026-01-01'),
    ('Brazil', 'BR', 'BRA', 203000000, 8515767, DATE '2026-01-01'),
    ('Canada', 'CA', 'CAN', 41000000, 9984670, DATE '2026-01-01'),
    ('Chile', 'CL', 'CHL', 19700000, 756102, DATE '2026-01-01'),
    ('China', 'CN', 'CHN', 1408000000, 9596961, DATE '2026-01-01'),
    ('Czechia', 'CZ', 'CZE', 10900000, 78871, DATE '2026-01-01'),
    ('Germany', 'DE', 'DEU', 83500000, 357022, DATE '2026-01-01'),
    ('United Kingdom', 'GB', 'GBR', 68300000, 243610, DATE '2026-01-01'),
    ('Italy', 'IT', 'ITA', 58900000, 301340, DATE '2026-01-01'),
    ('Japan', 'JP', 'JPN', 123000000, 377975, DATE '2026-01-01'),
    ('Morocco', 'MA', 'MAR', 38000000, 446550, DATE '2026-01-01'),
    ('Netherlands', 'NL', 'NLD', 18000000, 41850, DATE '2026-01-01'),
    ('Russia', 'RU', 'RUS', 146000000, 17098246, DATE '2026-01-01'),
    ('Thailand', 'TH', 'THA', 71600000, 513120, DATE '2026-01-01'),
    ('United States of America', 'US', 'USA', 340000000, 9833517, DATE '2026-01-01'),
    ('Wallis and Futuna', 'WF', 'WLF', 12000, 142, DATE '2026-01-01');

INSERT INTO cities (
    geoname_id,
    city,
    country_code,
    population,
    latitude,
    longitude,
    timezone,
    updated_date,
    updated_by
)
VALUES
    (524901, 'Moscow', 'RUS', 13000000, 55.7522, 37.6156, 'Europe/Moscow', DATE '2026-01-01', 'test'),
    (1850147, 'Tokyo', 'JPN', 14000000, 35.6895, 139.6917, 'Asia/Tokyo', DATE '2026-01-01', 'test'),
    (2147714, 'Sydney', 'AUS', 5300000, -33.8679, 151.2073, 'Australia/Sydney', DATE '2026-01-01', 'test'),
    (2542997, 'Marrakesh', 'MAR', 930000, 31.6342, -7.9999, 'Africa/Casablanca', DATE '2026-01-01', 'test'),
    (2562305, 'Casablanca', 'MAR', 3350000, 33.5883, -7.6114, 'Africa/Casablanca', DATE '2026-01-01', 'test'),
    (2643743, 'London', 'GBR', 8900000, 51.5085, -0.1257, 'Europe/London', DATE '2026-01-01', 'test'),
    (3069011, 'Prague', 'CZE', 1380000, 50.0880, 14.4208, 'Europe/Prague', DATE '2026-01-01', 'test'),
    (3871336, 'Santiago', 'CHL', 6200000, -33.4569, -70.6483, 'America/Santiago', DATE '2026-01-01', 'test'),
    (5128581, 'New York City', 'USA', 8300000, 40.7143, -74.0060, 'America/New_York', DATE '2026-01-01', 'test'),
    (5378538, 'Oakland', 'USA', 440000, 37.8044, -122.2708, 'America/Los_Angeles', DATE '2026-01-01', 'test'),
    (5809844, 'Seattle', 'USA', 755000, 47.6062, -122.3321, 'America/Los_Angeles', DATE '2026-01-01', 'test'),
    (6167865, 'Toronto', 'CAN', 2930000, 43.7001, -79.4163, 'America/Toronto', DATE '2026-01-01', 'test');

INSERT INTO numbeo_cost_categories (category)
VALUES ('Restaurants');

INSERT INTO numbeo_cost_params (category_id, param)
SELECT category_id, 'Meal, Inexpensive Restaurant'
FROM numbeo_cost_categories
WHERE category = 'Restaurants';

INSERT INTO numbeo_city_costs (
    geoname_id,
    param_id,
    cost,
    range,
    last_update,
    updated_date,
    updated_by
)
SELECT
    city.geoname_id,
    param.param_id,
    city.cost,
    numrange(city.cost - 2, city.cost + 2, '[]'),
    DATE '2026-01-01',
    DATE '2026-01-01',
    'test'
FROM (
    VALUES
        (1850147::bigint, 8::numeric),
        (2542997::bigint, 6::numeric),
        (2562305::bigint, 7::numeric),
        (5378538::bigint, 20::numeric)
) AS city(geoname_id, cost)
CROSS JOIN numbeo_cost_params AS param;

INSERT INTO numbeo_city_indices (
    geoname_id,
    cost_of_living,
    rent,
    cost_of_living_plus_rent,
    groceries,
    local_purchasing_power,
    quality_of_life,
    climate,
    safety,
    health_care,
    pollution,
    updated_date,
    updated_by
)
VALUES
    (1850147, 55, 30, 45, 60, 80, 175, 85, 75, 82, 40, DATE '2026-01-01', 'test'),
    (2542997, 35, 15, 25, 30, 40, 140, 90, 65, 55, 45, DATE '2026-01-01', 'test'),
    (2562305, 38, 18, 28, 33, 42, 135, 88, 60, 58, 50, DATE '2026-01-01', 'test'),
    (3069011, 48, 25, 37, 45, 70, 165, 80, 72, 76, 35, DATE '2026-01-01', 'test');

INSERT INTO avg_climate (
    geoname_id,
    month,
    high_temp,
    low_temp,
    pressure,
    wind_speed,
    humidity,
    rainfall,
    rainfall_days,
    snowfall,
    snowfall_days,
    sea_temp,
    daylight,
    sunshine,
    sunshine_days,
    uv_index,
    cloud_cover,
    visibility,
    updated_date,
    updated_by
)
SELECT
    city.geoname_id,
    month.number,
    10 + month.number,
    2 + month.number,
    1010 + month.number,
    8 + month.number / 10.0,
    60 + month.number,
    20 + month.number,
    5 + month.number / 10.0,
    0,
    0,
    CASE WHEN city.has_sea_temp THEN 15 + month.number / 2.0 ELSE NULL END,
    8 + month.number / 2.0,
    4 + month.number / 3.0,
    10 + month.number,
    2 + month.number / 2.0,
    40 + month.number,
    8 + month.number / 10.0,
    DATE '2026-01-01',
    'test'
FROM (
    VALUES
        (1850147::bigint, false),
        (2542997::bigint, true)
) AS city(geoname_id, has_sea_temp)
CROSS JOIN generate_series(1, 12) AS month(number);

INSERT INTO numbeo_country_indices (
    country_code,
    cost_of_living,
    rent,
    cost_of_living_plus_rent,
    groceries,
    restaurant_price,
    local_purchasing_power,
    quality_of_life,
    climate,
    safety,
    health_care,
    pollution,
    avg_salary_usd,
    updated_date,
    updated_by
)
VALUES
    ('BRA', 35, 12, 25, 30, 28, 40, 125, 85, 45, 55, 60, 600, DATE '2026-01-01', 'test'),
    ('CAN', 65, 40, 55, 70, 68, 90, 180, 60, 70, 78, 30, 3200, DATE '2026-01-01', 'test'),
    ('CHN', 40, 20, 32, 38, 30, 75, 135, 70, 72, 68, 65, 1400, DATE '2026-01-01', 'test'),
    ('RUS', 38, 18, 29, 35, 32, 50, 130, 45, 60, 62, 55, 900, DATE '2026-01-01', 'test'),
    ('USA', 72, 55, 65, 75, 78, 105, 185, 72, 68, 70, 42, 4500, DATE '2026-01-01', 'test');

INSERT INTO legatum_country_indices (
    country_code,
    area_group,
    pillar_name,
    rank_2023,
    score_2023
)
VALUES
    ('BRA', 'Americas', 'Safety and Security', 75, 55.5),
    ('CHN', 'Asia-Pacific', 'Safety and Security', 60, 60.0),
    ('THA', 'Asia-Pacific', 'Safety and Security', 70, 57.5),
    ('USA', 'Americas', 'Safety and Security', 30, 72.0);
