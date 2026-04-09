const state = {
  allCities: [],
  countries: [],
  cities: [],
  selectedCountryCodes: new Set(),
  selectedCityIds: new Set(),
  comparisonData: [],
  costBreakdownData: [],
  collapsedCostCategories: new Set(),
  breakdownSort: null,
  climateChart: null,
};

const countryDisplayNames = {
  "Bolivia, Plurinational State of": "Bolivia",
  "Bosnia and Herzegovina": "Bosnia & Herzegovina",
  "Dominican Republic": "Dominican Rep.",
  "Iran, Islamic Republic of": "Iran",
  "Korea, Republic of": "South Korea",
  "Moldova, Republic of": "Moldova",
  "Netherlands, Kingdom of the": "Netherlands",
  "North Macedonia": "N. Macedonia",
  "Russian Federation": "Russia",
  "Syrian Arab Republic": "Syria",
  "Taiwan, Province of China": "Taiwan",
  "Tanzania, United Republic of": "Tanzania",
  "United Arab Emirates": "UAE",
  "United Kingdom of Great Britain and Northern Ireland": "United Kingdom",
  "United States of America": "USA",
  "Venezuela, Bolivarian Republic of": "Venezuela",
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
  selectedCountriesCount: document.getElementById("selectedCountriesCount"),
  selectedCitiesCount: document.getElementById("selectedCitiesCount"),
  selectedCitiesLimitHint: document.getElementById("selectedCitiesLimitHint"),
  compareBtn: document.getElementById("compareBtn"),
  resetBtn: document.getElementById("resetBtn"),
  chartEmptyState: document.getElementById("chartEmptyState"),
  costBreakdownEmptyState: document.getElementById("costBreakdownEmptyState"),
  costBreakdownWrapper: document.getElementById("costBreakdownWrapper"),
  costBreakdownHeadRow: document.getElementById("costBreakdownHeadRow"),
  costBreakdownBody: document.getElementById("costBreakdownBody"),
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
    els.countryList.innerHTML = `<div class="selection-list-empty">Failed to load countries.</div>`;
    els.cityList.innerHTML = `<div class="selection-list-empty">Failed to load cities.</div>`;
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
    els.countryList.innerHTML = `<div class="selection-list-empty">No countries found.</div>`;
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
  const cityLimitReached = state.selectedCityIds.size >= 20;

  const filtered = state.cities.filter((city) => {
    const cityName = (city.city ?? city.name ?? "").toLowerCase();
    const countryName = (city.country ?? city.country_name ?? "").toLowerCase();
    return cityName.includes(q) || countryName.includes(q);
  });

  if (filtered.length === 0) {
    if (state.selectedCountryCodes.size === 0) {
      els.cityList.innerHTML = `<div class="selection-list-empty">Select countries to load cities.</div>`;
    } else {
      els.cityList.innerHTML = `<div class="selection-list-empty">No cities match the current selection.</div>`;
    }
    return;
  }

  els.cityList.innerHTML = filtered
    .map((city) => {
      const id = String(city.geoname_id ?? city.city_id ?? city.id ?? "");
      const cityName = city.city ?? city.name ?? id;
      const countryName =
        city.country ?? city.country_name ?? city.country_code ?? "";
      const isChecked = state.selectedCityIds.has(id);
      const checked = isChecked ? "checked" : "";
      const disabled = cityLimitReached && !isChecked ? "disabled" : "";

      return `
        <label class="selection-item">
          <input type="checkbox" data-city-id="${escapeHtml(id)}" ${checked} ${disabled} />
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

      renderCities();
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
      const code = country.country_code ?? country.code ?? "";
      const name =
        country.country ?? country.country_name ?? country.name ?? "";
      const displayName = getDisplayCountryName(name);
      return `
        <span class="chip chip-dismissible">
          <span>${escapeHtml(displayName)}</span>
          <button
            type="button"
            class="chip-remove"
            data-remove-country="${escapeHtml(code)}"
            aria-label="Remove ${escapeHtml(displayName)}"
          >
            ×
          </button>
        </span>
      `;
    })
    .join("");

  els.selectedCountries
    .querySelectorAll("[data-remove-country]")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const code = button.dataset.removeCountry;
        if (!code) return;

        state.selectedCountryCodes.delete(code);
        renderCountries();
        renderSelectedCountries();
        filterCitiesForSelectedCountries();
        applyLoadedComparisonSelection();
      });
    });
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
      const id = String(city.geoname_id ?? city.city_id ?? city.id ?? "");
      const name = city.city ?? city.name ?? "";
      return `
        <span class="chip chip-dismissible">
          <span>${escapeHtml(name)}</span>
          <button
            type="button"
            class="chip-remove"
            data-remove-city="${escapeHtml(id)}"
            aria-label="Remove ${escapeHtml(name)}"
          >
            ×
          </button>
        </span>
      `;
    })
    .join("");

  els.selectedCities
    .querySelectorAll("[data-remove-city]")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const id = button.dataset.removeCity;
        if (!id) return;

        state.selectedCityIds.delete(id);
        renderCities();
        renderSelectedCities();
        updateMetrics();
        applyLoadedComparisonSelection();
      });
    });
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
    state.breakdownSort = null;
    state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);

    renderCostBreakdownTable();
    renderClimateChart();
    updateMetrics();
    collapseFilters();
  } catch (error) {
    console.error(error);
    alert("Failed to load comparison data.");
  }
}

function renderClimateChart() {
  const chartCities = [...state.comparisonData].sort((a, b) => {
    const aName = (a.city ?? a.name ?? "").toLowerCase();
    const bName = (b.city ?? b.name ?? "").toLowerCase();
    return aName.localeCompare(bName);
  });
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

  for (const city of chartCities) {
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
        const country =
          city.country ?? city.country_name ?? city.country_code ?? "";
        const displayCountry = getDisplayCountryName(country);
        return `<th><span class="table-city-name" title="${escapeHtml(String(cityName))}">${escapeHtml(cityName)}</span><br><span class="table-city-meta" title="${escapeHtml(String(country))}">${escapeHtml(String(displayCountry))}</span></th>`;
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
            item.param ===
            "Annual Mortgage Interest Rate (20-Year Fixed, in %)";
          const isActiveSort = state.breakdownSort?.param === item.param;
          const sortDirection = isActiveSort
            ? state.breakdownSort.direction
            : null;
          const numericEntries = item.values
            .map((value, index) => ({
              index,
              numericValue: toNumber(value?.cost),
            }))
            .filter((entry) => entry.numericValue != null);
          const numericValues = numericEntries.map(
            (entry) => entry.numericValue,
          );
          const minValue =
            numericValues.length > 0 ? Math.min(...numericValues) : null;
          const maxValue =
            numericValues.length > 0 ? Math.max(...numericValues) : null;
          const shouldApplySecondary =
            item.values.length - numericValues.length <= 1 &&
            numericValues.length >= 3;
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

              if (
                numericValue != null &&
                minValue != null &&
                maxValue != null &&
                minValue !== maxValue
              ) {
                if (isSalaryRow) {
                  if (numericValue === minValue) {
                    cellClass += " cost-value-high";
                  } else if (numericValue === maxValue) {
                    cellClass += " cost-value-low";
                  } else if (
                    shouldApplySecondary &&
                    secondMinValue != null &&
                    numericValue === secondMinValue
                  ) {
                    cellClass += " cost-value-high-soft";
                  } else if (
                    shouldApplySecondary &&
                    secondMaxValue != null &&
                    numericValue === secondMaxValue
                  ) {
                    cellClass += " cost-value-low-soft";
                  }
                } else if (numericValue === minValue) {
                  cellClass += " cost-value-low";
                } else if (numericValue === maxValue) {
                  cellClass += " cost-value-high";
                } else if (
                  shouldApplySecondary &&
                  secondMinValue != null &&
                  numericValue === secondMinValue
                ) {
                  cellClass += " cost-value-low-soft";
                } else if (
                  shouldApplySecondary &&
                  secondMaxValue != null &&
                  numericValue === secondMaxValue
                ) {
                  cellClass += " cost-value-high-soft";
                }
              }

              return `<td class="${cellClass}">${escapeHtml(formatCostValue(value, { isMortgageRate: isMortgageRateRow }))}</td>`;
            })
            .join("");

          return `
            <tr ${isCollapsed ? 'class="hidden"' : ""}>
              <td class="cost-item-name">
                <button
                  class="cost-item-sort ${isActiveSort ? "is-active" : ""}"
                  type="button"
                  data-cost-param="${escapeHtml(item.param)}"
                  data-cost-sort-direction="${sortDirection ?? ""}"
                >
                  <span>${escapeHtml(item.param)}</span>
                  <span class="cost-item-sort-indicator" aria-hidden="true">${sortDirection === "asc" ? "↑" : sortDirection === "desc" ? "↓" : ""}</span>
                </button>
              </td>
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

  els.costBreakdownBody
    .querySelectorAll("[data-cost-param]")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const param = button.dataset.costParam;
        if (!param) return;
        applyBreakdownSort(param);
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
  const cityCount = state.selectedCityIds.size;
  const countryCount = state.selectedCountryCodes.size;
  const cityLimitReached = cityCount >= 20;

  els.selectedCitiesCount.textContent = String(cityCount);
  els.selectedCountriesCount.textContent = String(countryCount);
  els.selectedCitiesCount.classList.toggle("is-warning", cityLimitReached);
  els.selectedCitiesLimitHint.classList.toggle("hidden", !cityLimitReached);
}

function resetDashboard() {
  state.selectedCountryCodes.clear();
  state.selectedCityIds.clear();
  state.cities = [];
  state.comparisonData = [];
  state.costBreakdownData = [];
  state.collapsedCostCategories.clear();
  state.breakdownSort = null;

  els.countrySearch.value = "";
  els.citySearch.value = "";

  renderCountries();
  renderCities();
  renderSelectedCountries();
  renderSelectedCities();
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
  els.compareBtn.classList.add("hidden");
  els.resetBtn.classList.add("hidden");
  els.editFiltersBtn.classList.remove("hidden");
}

function expandFilters() {
  els.filtersCard.classList.remove("is-collapsed");
  els.compareBtn.classList.remove("hidden");
  els.resetBtn.classList.remove("hidden");
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

function applyBreakdownSort(param) {
  if (!state.comparisonData.length) return;

  const nextDirection =
    state.breakdownSort?.param === param &&
    state.breakdownSort.direction === "desc"
      ? "asc"
      : "desc";

  sortComparisonDataByCostParam(param, nextDirection);
  state.breakdownSort = { param, direction: nextDirection };
  state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);

  renderCostBreakdownTable();
}

function applyLoadedComparisonSelection() {
  if (!state.comparisonData.length) return;

  const selectedIDs = new Set(Array.from(state.selectedCityIds));
  state.comparisonData = state.comparisonData.filter((city) =>
    selectedIDs.has(String(city.geoname_id ?? city.city_id ?? city.id ?? "")),
  );

  if (state.comparisonData.length === 0) {
    state.costBreakdownData = [];
    state.breakdownSort = null;
    renderCostBreakdownTable();
    renderClimateChart();
    return;
  }

  state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);
  renderCostBreakdownTable();
  renderClimateChart();
}

function sortComparisonDataByCostParam(param, direction) {
  const multiplier = direction === "asc" ? 1 : -1;

  state.comparisonData = state.comparisonData
    .map((city, index) => ({
      city,
      index,
      value: getCostParamValue(city, param),
    }))
    .sort((a, b) => {
      const aMissing = a.value == null;
      const bMissing = b.value == null;

      if (aMissing && bMissing) return a.index - b.index;
      if (aMissing) return 1;
      if (bMissing) return -1;
      if (a.value === b.value) return a.index - b.index;

      return (a.value - b.value) * multiplier;
    })
    .map((entry) => entry.city);
}

function getCostParamValue(city, param) {
  const prices = city.numbeo_cost?.prices;
  if (!Array.isArray(prices)) return null;

  const entry = prices.find((price) => price?.param === param);
  return toNumber(entry?.cost);
}

function getDisplayCountryName(country) {
  const raw = String(country ?? "");
  return countryDisplayNames[raw] ?? raw;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
