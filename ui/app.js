const state = {
  allCities: [],
  countries: [],
  cities: [],
  selectedCountryCodes: new Set(),
  selectedCityIds: new Set(),
  comparisonData: [],
  climateChart: null,
};

const els = {
  filtersCard: document.querySelector(".filters-card"),
  filtersBody: document.getElementById("filtersBody"),
  editFiltersBtn: document.getElementById("editFiltersBtn"),
  countrySearch: document.getElementById("countrySearch"),
  citySearch: document.getElementById("citySearch"),
  countryList: document.getElementById("countryList"),
  cityList: document.getElementById("cityList"),
  selectedCountries: document.getElementById("selectedCountries"),
  selectedCities: document.getElementById("selectedCities"),
  compareBtn: document.getElementById("compareBtn"),
  resetBtn: document.getElementById("resetBtn"),
  tableEmptyState: document.getElementById("tableEmptyState"),
  chartEmptyState: document.getElementById("chartEmptyState"),
  comparisonTableWrapper: document.getElementById("comparisonTableWrapper"),
  comparisonTableBody: document.querySelector("#comparisonTable tbody"),
  metricCities: document.getElementById("metricCities"),
  metricCountries: document.getElementById("metricCountries"),
  metricLowestRent: document.getElementById("metricLowestRent"),
  metricBestQol: document.getElementById("metricBestQol"),
  climateCanvas: document.getElementById("climateChart"),
};

document.addEventListener("DOMContentLoaded", () => {
  bindEvents();
  loadInitialData();
});

function bindEvents() {
  els.countrySearch.addEventListener("input", renderCountries);
  els.citySearch.addEventListener("input", renderCities);
  els.compareBtn.addEventListener("click", loadComparison);
  els.resetBtn.addEventListener("click", resetDashboard);
  els.editFiltersBtn.addEventListener("click", expandFilters);
}

async function loadInitialData() {
  try {
    const res = await fetch("/cities");
    if (!res.ok) throw new Error(`Failed to load cities: ${res.status}`);

    const data = await res.json();
    state.allCities = Array.isArray(data.cities)
      ? data.cities
      : Array.isArray(data.data)
        ? data.data
        : [];
    state.countries = buildCountriesFromCities(state.allCities);

    renderCountries();
    renderCities();
    renderSelectedCountries();
    renderSelectedCities();
    updateMetrics();
  } catch (error) {
    console.error(error);
    els.countryList.innerHTML = `<div class="empty-state">Failed to load countries.</div>`;
    els.cityList.innerHTML = `<div class="empty-state">Failed to load cities.</div>`;
  }
}

function buildCountriesFromCities(cities) {
  const countryMap = new Map();

  for (const city of cities) {
    const code = city.country_code ?? city.code;
    const name = city.country ?? city.country_name ?? code;
    if (!code || !name || countryMap.has(code)) continue;

    countryMap.set(code, {
      country_code: code,
      country: name,
    });
  }

  return Array.from(countryMap.values()).sort((a, b) =>
    a.country.localeCompare(b.country),
  );
}

function filterCitiesForSelectedCountries() {
  const selectedCodes = Array.from(state.selectedCountryCodes);

  if (selectedCodes.length === 0) {
    state.cities = [];
    state.selectedCityIds.clear();
    renderCities();
    renderSelectedCities();
    updateMetrics();
    return;
  }

  const selectedCodeSet = new Set(selectedCodes);
  state.cities = state.allCities
    .filter((city) => selectedCodeSet.has(city.country_code))
    .sort((a, b) => {
      const aName = (a.city ?? a.name ?? "").toLowerCase();
      const bName = (b.city ?? b.name ?? "").toLowerCase();
      return aName.localeCompare(bName);
    });

  const availableCityIDs = new Set(
    state.cities.map((city) =>
      String(city.geoname_id ?? city.city_id ?? city.id ?? ""),
    ),
  );

  for (const cityID of Array.from(state.selectedCityIds)) {
    if (!availableCityIDs.has(cityID)) {
      state.selectedCityIds.delete(cityID);
    }
  }

  renderCities();
  renderSelectedCities();
  updateMetrics();
}

function renderCountries() {
  const q = els.countrySearch.value.trim().toLowerCase();

  const filtered = state.countries.filter((country) => {
    const code = (country.country_code ?? country.code ?? "").toLowerCase();
    const name = (
      country.country ??
      country.country_name ??
      country.name ??
      ""
    ).toLowerCase();
    return name.includes(q) || code.includes(q);
  });

  if (filtered.length === 0) {
    els.countryList.innerHTML = `<div class="empty-state">No countries found.</div>`;
    return;
  }

  els.countryList.innerHTML = filtered
    .map((country) => {
      const code = country.country_code ?? country.code ?? "";
      const name =
        country.country ?? country.country_name ?? country.name ?? code;
      const checked = state.selectedCountryCodes.has(code) ? "checked" : "";

      return `
        <label class="selection-item">
          <input type="checkbox" data-country-code="${escapeHtml(code)}" ${checked} />
          <div class="selection-item-main">
            <span class="selection-title">${escapeHtml(name)}</span>
            <span class="selection-subtitle">${escapeHtml(code)}</span>
          </div>
        </label>
      `;
    })
    .join("");

  els.countryList
    .querySelectorAll("input[data-country-code]")
    .forEach((input) => {
      input.addEventListener("change", (e) => {
        const code = e.target.dataset.countryCode;
        if (!code) return;

        if (e.target.checked) {
          state.selectedCountryCodes.add(code);
        } else {
          state.selectedCountryCodes.delete(code);
        }

        renderSelectedCountries();
        filterCitiesForSelectedCountries();
      });
    });
}

function renderCities() {
  const q = els.citySearch.value.trim().toLowerCase();

  const filtered = state.cities.filter((city) => {
    const cityName = (city.city ?? city.name ?? "").toLowerCase();
    const countryName = (city.country ?? city.country_name ?? "").toLowerCase();
    return cityName.includes(q) || countryName.includes(q);
  });

  if (filtered.length === 0) {
    if (state.selectedCountryCodes.size === 0) {
      els.cityList.innerHTML = `<div class="empty-state">Select countries to load cities.</div>`;
    } else {
      els.cityList.innerHTML = `<div class="empty-state">No cities match the current selection.</div>`;
    }
    return;
  }

  els.cityList.innerHTML = filtered
    .map((city) => {
      const id = String(city.geoname_id ?? city.city_id ?? city.id ?? "");
      const cityName = city.city ?? city.name ?? id;
      const countryName =
        city.country ?? city.country_name ?? city.country_code ?? "";
      const checked = state.selectedCityIds.has(id) ? "checked" : "";

      return `
        <label class="selection-item">
          <input type="checkbox" data-city-id="${escapeHtml(id)}" ${checked} />
          <div class="selection-item-main">
            <span class="selection-title">${escapeHtml(cityName)}</span>
            <span class="selection-subtitle">${escapeHtml(countryName)}</span>
          </div>
        </label>
      `;
    })
    .join("");

  els.cityList.querySelectorAll("input[data-city-id]").forEach((input) => {
    input.addEventListener("change", (e) => {
      const id = e.target.dataset.cityId;
      if (!id) return;

      if (e.target.checked) {
        state.selectedCityIds.add(id);
      } else {
        state.selectedCityIds.delete(id);
      }

      renderSelectedCities();
      updateMetrics();
    });
  });
}

function renderSelectedCountries() {
  const items = state.countries.filter((c) =>
    state.selectedCountryCodes.has(c.country_code ?? c.code ?? ""),
  );

  if (items.length === 0) {
    els.selectedCountries.innerHTML = `<span class="chip">None</span>`;
    return;
  }

  els.selectedCountries.innerHTML = items
    .map((country) => {
      const name =
        country.country ?? country.country_name ?? country.name ?? "";
      return `<span class="chip">${escapeHtml(name)}</span>`;
    })
    .join("");
}

function renderSelectedCities() {
  const items = state.cities.filter((c) =>
    state.selectedCityIds.has(String(c.geoname_id ?? c.city_id ?? c.id ?? "")),
  );

  if (items.length === 0) {
    els.selectedCities.innerHTML = `<span class="chip">None</span>`;
    return;
  }

  els.selectedCities.innerHTML = items
    .map((city) => {
      const name = city.city ?? city.name ?? "";
      return `<span class="chip">${escapeHtml(name)}</span>`;
    })
    .join("");
}

async function loadComparison() {
  const cityIds = Array.from(state.selectedCityIds);
  if (cityIds.length === 0) {
    alert("Select at least one city.");
    return;
  }

  const include = "numbeo_cost,numbeo_indices,avg_climate";
  const qs = new URLSearchParams();
  qs.set("ids", cityIds.join(","));
  qs.set("include", include);

  try {
    const res = await fetch(`/cities?${qs.toString()}`);
    if (!res.ok) throw new Error(`Failed to load comparison: ${res.status}`);

    const data = await res.json();
    state.comparisonData = Array.isArray(data.cities)
      ? data.cities
      : Array.isArray(data.data)
        ? data.data
        : [];

    renderComparisonTable();
    renderClimateChart();
    updateMetrics();
    collapseFilters();
  } catch (error) {
    console.error(error);
    alert("Failed to load comparison data.");
  }
}

function renderComparisonTable() {
  if (!state.comparisonData.length) {
    els.tableEmptyState.classList.remove("hidden");
    els.comparisonTableWrapper.classList.add("hidden");
    return;
  }

  els.tableEmptyState.classList.add("hidden");
  els.comparisonTableWrapper.classList.remove("hidden");

  els.comparisonTableBody.innerHTML = state.comparisonData
    .map((city) => {
      const cityName = city.city ?? city.name ?? "—";
      const country =
        city.country ?? city.country_name ?? city.country_code ?? "—";
      const idx = city.numbeo_indices ?? {};
      return `
        <tr>
          <td>${escapeHtml(String(cityName))}</td>
          <td>${escapeHtml(String(country))}</td>
          <td>${fmt(idx.cost_of_living)}</td>
          <td>${fmt(idx.rent_index ?? idx.rent)}</td>
          <td>${fmt(idx.groceries_index ?? idx.groceries)}</td>
          <td>${fmt(idx.restaurant_price_index ?? idx.restaurant)}</td>
          <td>${fmt(idx.quality_of_life)}</td>
          <td>${fmt(idx.safety)}</td>
          <td>${fmt(idx.health_care)}</td>
        </tr>
      `;
    })
    .join("");
}

function renderClimateChart() {
  const datasets = [];
  const labels = [
    "Jan",
    "Feb",
    "Mar",
    "Apr",
    "May",
    "Jun",
    "Jul",
    "Aug",
    "Sep",
    "Oct",
    "Nov",
    "Dec",
  ];

  for (const city of state.comparisonData) {
    const climate = city.avg_climate ?? {};
    const highTemp =
      climate.high_temp ?? climate.high_temps ?? climate.avg_high_temp ?? [];

    if (Array.isArray(highTemp) && highTemp.length > 0) {
      datasets.push({
        label: city.city ?? city.name ?? "City",
        data: normalizeMonthlySeries(highTemp),
        tension: 0.3,
      });
    }
  }

  if (datasets.length === 0) {
    els.chartEmptyState.classList.remove("hidden");
    if (state.climateChart) {
      state.climateChart.destroy();
      state.climateChart = null;
    }
    return;
  }

  els.chartEmptyState.classList.add("hidden");

  if (state.climateChart) {
    state.climateChart.destroy();
  }

  state.climateChart = new Chart(els.climateCanvas, {
    type: "line",
    data: {
      labels,
      datasets,
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: "nearest",
        intersect: false,
      },
      plugins: {
        legend: {
          position: "top",
        },
      },
      scales: {
        y: {
          title: {
            display: true,
            text: "Temperature",
          },
        },
      },
    },
  });
}

function updateMetrics() {
  els.metricCities.textContent = String(state.selectedCityIds.size);
  els.metricCountries.textContent = String(state.selectedCountryCodes.size);

  if (!state.comparisonData.length) {
    els.metricLowestRent.textContent = "—";
    els.metricBestQol.textContent = "—";
    return;
  }

  let lowestRentCity = null;
  let bestQolCity = null;

  for (const city of state.comparisonData) {
    const idx = city.numbeo_indices ?? {};
    const rent = toNumber(idx.rent_index ?? idx.rent);
    const qol = toNumber(idx.quality_of_life);

    if (rent != null) {
      if (!lowestRentCity || rent < lowestRentCity.value) {
        lowestRentCity = {
          name: city.city ?? city.name ?? "—",
          value: rent,
        };
      }
    }

    if (qol != null) {
      if (!bestQolCity || qol > bestQolCity.value) {
        bestQolCity = {
          name: city.city ?? city.name ?? "—",
          value: qol,
        };
      }
    }
  }

  els.metricLowestRent.textContent = lowestRentCity
    ? `${lowestRentCity.name} (${fmt(lowestRentCity.value)})`
    : "—";

  els.metricBestQol.textContent = bestQolCity
    ? `${bestQolCity.name} (${fmt(bestQolCity.value)})`
    : "—";
}

function resetDashboard() {
  state.selectedCountryCodes.clear();
  state.selectedCityIds.clear();
  state.cities = [];
  state.comparisonData = [];

  els.countrySearch.value = "";
  els.citySearch.value = "";

  renderCountries();
  renderCities();
  renderSelectedCountries();
  renderSelectedCities();
  renderComparisonTable();

  els.chartEmptyState.classList.remove("hidden");
  if (state.climateChart) {
    state.climateChart.destroy();
    state.climateChart = null;
  }

  updateMetrics();
  expandFilters();
}

function collapseFilters() {
  els.filtersCard.classList.add("is-collapsed");
  els.editFiltersBtn.classList.remove("hidden");
}

function expandFilters() {
  els.filtersCard.classList.remove("is-collapsed");
  els.editFiltersBtn.classList.add("hidden");
}

function normalizeMonthlySeries(arr) {
  const out = new Array(12).fill(null);
  for (let i = 0; i < Math.min(arr.length, 12); i += 1) {
    out[i] = toNumber(arr[i]);
  }
  return out;
}

function toNumber(value) {
  if (value == null) return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

function fmt(value) {
  const n = toNumber(value);
  return n == null ? "—" : n.toFixed(1);
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
