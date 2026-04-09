const state = {
  allCities: [],
  countries: [],
  cities: [],
  selectedCountryCodes: new Set(),
  selectedCityIds: new Set(),
  comparisonData: [],
  costBreakdownData: [],
  collapsedCostCategories: new Set(),
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
  costBreakdownEmptyState: document.getElementById("costBreakdownEmptyState"),
  costBreakdownWrapper: document.getElementById("costBreakdownWrapper"),
  costBreakdownHeadRow: document.getElementById("costBreakdownHeadRow"),
  costBreakdownBody: document.getElementById("costBreakdownBody"),
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
    state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);

    renderComparisonTable();
    renderCostBreakdownTable();
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

function renderCostBreakdownTable() {
  if (!state.costBreakdownData.length || !state.comparisonData.length) {
    els.costBreakdownEmptyState.classList.remove("hidden");
    els.costBreakdownWrapper.classList.add("hidden");
    return;
  }

  els.costBreakdownEmptyState.classList.add("hidden");
  els.costBreakdownWrapper.classList.remove("hidden");

  els.costBreakdownHeadRow.innerHTML = `
    <th>Item</th>
    ${state.comparisonData
      .map((city) => {
        const cityName = city.city ?? city.name ?? "City";
        const country = city.country_code ?? city.country ?? "";
        return `<th>${escapeHtml(cityName)}<br><span class="selection-subtitle">${escapeHtml(String(country))}</span></th>`;
      })
      .join("")}
  `;

  els.costBreakdownBody.innerHTML = state.costBreakdownData
    .map((group) => {
      const isCollapsed = state.collapsedCostCategories.has(group.category);
      const groupRows = group.items
        .map((item) => {
          const isSalaryRow =
            item.param === "Average Monthly Net Salary (After Tax)";
          const isMortgageRateRow =
            item.param === "Annual Mortgage Interest Rate (20-Year Fixed, in %)";
          const numericEntries = item.values
            .map((value, index) => ({
              index,
              numericValue: toNumber(value?.cost),
            }))
            .filter((entry) => entry.numericValue != null);
          const numericValues = numericEntries.map((entry) => entry.numericValue);
          const minValue =
            numericValues.length > 0 ? Math.min(...numericValues) : null;
          const maxValue =
            numericValues.length > 0 ? Math.max(...numericValues) : null;
          const shouldApplySecondary =
            item.values.length - numericValues.length <= 1 && numericValues.length >= 3;
          const sortedUniqueValues = shouldApplySecondary
            ? Array.from(new Set(numericValues)).sort((a, b) => a - b)
            : [];
          const secondMinValue =
            sortedUniqueValues.length >= 3 ? sortedUniqueValues[1] : null;
          const secondMaxValue =
            sortedUniqueValues.length >= 3
              ? sortedUniqueValues[sortedUniqueValues.length - 2]
              : null;
          const cells = item.values
            .map((value) => {
              const numericValue = toNumber(value?.cost);
              let cellClass = "cost-value";

              if (numericValue != null && minValue != null && maxValue != null && minValue !== maxValue) {
                if (isSalaryRow) {
                  if (numericValue === minValue) {
                    cellClass += " cost-value-high";
                  } else if (numericValue === maxValue) {
                    cellClass += " cost-value-low";
                  } else if (shouldApplySecondary && secondMinValue != null && numericValue === secondMinValue) {
                    cellClass += " cost-value-high-soft";
                  } else if (shouldApplySecondary && secondMaxValue != null && numericValue === secondMaxValue) {
                    cellClass += " cost-value-low-soft";
                  }
                } else if (numericValue === minValue) {
                  cellClass += " cost-value-low";
                } else if (numericValue === maxValue) {
                  cellClass += " cost-value-high";
                } else if (shouldApplySecondary && secondMinValue != null && numericValue === secondMinValue) {
                  cellClass += " cost-value-low-soft";
                } else if (shouldApplySecondary && secondMaxValue != null && numericValue === secondMaxValue) {
                  cellClass += " cost-value-high-soft";
                }
              }

              return `<td class="${cellClass}">${escapeHtml(formatCostValue(value, { isMortgageRate: isMortgageRateRow }))}</td>`;
            })
            .join("");

          return `
            <tr ${isCollapsed ? 'class="hidden"' : ""}>
              <td class="cost-item-name">${escapeHtml(item.param)}</td>
              ${cells}
            </tr>
          `;
        })
        .join("");

      return `
        <tr class="cost-group-row ${isCollapsed ? "is-collapsed" : ""}">
          <td class="cost-group-cell cost-group-cell-main">
            <button class="cost-group-toggle" type="button" data-cost-group="${escapeHtml(group.category)}" aria-expanded="${isCollapsed ? "false" : "true"}">
              <span class="cost-group-label">
                <span class="cost-group-arrow">▾</span>
                <span>${escapeHtml(group.category)}</span>
              </span>
            </button>
          </td>
          <td class="cost-group-cell cost-group-cell-fill" colspan="${state.comparisonData.length}">
            <span class="cost-group-count">${group.items.length} items</span>
          </td>
        </tr>
        ${groupRows}
      `;
    })
    .join("");

  els.costBreakdownBody
    .querySelectorAll("[data-cost-group]")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const category = button.dataset.costGroup;
        if (!category) return;

        if (state.collapsedCostCategories.has(category)) {
          state.collapsedCostCategories.delete(category);
        } else {
          state.collapsedCostCategories.add(category);
        }

        renderCostBreakdownTable();
      });
    });
}

function buildCostBreakdownDataset(cities) {
  if (!cities.length) return [];

  const preferredCategoryOrder = [
    "Summary",
    "Rent Per Month",
    "Markets",
    "Restaurants",
    "Transportation",
    "Utilities (Monthly)",
    "Sports And Leisure",
    "Childcare",
    "Clothing And Shoes",
    "Buy Apartment Price",
    "Salaries And Financing",
  ];
  const categories = new Map();

  for (const city of cities) {
    const prices = city.numbeo_cost?.prices;
    if (!Array.isArray(prices)) continue;

    for (const price of prices) {
      if (!price?.category || !price?.param) continue;

      if (!categories.has(price.category)) {
        categories.set(price.category, new Map());
      }

      const categoryItems = categories.get(price.category);
      if (!categoryItems.has(price.param)) {
        categoryItems.set(price.param, true);
      }
    }
  }

  const sortedCategories = Array.from(categories.keys()).sort((a, b) => {
    const aIndex = preferredCategoryOrder.indexOf(a);
    const bIndex = preferredCategoryOrder.indexOf(b);
    if (aIndex === -1 && bIndex === -1) return a.localeCompare(b);
    if (aIndex === -1) return 1;
    if (bIndex === -1) return -1;
    return aIndex - bIndex;
  });

  return sortedCategories.map((category) => {
    const params = Array.from(categories.get(category).keys());

    return {
      category,
      items: params.map((param) => ({
        param,
        values: cities.map((city) => {
          const prices = city.numbeo_cost?.prices;
          const currency = city.numbeo_cost?.currency ?? "USD";
          if (!Array.isArray(prices)) {
            return { cost: null, currency };
          }

          const entry = prices.find(
            (price) => price.category === category && price.param === param,
          );

          return {
            cost: entry?.cost ?? null,
            currency,
          };
        }),
      })),
    };
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
  state.costBreakdownData = [];
  state.collapsedCostCategories.clear();

  els.countrySearch.value = "";
  els.citySearch.value = "";

  renderCountries();
  renderCities();
  renderSelectedCountries();
  renderSelectedCities();
  renderComparisonTable();
  renderCostBreakdownTable();

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

function formatCostValue(value, options = {}) {
  if (!value) return "—";
  const n = toNumber(value.cost);
  if (n == null) return "—";

  if (options.isMortgageRate) {
    return `${n.toFixed(2)}%`;
  }

  const fractionDigits = n >= 100 ? 0 : 2;
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: value.currency ?? "USD",
    maximumFractionDigits: fractionDigits,
    minimumFractionDigits: fractionDigits,
  }).format(n);
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
