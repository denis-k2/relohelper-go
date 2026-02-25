import random

import requests  # type: ignore
from locust import HttpUser, between, events, task  # type: ignore

@events.test_start.add_listener
def on_locust_init(environment, **kwargs):
    response = requests.get("http://195.133.194.151:4000/countries", timeout=5)
    countries = [country["country_code"] for country in response.json()["countries"]]
    environment.parsed_options.country_codes = countries

    # response = requests.post(
    #     "http://localhost:4000/tokens/authentication",
    #     json={"email": "faith@example.com", "password": "pa55word"},
    #     headers={
    #         "accept": "application/json",
    #     },
    # )
    # token = response.json()["authentication_token"]["token"]
    # print("**********", token)
    # environment.parsed_options.token = token

class User(HttpUser):
    host = "http://195.133.194.151:4000"
    wait_time = between(1, 5)

    @task(15)
    def read_city_id(self):
        city_id = random.randint(1, 534)  # noqa: S311
        self.client.get(
            f"/cities/{city_id}",
            params={"numbeo_cost": True, "numbeo_indices": True, "avg_climate": True},
            # headers={"Authorization": f"Bearer {self.environment.parsed_options.token}"},
            name="/cities/city_id?all",
        )

    @task(10)
    def read_country_id(self):
        country_code = random.choice(self.environment.parsed_options.country_codes)  # noqa: S311
        self.client.get(
            f"/countries/{country_code}",
            params={"numbeo_indices": True, "legatum_indices": True},
            # headers={"Authorization": f"Bearer {self.environment.parsed_options.token}"},
            name="/countries/country_id?all",
        )

    @task
    def read_countries(self):
        self.client.get("/countries", name="/countries")

    @task
    def read_cities(self):
        self.client.get("/cities", name="/cities")
