import random

import requests  # type: ignore
from locust import HttpUser, between, events, task  # type: ignore


@events.test_start.add_listener
def on_locust_init(environment, **kwargs):
    response = requests.get("http://195.133.194.151:8000/country", timeout=5)
    countries = [country["country_code"] for country in response.json()]
    environment.parsed_options.country_codes = countries


class User(HttpUser):
    host = "http://195.133.194.151:8000"
    wait_time = between(1, 5)

    # def on_start(self):
    #     response = self.client.post(
    #         "/login",
    #         data={"username": "string", "password": "string"},
    #         headers={
    #             "accept": "application/json",
    #             "Content-Type": "application/x-www-form-urlencoded",
    #         },
    #     )
    #     self.token = response.json()["access_token"]

    @task(15)
    def read_city_id(self):
        city_id = random.randint(1, 534)  # noqa: S311
        self.client.get(
            f"/city/{city_id}",
            params={"numbeo_cost": True, "numbeo_indices": True, "avg_climate": True},
            # headers={"Authorization": f"Bearer {self.token}"},
            name="/city/city_id?all",
        )

    @task(10)
    def read_country_id(self):
        country_code = random.choice(self.environment.parsed_options.country_codes)  # noqa: S311
        self.client.get(
            f"/country/{country_code}",
            params={"numbeo_indices": True, "legatum_indices": True},
            # headers={"Authorization": f"Bearer {self.token}"},
            name="/country/country_id?all",
        )

    @task
    def read_countries(self):
        self.client.get("/country", name="/country")

    @task
    def read_cities(self):
        self.client.get("/city", name="/city")
