CREATE INDEX idx_cities_country_code_lower
ON public.cities (LOWER(country_code));
