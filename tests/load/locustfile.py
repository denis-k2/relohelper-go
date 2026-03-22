from __future__ import annotations

import random
from dataclasses import dataclass
from typing import List

import requests
from locust import HttpUser, between, events, task

CITY_HEAVY_INCLUDE = "numbeo_cost,numbeo_indices,avg_climate"
COUNTRY_HEAVY_INCLUDE = "numbeo_indices,legatum_indices"

CITY_BATCH_MIN = 2
CITY_BATCH_MAX = 5
COUNTRY_BATCH_MIN = 2
COUNTRY_BATCH_MAX = 4
BOOTSTRAP_TIMEOUT_SECONDS = 15


@dataclass(frozen=True)
class LoadTestData:
    city_ids: List[int]
    city_country_codes: List[str]
    country_codes: List[str]


@events.test_start.add_listener
def bootstrap_api_data(environment, **_kwargs) -> None:
    host = environment.host
    if not host:
        raise RuntimeError(
            "Locust host is not set. Run with --host=http://127.0.0.1:4000 or set host in the UI."
        )

    session = requests.Session()

    countries_response = session.get(
        f"{host}/countries", timeout=BOOTSTRAP_TIMEOUT_SECONDS
    )
    countries_response.raise_for_status()
    countries_payload = countries_response.json()
    country_codes = [
        country["country_code"] for country in countries_payload["countries"]
    ]

    cities_response = session.get(f"{host}/cities", timeout=BOOTSTRAP_TIMEOUT_SECONDS)
    cities_response.raise_for_status()
    cities_payload = cities_response.json()
    cities = cities_payload["cities"]
    city_ids = [city["geoname_id"] for city in cities]
    city_country_codes = sorted({city["country_code"] for city in cities})

    if not country_codes:
        raise RuntimeError("Bootstrap failed: /countries returned no country codes")
    if not city_ids:
        raise RuntimeError("Bootstrap failed: /cities returned no city ids")
    if not city_country_codes:
        raise RuntimeError(
            "Bootstrap failed: /cities returned no country codes with cities"
        )

    environment.relohelper_data = LoadTestData(
        city_ids=city_ids,
        city_country_codes=city_country_codes,
        country_codes=country_codes,
    )


class RelohelperApiUser(HttpUser):
    wait_time = between(1, 3)

    @property
    def data(self) -> LoadTestData:
        data = getattr(self.environment, "relohelper_data", None)
        if data is None:
            raise RuntimeError("Load test data has not been initialized")
        return data

    def random_city_id(self) -> int:
        return random.choice(self.data.city_ids)

    def random_country_code(self) -> str:
        return random.choice(self.data.country_codes)

    def random_city_country_code(self) -> str:
        return random.choice(self.data.city_country_codes)

    def random_city_batch_ids(self) -> str:
        batch_size = random.randint(CITY_BATCH_MIN, CITY_BATCH_MAX)
        selected_ids = random.sample(self.data.city_ids, k=batch_size)
        return ",".join(str(city_id) for city_id in selected_ids)

    def random_country_batch_codes(self) -> str:
        batch_size = random.randint(COUNTRY_BATCH_MIN, COUNTRY_BATCH_MAX)
        selected_codes = random.sample(self.data.country_codes, k=batch_size)
        return ",".join(selected_codes)

    @task(15)
    def cities_list_by_country_code(self) -> None:
        self.client.get(
            "/cities",
            name="GET /cities?country_code",
            params={"country_code": self.random_city_country_code()},
        )

    @task(15)
    def cities_detail(self) -> None:
        city_id = self.random_city_id()
        self.client.get(f"/cities/{city_id}", name="GET /cities/{id}")

    @task(20)
    def cities_detail_with_heavy_include(self) -> None:
        city_id = self.random_city_id()
        self.client.get(
            f"/cities/{city_id}",
            name="GET /cities/{id}?include=heavy",
            params={"include": CITY_HEAVY_INCLUDE},
        )

    @task(15)
    def cities_batch(self) -> None:
        self.client.get(
            "/cities",
            name="GET /cities?ids",
            params={"ids": self.random_city_batch_ids()},
        )

    @task(20)
    def cities_batch_with_heavy_include(self) -> None:
        self.client.get(
            "/cities",
            name="GET /cities?ids&include=heavy",
            params={
                "ids": self.random_city_batch_ids(),
                "include": CITY_HEAVY_INCLUDE,
            },
        )

    @task(5)
    def countries_detail(self) -> None:
        country_code = self.random_country_code()
        self.client.get(
            f"/countries/{country_code}", name="GET /countries/{country_code}"
        )

    @task(5)
    def countries_detail_with_heavy_include(self) -> None:
        country_code = self.random_country_code()
        self.client.get(
            f"/countries/{country_code}",
            name="GET /countries/{country_code}?include=heavy",
            params={"include": COUNTRY_HEAVY_INCLUDE},
        )

    @task(5)
    def countries_batch(self) -> None:
        self.client.get(
            "/countries",
            name="GET /countries?country_codes",
            params={"country_codes": self.random_country_batch_codes()},
        )

    @task(10)
    def countries_batch_with_heavy_include(self) -> None:
        self.client.get(
            "/countries",
            name="GET /countries?country_codes&include=heavy",
            params={
                "country_codes": self.random_country_batch_codes(),
                "include": COUNTRY_HEAVY_INCLUDE,
            },
        )
