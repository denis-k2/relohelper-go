CREATE TABLE public.countries (
    country text NOT NULL,
    alpha2_code text,
    country_code text,
    numeric_code bigint,
    independent text,
    population bigint,
    area bigint,
    last_update date,
    CONSTRAINT countries_pkey PRIMARY KEY (country_code)
);

CREATE TABLE public.states (
    state_code text,
    state_name text NOT NULL,
    category text,
    CONSTRAINT states_pkey PRIMARY KEY (state_code)
);

CREATE TABLE public.cities (
    geoname_id bigint,
    city text NOT NULL,
    state_code text,
    country_code text NOT NULL,
    population bigint,
    latitude double precision,
    longitude double precision,
    timezone text,
    updated_date date,
    updated_by text,
    CONSTRAINT cities_pkey PRIMARY KEY (geoname_id),
    CONSTRAINT cities_country_code_fkey FOREIGN KEY (country_code)
        REFERENCES public.countries(country_code),
    CONSTRAINT cities_state_code_fkey FOREIGN KEY (state_code)
        REFERENCES public.states(state_code)
);

CREATE TABLE public.avg_climate (
    geoname_id bigint,
    month smallint,
    high_temp numeric,
    low_temp numeric,
    pressure numeric,
    wind_speed numeric,
    humidity numeric,
    rainfall numeric,
    rainfall_days numeric,
    snowfall numeric,
    snowfall_days numeric,
    sea_temp numeric,
    daylight numeric,
    sunshine numeric,
    sunshine_days numeric,
    uv_index numeric,
    cloud_cover numeric,
    visibility numeric,
    updated_date date,
    updated_by character varying(30),
    CONSTRAINT avg_climate_pkey PRIMARY KEY (geoname_id, month),
    CONSTRAINT avg_climate_geoname_id_fkey FOREIGN KEY (geoname_id)
        REFERENCES public.cities(geoname_id)
);

CREATE TABLE public.numbeo_cost_categories (
    category_id integer GENERATED ALWAYS AS IDENTITY,
    category character varying(100) NOT NULL,
    CONSTRAINT numbeo_cost_categories_pkey PRIMARY KEY (category_id),
    CONSTRAINT numbeo_cost_categories_category_key UNIQUE (category)
);

CREATE TABLE public.numbeo_cost_params (
    param_id integer GENERATED ALWAYS AS IDENTITY,
    category_id integer NOT NULL,
    param character varying(255) NOT NULL,
    CONSTRAINT numbeo_cost_params_pkey PRIMARY KEY (param_id),
    CONSTRAINT numbeo_cost_params_category_id_param_key UNIQUE (category_id, param),
    CONSTRAINT numbeo_cost_params_category_id_fkey FOREIGN KEY (category_id)
        REFERENCES public.numbeo_cost_categories(category_id)
);

CREATE TABLE public.numbeo_city_costs (
    geoname_id bigint,
    param_id integer,
    cost numeric,
    range numrange,
    last_update date,
    updated_date date,
    updated_by character varying(30),
    CONSTRAINT numbeo_city_costs_pkey PRIMARY KEY (geoname_id, param_id),
    CONSTRAINT numbeo_city_costs_geoname_id_fkey FOREIGN KEY (geoname_id)
        REFERENCES public.cities(geoname_id),
    CONSTRAINT numbeo_city_costs_param_id_fkey FOREIGN KEY (param_id)
        REFERENCES public.numbeo_cost_params(param_id)
);

CREATE TABLE public.numbeo_city_indices (
    geoname_id bigint,
    cost_of_living double precision,
    rent double precision,
    cost_of_living_plus_rent double precision,
    groceries double precision,
    local_purchasing_power double precision,
    quality_of_life double precision,
    property_price_to_income_ratio double precision,
    traffic_commute_time double precision,
    climate double precision,
    safety double precision,
    health_care double precision,
    pollution double precision,
    updated_date date,
    updated_by text,
    CONSTRAINT numbeo_city_indices_pkey PRIMARY KEY (geoname_id),
    CONSTRAINT numbeo_city_indices_geoname_id_fkey FOREIGN KEY (geoname_id)
        REFERENCES public.cities(geoname_id)
);

CREATE TABLE public.numbeo_country_indices (
    country_code text,
    cost_of_living double precision,
    rent double precision,
    cost_of_living_plus_rent double precision,
    groceries double precision,
    restaurant_price double precision,
    local_purchasing_power double precision,
    quality_of_life double precision,
    purchasing_power double precision,
    health_care double precision,
    property_price_to_income_ratio double precision,
    traffic_commute_time double precision,
    pollution double precision,
    climate double precision,
    avg_salary_usd double precision,
    safety double precision,
    updated_date date,
    updated_by text,
    CONSTRAINT numbeo_country_indices_pkey PRIMARY KEY (country_code),
    CONSTRAINT numbeo_country_indices_country_code_fkey FOREIGN KEY (country_code)
        REFERENCES public.countries(country_code)
);

CREATE TABLE public.legatum_country_indices (
    country_code text,
    area_group text,
    pillar_name text,
    rank_2007 bigint,
    rank_2008 bigint,
    rank_2009 bigint,
    rank_2010 bigint,
    rank_2011 bigint,
    rank_2012 bigint,
    rank_2013 bigint,
    rank_2014 bigint,
    rank_2015 bigint,
    rank_2016 bigint,
    rank_2017 bigint,
    rank_2018 bigint,
    rank_2019 bigint,
    rank_2020 bigint,
    rank_2021 bigint,
    rank_2022 bigint,
    rank_2023 bigint,
    score_2007 double precision,
    score_2008 double precision,
    score_2009 double precision,
    score_2010 double precision,
    score_2011 double precision,
    score_2012 double precision,
    score_2013 double precision,
    score_2014 double precision,
    score_2015 double precision,
    score_2016 double precision,
    score_2017 double precision,
    score_2018 double precision,
    score_2019 double precision,
    score_2020 double precision,
    score_2021 double precision,
    score_2022 double precision,
    score_2023 double precision,
    CONSTRAINT legatum_country_indices_pkey PRIMARY KEY (country_code, pillar_name),
    CONSTRAINT legatum_country_indices_country_code_fkey FOREIGN KEY (country_code)
        REFERENCES public.countries(country_code)
);
