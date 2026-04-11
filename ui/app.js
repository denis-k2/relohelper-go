const state = {
  allCities: [],
  countries: [],
  cities: [],
  selectedCountryCodes: new Set(),
  selectedCityIds: new Set(),
  comparisonData: [],
  countryComparisonData: [],
  costBreakdownData: [],
  collapsedCostCategories: new Set(),
  breakdownSort: null,
  indicesSort: null,
  countryIndicesSort: null,
  legatumIndicesSort: null,
  legatumMetricMode: "score",
  climateCharts: [],
  climateHiddenCities: new Set(),
  climateHoveredCityKey: null,
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

const numbeoIndexRows = [
  { key: "quality_of_life", label: "Quality of Life", better: "high" },
  { key: "safety", label: "Safety", better: "high" },
  { key: "health_care", label: "Health Care", better: "high" },
  { key: "climate", label: "Climate", better: "high" },
  { key: "local_purchasing_power", label: "Local Purchasing Power", better: "high" },
  { key: "cost_of_living", label: "Cost of Living", better: "low" },
  { key: "rent", label: "Rent", better: "low" },
  { key: "cost_of_living_plus_rent", label: "Cost of Living Plus Rent", better: "low" },
  { key: "groceries", label: "Groceries", better: "low" },
  { key: "property_price_to_income_ratio", label: "Property Price to Income Ratio", better: "low" },
  { key: "traffic_commute_time", label: "Traffic Commute Time", better: "low" },
  { key: "pollution", label: "Pollution", better: "low" },
];

const countryNumbeoIndexRows = [
  ...numbeoIndexRows,
  { key: "avg_salary_usd", label: "Average Salary (USD)", better: "high", format: "currency" },
];

const legatumIndexRows = [
  { key: "safety_and_security", label: "Safety and Security", better: "high" },
  { key: "personal_freedom", label: "Personal Freedom", better: "high" },
  { key: "governance", label: "Governance", better: "high" },
  { key: "social_capital", label: "Social Capital", better: "high" },
  { key: "investment_invironment", label: "Investment Invironment", better: "high" },
  { key: "enterprise_conditions", label: "Enterprise Conditions", better: "high" },
  { key: "infrastructure_and_market_access", label: "Infrastructure and Market Access", better: "high" },
  { key: "economic_quality", label: "Economic Quality", better: "high" },
  { key: "living_conditions", label: "Living Conditions", better: "high" },
  { key: "health", label: "Health", better: "high" },
  { key: "education", label: "Education", better: "high" },
  { key: "natural_environment", label: "Natural Environment", better: "high" },
];

const legatumYears = Array.from({ length: 17 }, (_, index) => 2007 + index);

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
  climateChartsGrid: document.getElementById("climateChartsGrid"),
  costBreakdownEmptyState: document.getElementById("costBreakdownEmptyState"),
  costBreakdownWrapper: document.getElementById("costBreakdownWrapper"),
  costBreakdownHeadRow: document.getElementById("costBreakdownHeadRow"),
  costBreakdownBody: document.getElementById("costBreakdownBody"),
  numbeoIndicesEmptyState: document.getElementById("numbeoIndicesEmptyState"),
  numbeoIndicesWrapper: document.getElementById("numbeoIndicesWrapper"),
  numbeoIndicesHeadRow: document.getElementById("numbeoIndicesHeadRow"),
  numbeoIndicesBody: document.getElementById("numbeoIndicesBody"),
  countryNumbeoIndicesEmptyState: document.getElementById("countryNumbeoIndicesEmptyState"),
  countryNumbeoIndicesWrapper: document.getElementById("countryNumbeoIndicesWrapper"),
  countryNumbeoIndicesHeadRow: document.getElementById("countryNumbeoIndicesHeadRow"),
  countryNumbeoIndicesBody: document.getElementById("countryNumbeoIndicesBody"),
  legatumIndicesEmptyState: document.getElementById("legatumIndicesEmptyState"),
  legatumIndicesWrapper: document.getElementById("legatumIndicesWrapper"),
  legatumIndicesHeadRow: document.getElementById("legatumIndicesHeadRow"),
  legatumIndicesBody: document.getElementById("legatumIndicesBody"),
  legatumScoreToggle: document.getElementById("legatumScoreToggle"),
  legatumRankToggle: document.getElementById("legatumRankToggle"),
  climateTemperatureCanvas: document.getElementById("climateTemperatureChart"),
  climateSunshineCanvas: document.getElementById("climateSunshineChart"),
  climateDaylightCanvas: document.getElementById("climateDaylightChart"),
  climateHumidityCanvas: document.getElementById("climateHumidityChart"),
  climateRainfallCanvas: document.getElementById("climateRainfallChart"),
  climateWindCanvas: document.getElementById("climateWindChart"),
  climateUVIndexCanvas: document.getElementById("climateUVIndexChart"),
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
  els.legatumScoreToggle.addEventListener("click", () =>
    setLegatumMetricMode("score"),
  );
  els.legatumRankToggle.addEventListener("click", () =>
    setLegatumMetricMode("rank"),
  );
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
      const cityDisplayName = formatFilterCityName(city);
      const countryName =
        city.country ?? city.country_name ?? city.country_code ?? "";
      const isChecked = state.selectedCityIds.has(id);
      const checked = isChecked ? "checked" : "";
      const disabled = cityLimitReached && !isChecked ? "disabled" : "";

      return `
        <label class="selection-item">
          <input type="checkbox" data-city-id="${escapeHtml(id)}" ${checked} ${disabled} />
          <div class="selection-item-main">
            <span class="selection-title">${cityDisplayName}</span>
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
    state.countryComparisonData = await loadCountryComparisonData(state.comparisonData);
    state.breakdownSort = null;
    state.climateHiddenCities.clear();
    state.climateHoveredCityKey = null;
    state.indicesSort = null;
    state.countryIndicesSort = null;
    state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);

    renderCostBreakdownTable();
    renderClimateChart();
    renderNumbeoIndicesTable();
    renderCountryNumbeoIndicesTable();
    renderLegatumIndicesTable();
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
  const climateCities = chartCities.map((city, index) =>
    buildClimateCityConfig(city, index),
  );

  if (climateCities.length === 0) {
    els.chartEmptyState.classList.remove("hidden");
    els.climateChartsGrid.classList.add("hidden");
    destroyClimateCharts();
    return;
  }

  els.chartEmptyState.classList.add("hidden");
  els.climateChartsGrid.classList.remove("hidden");

  destroyClimateCharts();

  state.climateCharts = [
    new Chart(
      els.climateTemperatureCanvas,
      buildClimateChartConfig({
        chartKey: "temperature",
        labels,
        datasets: climateCities.flatMap((city) => [
          buildClimateDataset(city, "high_temp", {
            datasetLabel: `${city.name} high`,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
            borderDash: [],
            variant: "high",
          }),
          buildClimateDataset(city, "low_temp", {
            datasetLabel: `${city.name} low`,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
            borderDash: [8, 5],
            variant: "low",
          }),
        ]),
        yTitle: "°C",
        showLegend: true,
      }),
    ),
    new Chart(
      els.climateSunshineCanvas,
      buildClimateChartConfig({
        chartKey: "sunshine",
        labels,
        datasets: climateCities.map((city) =>
          buildClimateDataset(city, "sunshine", {
            datasetLabel: city.name,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
          }),
        ),
        yTitle: "Hours",
      }),
    ),
    new Chart(
      els.climateDaylightCanvas,
      buildClimateChartConfig({
        chartKey: "daylight",
        labels,
        datasets: climateCities.map((city) =>
          buildClimateDataset(city, "daylight", {
            datasetLabel: city.name,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
          }),
        ),
        yTitle: "Hours",
      }),
    ),
    new Chart(
      els.climateHumidityCanvas,
      buildClimateChartConfig({
        chartKey: "humidity",
        labels,
        datasets: climateCities.map((city) =>
          buildClimateDataset(city, "humidity", {
            datasetLabel: city.name,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
          }),
        ),
        yTitle: "%",
      }),
    ),
    new Chart(
      els.climateRainfallCanvas,
      buildClimateChartConfig({
        chartKey: "rainfall",
        labels,
        datasets: climateCities.map((city) =>
          buildClimateDataset(city, "rainfall", {
            datasetLabel: city.name,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
          }),
        ),
        yTitle: "mm",
      }),
    ),
    new Chart(
      els.climateWindCanvas,
      buildClimateChartConfig({
        chartKey: "wind",
        labels,
        datasets: climateCities.map((city) =>
          buildClimateDataset(city, "wind_speed", {
            datasetLabel: city.name,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
          }),
        ),
        yTitle: "km/h",
      }),
    ),
    new Chart(
      els.climateUVIndexCanvas,
      buildClimateChartConfig({
        chartKey: "uv_index",
        labels,
        datasets: climateCities.map((city) =>
          buildClimateDataset(city, "uv_index", {
            datasetLabel: city.name,
            cityLabel: city.name,
            cityKey: city.key,
            color: city.color,
          }),
        ),
        yTitle: "",
        reserveYAxisTitleSpace: true,
      }),
    ),
  ];

  syncClimateChartStyles();
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
        const cityID = String(city.geoname_id ?? city.city_id ?? city.id ?? "");
        return `<th><span class="table-header-trigger" data-header-kind="city" data-city-id="${escapeHtml(cityID)}"><span class="table-city-name">${escapeHtml(cityName)}</span><br><span class="table-city-meta">${escapeHtml(String(displayCountry))}</span></span></th>`;
      })
      .join("")}
  `;
  bindHeaderInfoTooltips(els.costBreakdownHeadRow);

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
            item.values.length - numericValues.length < 5 &&
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

function renderNumbeoIndicesTable() {
  renderNumbeoIndicesTableSection({
    data: state.comparisonData,
    emptyStateEl: els.numbeoIndicesEmptyState,
    wrapperEl: els.numbeoIndicesWrapper,
    headRowEl: els.numbeoIndicesHeadRow,
    bodyEl: els.numbeoIndicesBody,
    headLabel: "Index",
    rows: numbeoIndexRows,
    sortState: state.indicesSort,
    keyDataAttr: "data-index-key",
    applySort: applyIndicesSort,
    getColumnLabel: (city) => city.city ?? city.name ?? "City",
    getColumnMeta: (city) =>
      getDisplayCountryName(
        city.country ?? city.country_name ?? city.country_code ?? "",
      ),
    getColumnMetaTitle: (city) =>
      city.country ?? city.country_name ?? city.country_code ?? "",
    getValue: (city, key) => toNumber(city.numbeo_indices?.[key]),
    getHeaderDataAttrs: (city) =>
      `data-header-kind="city" data-city-id="${escapeHtml(String(city.geoname_id ?? city.city_id ?? city.id ?? ""))}"`,
  });
}

function renderCountryNumbeoIndicesTable() {
  renderNumbeoIndicesTableSection({
    data: state.countryComparisonData,
    emptyStateEl: els.countryNumbeoIndicesEmptyState,
    wrapperEl: els.countryNumbeoIndicesWrapper,
    headRowEl: els.countryNumbeoIndicesHeadRow,
    bodyEl: els.countryNumbeoIndicesBody,
    headLabel: "Index",
    rows: countryNumbeoIndexRows,
    sortState: state.countryIndicesSort,
    keyDataAttr: "data-country-index-key",
    applySort: applyCountryIndicesSort,
    getColumnLabel: (country) =>
      getDisplayCountryName(
        country.country ?? country.country_name ?? country.country_code ?? "Country",
      ),
    getColumnMeta: () => "",
    getColumnMetaTitle: () => "",
    getValue: (country, key) => toNumber(country.numbeo_indices?.[key]),
    getHeaderDataAttrs: (country) =>
      `data-header-kind="country" data-country-code="${escapeHtml(String(country.country_code ?? ""))}"`,
  });
}

function renderLegatumIndicesTable() {
  const isRankMode = state.legatumMetricMode === "rank";
  renderNumbeoIndicesTableSection({
    data: state.countryComparisonData,
    emptyStateEl: els.legatumIndicesEmptyState,
    wrapperEl: els.legatumIndicesWrapper,
    headRowEl: els.legatumIndicesHeadRow,
    bodyEl: els.legatumIndicesBody,
    headLabel: "Index",
    rows: legatumIndexRows.map((row) => ({
      ...row,
      better: isRankMode ? "low" : "high",
    })),
    sortState: state.legatumIndicesSort,
    keyDataAttr: "data-legatum-index-key",
    applySort: applyLegatumIndicesSort,
    getColumnLabel: (country) =>
      getDisplayCountryName(
        country.country ?? country.country_name ?? country.country_code ?? "Country",
      ),
    getColumnMeta: () => "",
    getColumnMetaTitle: () => "",
    getValue: (country, key) =>
      toNumber(
        country.legatum_indices?.[key]?.[
          isRankMode ? "rank_2023" : "score_2023"
        ],
      ),
    getHeaderDataAttrs: (country) =>
      `data-header-kind="country" data-country-code="${escapeHtml(String(country.country_code ?? ""))}"`,
    getCellDataAttrs: (country, row) =>
      `data-legatum-row-key="${escapeHtml(row.key)}" data-legatum-country-code="${escapeHtml(String(country.country_code ?? ""))}"`,
  });
  bindLegatumTooltipEvents();
  syncLegatumToggleButtons();
}

function renderNumbeoIndicesTableSection({
  data,
  emptyStateEl,
  wrapperEl,
  headRowEl,
  bodyEl,
  headLabel,
  rows,
  sortState,
  keyDataAttr,
  applySort,
  getColumnLabel,
  getColumnMeta,
  getColumnMetaTitle,
  getValue,
  getHeaderDataAttrs = null,
  getCellDataAttrs = null,
}) {
  if (!data.length) {
    emptyStateEl.classList.remove("hidden");
    wrapperEl.classList.add("hidden");
    return;
  }

  emptyStateEl.classList.add("hidden");
  wrapperEl.classList.remove("hidden");

  headRowEl.innerHTML = `
    <th>${escapeHtml(headLabel)}</th>
    ${data
      .map((item) => {
        const label = getColumnLabel(item);
        const meta = getColumnMeta(item);
        const metaHtml = meta
          ? `<br><span class="table-city-meta">${escapeHtml(String(meta))}</span>`
          : "";
        const headerDataAttrs = getHeaderDataAttrs
          ? getHeaderDataAttrs(item)
          : "";
        return `<th><span class="table-header-trigger" ${headerDataAttrs}><span class="table-city-name">${escapeHtml(String(label))}</span>${metaHtml}</span></th>`;
      })
      .join("")}
  `;
  bindHeaderInfoTooltips(headRowEl);

  bodyEl.innerHTML = rows
    .map((row) => {
      const isHigherBetter = row.better === "high";
      const isActiveSort = sortState?.key === row.key;
      const sortDirection = isActiveSort ? sortState.direction : null;
      const numericEntries = data
        .map((item, index) => ({
          index,
          numericValue: getValue(item, row.key),
        }))
        .filter((entry) => entry.numericValue != null);
      const numericValues = numericEntries.map((entry) => entry.numericValue);
      const minValue =
        numericValues.length > 0 ? Math.min(...numericValues) : null;
      const maxValue =
        numericValues.length > 0 ? Math.max(...numericValues) : null;
      const shouldApplySecondary =
        data.length - numericValues.length < 5 &&
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

      const cells = data
        .map((item) => {
          const value = getValue(item, row.key);
          const cellDataAttrs = getCellDataAttrs
            ? getCellDataAttrs(item, row)
            : "";
          let cellClass = "indices-value";

          if (
            value != null &&
            minValue != null &&
            maxValue != null &&
            minValue !== maxValue
          ) {
            if (isHigherBetter) {
              if (value === minValue) {
                cellClass += " cost-value-high";
              } else if (value === maxValue) {
                cellClass += " cost-value-low";
              } else if (
                shouldApplySecondary &&
                secondMinValue != null &&
                value === secondMinValue
              ) {
                cellClass += " cost-value-high-soft";
              } else if (
                shouldApplySecondary &&
                secondMaxValue != null &&
                value === secondMaxValue
              ) {
                cellClass += " cost-value-low-soft";
              }
            } else if (value === minValue) {
              cellClass += " cost-value-low";
            } else if (value === maxValue) {
              cellClass += " cost-value-high";
            } else if (
              shouldApplySecondary &&
              secondMinValue != null &&
              value === secondMinValue
            ) {
              cellClass += " cost-value-low-soft";
            } else if (
              shouldApplySecondary &&
              secondMaxValue != null &&
              value === secondMaxValue
            ) {
              cellClass += " cost-value-high-soft";
            }
          }

          return `<td class="${cellClass}" ${cellDataAttrs}>${escapeHtml(formatIndexValue(value, row))}</td>`;
        })
        .join("");

      return `
        <tr>
          <td class="indices-item-name">
            <button
              class="cost-item-sort ${isActiveSort ? "is-active" : ""}"
              type="button"
              ${keyDataAttr}="${escapeHtml(row.key)}"
              data-index-sort-direction="${sortDirection ?? ""}"
            >
              <span>${escapeHtml(row.label)}</span>
              <span class="cost-item-sort-indicator" aria-hidden="true">${sortDirection === "asc" ? "↑" : sortDirection === "desc" ? "↓" : ""}</span>
            </button>
          </td>
          ${cells}
        </tr>
      `;
    })
    .join("");

  bodyEl.querySelectorAll(`[${keyDataAttr}]`).forEach((button) => {
    button.addEventListener("click", () => {
      const key = button.getAttribute(keyDataAttr);
      if (!key) return;
      applySort(key);
    });
  });
}

function bindHeaderInfoTooltips(container) {
  container
    .querySelectorAll(".table-header-trigger[data-header-kind]")
    .forEach((cell) => {
      cell.addEventListener("mouseenter", () => {
        showHeaderInfoTooltip(cell);
      });
      cell.addEventListener("mousemove", (event) => {
        positionHeaderInfoTooltip(event.clientX, event.clientY);
      });
      cell.addEventListener("mouseleave", hideHeaderInfoTooltip);
    });
}

function showHeaderInfoTooltip(cell) {
  const kind = cell.dataset.headerKind;
  if (!kind) return;

  const tooltipEl = getHeaderInfoTooltipElement();
  const content = kind === "city"
    ? buildCityHeaderTooltipContent(cell.dataset.cityId)
    : buildCountryHeaderTooltipContent(cell.dataset.countryCode);

  if (!content) {
    tooltipEl.classList.remove("is-visible");
    return;
  }

  tooltipEl.innerHTML = content;
  tooltipEl.classList.add("is-visible");
}

function hideHeaderInfoTooltip() {
  getHeaderInfoTooltipElement().classList.remove("is-visible");
}

function positionHeaderInfoTooltip(clientX, clientY) {
  const tooltipEl = getHeaderInfoTooltipElement();
  if (!tooltipEl.classList.contains("is-visible")) return;

  const offset = 14;
  const rect = tooltipEl.getBoundingClientRect();
  let left = clientX + offset;
  let top = clientY + offset;

  if (left + rect.width > window.innerWidth - 12) {
    left = clientX - rect.width - offset;
  }
  if (top + rect.height > window.innerHeight - 12) {
    top = clientY - rect.height - offset;
  }

  tooltipEl.style.left = `${Math.max(12, left)}px`;
  tooltipEl.style.top = `${Math.max(12, top)}px`;
}

function getHeaderInfoTooltipElement() {
  let tooltipEl = document.querySelector(".header-info-tooltip");
  if (!tooltipEl) {
    tooltipEl = document.createElement("div");
    tooltipEl.className = "header-info-tooltip";
    document.body.appendChild(tooltipEl);
  }
  return tooltipEl;
}

function buildCityHeaderTooltipContent(cityID) {
  if (!cityID) return "";

  const city = state.comparisonData.find(
    (item) => String(item.geoname_id ?? item.city_id ?? item.id ?? "") === String(cityID),
  );
  if (!city) return "";

  const cityName = city.city ?? city.name ?? "City";
  const population = formatCompactPopulation(city.population);
  const utcOffset = formatUTCOffset(city.timezone);

  return `
    <div class="header-info-title">${escapeHtml(cityName)}</div>
    <div class="header-info-row"><span class="header-info-key">Pop.</span><span class="header-info-value">${escapeHtml(population)}</span></div>
    <div class="header-info-row"><span class="header-info-key">UTC</span><span class="header-info-value">${escapeHtml(utcOffset)}</span></div>
  `;
}

function buildCountryHeaderTooltipContent(countryCode) {
  if (!countryCode) return "";

  const country = state.countryComparisonData.find(
    (item) => String(item.country_code ?? "") === String(countryCode),
  );
  if (!country) return "";

  const countryName =
    country.country ?? country.country_name ?? country.country_code ?? "Country";
  const population = formatCompactPopulation(country.population);
  const area = formatCompactArea(country.area);

  return `
    <div class="header-info-title">${escapeHtml(countryName)}</div>
    <div class="header-info-row"><span class="header-info-key">Pop.</span><span class="header-info-value">${escapeHtml(population)}</span></div>
    <div class="header-info-row"><span class="header-info-key">Area</span><span class="header-info-value">${escapeHtml(area)}</span></div>
  `;
}

function formatCompactPopulation(value) {
  const n = toNumber(value);
  if (n == null) return "—";
  return new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(n);
}

function formatCompactArea(value) {
  const n = toNumber(value);
  if (n == null) return "—";
  return `${new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(n)} km²`;
}

function formatUTCOffset(timezone) {
  if (!timezone) return "—";

  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      timeZone: timezone,
      timeZoneName: "shortOffset",
    }).formatToParts(new Date());
    const timeZoneName =
      parts.find((part) => part.type === "timeZoneName")?.value ?? "";

    if (!timeZoneName) return String(timezone);
    return timeZoneName.replace("GMT", "UTC");
  } catch {
    return String(timezone);
  }
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
  state.countryComparisonData = [];
  state.costBreakdownData = [];
  state.collapsedCostCategories.clear();
  state.breakdownSort = null;
  state.indicesSort = null;
  state.countryIndicesSort = null;
  state.legatumIndicesSort = null;
  state.legatumMetricMode = "score";
  state.climateHiddenCities.clear();
  state.climateHoveredCityKey = null;

  els.countrySearch.value = "";
  els.citySearch.value = "";

  renderCountries();
  renderCities();
  renderSelectedCountries();
  renderSelectedCities();
  renderCostBreakdownTable();
  renderNumbeoIndicesTable();
  renderCountryNumbeoIndicesTable();
  renderLegatumIndicesTable();

  els.chartEmptyState.classList.remove("hidden");
  els.climateChartsGrid.classList.add("hidden");
  destroyClimateCharts();

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

function buildClimateCityConfig(city, index) {
  const climate = city.avg_climate ?? buildPlaceholderClimate(city, index);
  const key = String(city.geoname_id ?? city.city_id ?? city.id ?? city.city ?? index);
  return {
    key,
    name: city.city ?? city.name ?? `City ${index + 1}`,
    climate,
    color: getClimateColor(index),
  };
}

function buildPlaceholderClimate(city, index) {
  const offset = index * 1.4;
  return {
    high_temp: [5, 7, 11, 16, 21, 25, 28, 28, 23, 17, 11, 7].map((v) => v + offset),
    low_temp: [-1, 0, 3, 8, 13, 17, 20, 20, 16, 10, 5, 1].map((v) => v + offset),
    sunshine: [3, 4, 5, 7, 8, 9, 10, 9, 7, 5, 4, 3].map((v) => Math.max(0, v - index * 0.2)),
    daylight: [9, 10, 12, 13, 15, 16, 15, 14, 12, 11, 9, 8].map((v) => v),
    humidity: [76, 74, 71, 69, 68, 66, 64, 65, 68, 72, 75, 77].map((v) => v + index),
    rainfall: [62, 58, 61, 67, 72, 76, 81, 79, 70, 66, 64, 68].map((v) => v + index * 4),
    wind_speed: [14, 13, 13, 12, 11, 10, 9, 9, 10, 11, 12, 13].map((v) => Math.max(3, v - index * 0.3)),
    uv_index: [1, 2, 3, 5, 6, 7, 8, 7, 5, 3, 2, 1].map((v) => Math.min(11, Math.max(0, v + index * 0.2))),
  };
}

function getClimateColor(index) {
  const palette = [
    "#2563eb",
    "#ef4444",
    "#0f766e",
    "#d97706",
    "#7c3aed",
    "#db2777",
    "#0891b2",
    "#65a30d",
    "#c2410c",
    "#475569",
    "#16a34a",
    "#ea580c",
    "#4f46e5",
    "#be123c",
    "#0d9488",
  ];
  return palette[index % palette.length];
}

function buildClimateDataset(city, metricKey, options = {}) {
  const rawSeries = city.climate?.[metricKey] ?? [];
  const baseColor = options.color ?? city.color;
  return {
    label: options.datasetLabel ?? city.name,
    data: normalizeMonthlySeries(rawSeries),
    borderColor: baseColor,
    backgroundColor: withAlpha(baseColor, 0.12),
    borderWidth: 2.5,
    tension: 0.35,
    pointRadius: 0,
    pointHoverRadius: 4,
    pointHitRadius: 14,
    fill: false,
    borderDash: options.borderDash ?? [],
    cityKey: options.cityKey ?? city.key,
    cityLabel: options.cityLabel ?? city.name,
    variant: options.variant ?? "default",
    _baseColor: baseColor,
  };
}

function buildClimateChartConfig({
  chartKey,
  labels,
  datasets,
  yTitle,
  reserveYAxisTitleSpace = false,
  showLegend = false,
}) {
  return {
    type: "line",
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: "nearest",
        intersect: false,
      },
      onHover: (_event, elements, chart) => {
        const hoveredDataset = elements.length > 0
          ? chart.data.datasets[elements[0].datasetIndex]
          : null;
        setHoveredClimateCity(hoveredDataset?.cityKey ?? null);
      },
      plugins: {
        legend: showLegend
          ? {
              position: "top",
              labels: {
                usePointStyle: true,
                pointStyle: "circle",
                boxWidth: 10,
                boxHeight: 10,
                padding: 14,
                font: {
                  size: 13,
                  lineHeight: 1.15,
                },
                generateLabels: (chart) => {
                  const seen = new Set();
                  const labels = [];
                  chart.data.datasets.forEach((dataset, datasetIndex) => {
                    if (seen.has(dataset.cityKey)) return;
                    seen.add(dataset.cityKey);
                    const isHidden = state.climateHiddenCities.has(dataset.cityKey);
                    labels.push({
                      text: dataset.cityLabel,
                      fillStyle: withAlpha(dataset._baseColor, isHidden ? 0.18 : 0.9),
                      strokeStyle: withAlpha(dataset._baseColor, isHidden ? 0.18 : 0.9),
                      lineWidth: 2,
                      hidden: isHidden,
                      datasetIndex,
                      cityKey: dataset.cityKey,
                    });
                  });
                  return labels;
                },
              },
              onClick: (_event, legendItem) => {
                const cityKey = legendItem.cityKey;
                if (!cityKey) return;
                if (state.climateHiddenCities.has(cityKey)) {
                  state.climateHiddenCities.delete(cityKey);
                } else {
                  state.climateHiddenCities.add(cityKey);
                }
                syncClimateChartStyles();
              },
            }
          : {
              display: false,
            },
        tooltip: {
          enabled: false,
          mode: "nearest",
          intersect: false,
          external: (context) =>
            renderClimateTooltip(context, { chartKey, yTitle }),
        },
      },
      scales: {
        x: {
          grid: {
            color: "rgba(148, 163, 184, 0.12)",
          },
          ticks: {
            color: "#64748b",
          },
        },
        y: {
          grid: {
            color: "rgba(148, 163, 184, 0.12)",
          },
          ticks: {
            color: "#64748b",
          },
          title: {
            display: Boolean(yTitle) || reserveYAxisTitleSpace,
            text: yTitle || " ",
            color: yTitle ? "#64748b" : "rgba(0,0,0,0)",
            font: {
              size: 12,
              weight: "600",
            },
          },
        },
      },
      elements: {
        line: {
          capBezierPoints: true,
        },
      },
    },
    plugins: [
      {
        id: `climate-hover-reset-${chartKey}`,
        afterEvent: (_chart, args) => {
          if (args.event.type === "mouseout") {
            setHoveredClimateCity(null);
          }
        },
      },
    ],
  };
}

function destroyClimateCharts() {
  state.climateCharts.forEach((chart) => chart.destroy());
  state.climateCharts = [];
}

function setHoveredClimateCity(cityKey) {
  if (state.climateHoveredCityKey === cityKey) return;
  state.climateHoveredCityKey = cityKey;
  syncClimateChartStyles();
}

function syncClimateChartStyles() {
  state.climateCharts.forEach((chart) => {
    chart.data.datasets.forEach((dataset) => {
      const isHidden = state.climateHiddenCities.has(dataset.cityKey);
      const isHovered =
        state.climateHoveredCityKey &&
        state.climateHoveredCityKey === dataset.cityKey;
      const shouldFade =
        state.climateHoveredCityKey && !isHovered;
      const baseColor = getBaseDatasetColor(dataset);
      const isLowVariant = dataset.variant === "low";
      const alpha = isHidden
        ? 0.06
        : shouldFade
          ? isLowVariant ? 0.18 : 0.3
          : isHovered
            ? isLowVariant ? 0.86 : 1
            : isLowVariant ? 0.5 : 0.76;

      dataset.hidden = isHidden;
      dataset.borderColor = withAlpha(baseColor, alpha);
      dataset.backgroundColor = withAlpha(
        baseColor,
        isHidden ? 0.03 : shouldFade ? 0.05 : isLowVariant ? 0.08 : 0.15,
      );
      dataset.borderWidth = isHovered
        ? isLowVariant ? 2.3 : 3.3
        : isLowVariant ? 1.7 : 2.2;
      dataset.pointRadius = isHovered ? 2.5 : 0;
    });
    chart.update("none");
  });
}

function renderClimateTooltip(context, options) {
  const { chart, tooltip } = context;
  const tooltipEl = getClimateTooltipElement(chart);

  if (tooltip.opacity === 0 || !tooltip.dataPoints?.length) {
    tooltipEl.classList.remove("is-visible");
    return;
  }

  const dataPoint = tooltip.dataPoints[0];
  const dataset = dataPoint.dataset;
  const cityKey = dataset.cityKey;
  const cityLabel = dataset.cityLabel;
  const monthLabel = dataPoint.label;
  const color = dataset._baseColor ?? "#2563eb";

  let contentHtml = "";
  if (options.chartKey === "temperature") {
    const highDataset = chart.data.datasets.find(
      (item) => item.cityKey === cityKey && item.variant === "high",
    );
    const lowDataset = chart.data.datasets.find(
      (item) => item.cityKey === cityKey && item.variant === "low",
    );
    const highValue = toNumber(highDataset?.data?.[dataPoint.dataIndex]);
    const lowValue = toNumber(lowDataset?.data?.[dataPoint.dataIndex]);
    const highParts = formatTooltipValueParts(highValue, "°C");
    const lowParts = formatTooltipValueParts(lowValue, "°C");
    const highHtml = `
      <span class="chart-tooltip-value-number">${escapeHtml(highParts.value)}</span>${highParts.unit ? ` <span class="chart-tooltip-unit">${escapeHtml(highParts.unit)}</span>` : ""}
    `;
    const lowHtml = `
      <span class="chart-tooltip-value-number">${escapeHtml(lowParts.value)}</span>${lowParts.unit ? ` <span class="chart-tooltip-unit">${escapeHtml(lowParts.unit)}</span>` : ""}
    `;
    contentHtml = `
      <div class="chart-tooltip-city">${escapeHtml(cityLabel)}</div>
      <div class="chart-tooltip-text">${escapeHtml(monthLabel)}:</div>
      <div class="chart-tooltip-metric">
        <span class="chart-tooltip-key">High:</span>
        <span class="chart-tooltip-value">${highHtml}</span>
      </div>
      <div class="chart-tooltip-metric">
        <span class="chart-tooltip-key">Low:</span>
        <span class="chart-tooltip-value">${lowHtml}</span>
      </div>
    `;
  } else {
    const value = toNumber(dataPoint.parsed?.y);
    const valueParts = formatTooltipValueParts(value, options.yTitle);
    const valueHtml = valueParts.unit
      ? `${escapeHtml(valueParts.value)} <span class="chart-tooltip-unit">${escapeHtml(valueParts.unit)}</span>`
      : escapeHtml(valueParts.value);
    contentHtml = `
      <div class="chart-tooltip-city">${escapeHtml(cityLabel)}</div>
      <div class="chart-tooltip-metric chart-tooltip-metric-inline">
        <span class="chart-tooltip-key">${escapeHtml(monthLabel)}</span>
        <span class="chart-tooltip-value">${valueHtml}</span>
      </div>
    `;
  }

  tooltipEl.innerHTML = `
    <div class="chart-tooltip-row">
      <span class="chart-tooltip-swatch" style="background:${escapeHtml(color)}"></span>
      <div>${contentHtml}</div>
    </div>
  `;

  const { offsetLeft: positionX, offsetTop: positionY } = chart.canvas;
  tooltipEl.style.left = `${positionX + tooltip.caretX + 14}px`;
  tooltipEl.style.top = `${positionY + tooltip.caretY - 12}px`;
  tooltipEl.classList.add("is-visible");
}

function getClimateTooltipElement(chart) {
  const parent = chart.canvas.parentNode;
  let tooltipEl = parent.querySelector(".chart-tooltip");

  if (!tooltipEl) {
    tooltipEl = document.createElement("div");
    tooltipEl.className = "chart-tooltip";
    parent.appendChild(tooltipEl);
  }

  return tooltipEl;
}

function formatTooltipValue(value, unit) {
  if (value == null) return "—";
  const raw = String(value);
  if (unit === "%" || unit === "mm") {
    return `${raw}${unit}`;
  }
  if (unit === "km/h") {
    return `${raw} km/h`;
  }
  if (unit === "UV") {
    return raw;
  }
  if (unit === "Hours") {
    return `${raw} h`;
  }
  if (unit === "°C") {
    return `${raw}${unit}`;
  }
  return `${raw} ${unit}`;
}

function formatTooltipValueParts(value, unit) {
  if (value == null) {
    return { value: "—", unit: "" };
  }

  const raw = String(value);
  if (unit === "%" || unit === "mm" || unit === "°C") {
    return { value: raw, unit };
  }
  if (unit === "" || unit === "UV") {
    return { value: raw, unit: "" };
  }
  if (unit === "km/h") {
    return { value: raw, unit: "km/h" };
  }
  if (unit === "Hours") {
    return { value: raw, unit: "h" };
  }
  return { value: raw, unit };
}

function formatIndexValue(value, row = null) {
  if (value == null) return "—";
  if (row?.format === "currency") {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      maximumFractionDigits: value >= 100 ? 0 : 2,
      minimumFractionDigits: value >= 100 ? 0 : 2,
    }).format(value);
  }
  return String(value);
}

function bindLegatumTooltipEvents() {
  els.legatumIndicesBody
    .querySelectorAll("[data-legatum-row-key][data-legatum-country-code]")
    .forEach((cell) => {
      cell.addEventListener("mouseenter", () => {
        showLegatumTooltip(cell);
      });
      cell.addEventListener("mousemove", (event) => {
        positionLegatumTooltip(event.clientX, event.clientY);
      });
      cell.addEventListener("mouseleave", hideLegatumTooltip);
    });
}

function showLegatumTooltip(cell) {
  const rowKey = cell.dataset.legatumRowKey;
  const countryCode = cell.dataset.legatumCountryCode;
  if (!rowKey || !countryCode) return;

  const country = state.countryComparisonData.find(
    (item) => item.country_code === countryCode,
  );
  const row = legatumIndexRows.find((item) => item.key === rowKey);
  if (!country || !row) return;

  const tooltipEl = getLegatumTooltipElement();
  const isRankMode = state.legatumMetricMode === "rank";
  const series = buildLegatumSeries(
    country.legatum_indices?.[rowKey],
    isRankMode ? "rank" : "score",
  );
  if (!series.length) {
    tooltipEl.classList.remove("is-visible");
    return;
  }

  const values = series.map((point) => point.value);
  const minValue = Math.min(...values);
  const maxValue = Math.max(...values);
  const currentPoint = series[series.length - 1];
  const displayCountry = getDisplayCountryName(
    country.country ?? country.country_name ?? country.country_code ?? countryCode,
  );

  tooltipEl.innerHTML = `
    <div class="legatum-tooltip-title">${escapeHtml(displayCountry)}</div>
    <div class="legatum-tooltip-subtitle">${escapeHtml(row.label)} • ${escapeHtml(isRankMode ? "Rank" : "Score")}</div>
    <div class="legatum-tooltip-current">${escapeHtml(String(currentPoint.year))}: ${escapeHtml(String(currentPoint.value))}</div>
    <div class="legatum-tooltip-chart">
      ${renderLegatumSparkline(series, { minValue, maxValue, isRankMode })}
    </div>
  `;
  tooltipEl.classList.add("is-visible");
}

function hideLegatumTooltip() {
  getLegatumTooltipElement().classList.remove("is-visible");
}

function positionLegatumTooltip(clientX, clientY) {
  const tooltipEl = getLegatumTooltipElement();
  if (!tooltipEl.classList.contains("is-visible")) return;

  const offset = 16;
  const { innerWidth, innerHeight } = window;
  const rect = tooltipEl.getBoundingClientRect();
  let left = clientX + offset;
  let top = clientY + offset;

  if (left + rect.width > innerWidth - 12) {
    left = clientX - rect.width - offset;
  }
  if (top + rect.height > innerHeight - 12) {
    top = clientY - rect.height - offset;
  }

  tooltipEl.style.left = `${Math.max(12, left)}px`;
  tooltipEl.style.top = `${Math.max(12, top)}px`;
}

function getLegatumTooltipElement() {
  let tooltipEl = document.querySelector(".legatum-tooltip");
  if (!tooltipEl) {
    tooltipEl = document.createElement("div");
    tooltipEl.className = "legatum-tooltip";
    document.body.appendChild(tooltipEl);
  }
  return tooltipEl;
}

function buildLegatumSeries(entry, mode) {
  if (!entry) return [];

  return legatumYears
    .map((year) => ({
      year,
      value: toNumber(entry[`${mode}_${year}`]),
    }))
    .filter((point) => point.value != null);
}

function renderLegatumSparkline(series, { minValue, maxValue, isRankMode }) {
  const width = 230;
  const height = 122;
  const padding = { top: 10, right: 30, bottom: 24, left: 8 };
  const plotWidth = width - padding.left - padding.right;
  const plotHeight = height - padding.top - padding.bottom;
  const range = maxValue - minValue || 1;
  const midYear = 2015;
  const midX =
    padding.left +
    (plotWidth * (midYear - legatumYears[0])) /
      Math.max(1, legatumYears.length - 1);
  const midValue = (minValue + maxValue) / 2;
  const midNormalized = isRankMode
    ? (maxValue - midValue) / range
    : (midValue - minValue) / range;
  const midY = padding.top + plotHeight - midNormalized * plotHeight;

  const points = series.map((point, index) => {
    const x =
      padding.left +
      (plotWidth * index) / Math.max(1, series.length - 1);
    const normalized = isRankMode
      ? (maxValue - point.value) / range
      : (point.value - minValue) / range;
    const y = padding.top + plotHeight - normalized * plotHeight;
    return { ...point, x, y };
  });

  const polylinePoints = points.map((point) => `${point.x},${point.y}`).join(" ");
  const stroke = isRankMode ? "#2563eb" : "#0f766e";
  const topAxisValue = isRankMode ? minValue : maxValue;
  const bottomAxisValue = isRankMode ? maxValue : minValue;

  return `
    <svg viewBox="0 0 ${width} ${height}" class="legatum-tooltip-sparkline" aria-hidden="true">
      <line x1="${padding.left}" y1="${padding.top}" x2="${width - padding.right}" y2="${padding.top}" class="legatum-tooltip-grid" />
      <line x1="${padding.left}" y1="${midY}" x2="${width - padding.right}" y2="${midY}" class="legatum-tooltip-grid legatum-tooltip-grid-mid" />
      <line x1="${padding.left}" y1="${padding.top + plotHeight}" x2="${width - padding.right}" y2="${padding.top + plotHeight}" class="legatum-tooltip-grid" />
      <line x1="${midX}" y1="${padding.top + plotHeight}" x2="${midX}" y2="${padding.top + plotHeight + 5}" class="legatum-tooltip-axis-tick legatum-tooltip-axis-tick-mid" />
      <polyline points="${polylinePoints}" fill="none" stroke="${stroke}" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
      <circle cx="${points[0].x}" cy="${points[0].y}" r="2.5" fill="${stroke}" />
      <circle cx="${points[points.length - 1].x}" cy="${points[points.length - 1].y}" r="3" fill="${stroke}" />
      <text x="${width - 2}" y="${padding.top + 4}" text-anchor="end" class="legatum-tooltip-y-label">${escapeHtml(String(topAxisValue))}</text>
      <text x="${width - 2}" y="${midY + 3}" text-anchor="end" class="legatum-tooltip-y-label legatum-tooltip-y-label-mid">${escapeHtml(formatLegatumAxisMidValue(midValue))}</text>
      <text x="${width - 2}" y="${padding.top + plotHeight}" text-anchor="end" dominant-baseline="ideographic" class="legatum-tooltip-y-label">${escapeHtml(String(bottomAxisValue))}</text>
      <text x="${padding.left}" y="${height - 4}" text-anchor="start" class="legatum-tooltip-x-label">2007</text>
      <text x="${midX}" y="${height - 4}" text-anchor="middle" class="legatum-tooltip-x-label legatum-tooltip-x-label-mid">2015</text>
      <text x="${width - padding.right}" y="${height - 4}" text-anchor="end" class="legatum-tooltip-x-label">2023</text>
    </svg>
  `;
}

function formatLegatumAxisMidValue(value) {
  if (!Number.isFinite(value)) return "—";
  return String(Math.round(value));
}

function getBaseDatasetColor(dataset) {
  return dataset._baseColor ?? "#2563eb";
}

function withAlpha(color, alpha) {
  const hex = color.replace("#", "");
  const normalized =
    hex.length === 3
      ? hex
          .split("")
          .map((char) => char + char)
          .join("")
      : hex;
  const r = parseInt(normalized.slice(0, 2), 16);
  const g = parseInt(normalized.slice(2, 4), 16);
  const b = parseInt(normalized.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
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
  state.indicesSort = null;
  state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);

  renderCostBreakdownTable();
  renderNumbeoIndicesTable();
}

function applyIndicesSort(key) {
  if (!state.comparisonData.length) return;

  const nextDirection =
    state.indicesSort?.key === key && state.indicesSort.direction === "desc"
      ? "asc"
      : "desc";

  sortComparisonDataByIndexKey(key, nextDirection);
  state.indicesSort = { key, direction: nextDirection };
  state.breakdownSort = null;
  state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);

  renderCostBreakdownTable();
  renderNumbeoIndicesTable();
}

function applyCountryIndicesSort(key) {
  if (!state.countryComparisonData.length) return;

  const nextDirection =
    state.countryIndicesSort?.key === key &&
    state.countryIndicesSort.direction === "desc"
      ? "asc"
      : "desc";

  sortCountryComparisonDataByIndexKey(key, nextDirection);
  state.countryIndicesSort = { key, direction: nextDirection };
  state.legatumIndicesSort = null;

  renderCountryNumbeoIndicesTable();
  renderLegatumIndicesTable();
}

function applyLegatumIndicesSort(key) {
  if (!state.countryComparisonData.length) return;

  const nextDirection =
    state.legatumIndicesSort?.key === key &&
    state.legatumIndicesSort.direction === "desc"
      ? "asc"
      : "desc";

  sortCountryComparisonDataByLegatumKey(key, nextDirection);
  state.legatumIndicesSort = { key, direction: nextDirection };
  state.countryIndicesSort = null;

  renderCountryNumbeoIndicesTable();
  renderLegatumIndicesTable();
}

function setLegatumMetricMode(mode) {
  if (mode !== "score" && mode !== "rank") return;
  if (state.legatumMetricMode === mode) return;

  state.legatumMetricMode = mode;
  state.legatumIndicesSort = null;
  renderLegatumIndicesTable();
}

function applyLoadedComparisonSelection() {
  if (!state.comparisonData.length) return;

  const selectedIDs = new Set(Array.from(state.selectedCityIds));
  state.comparisonData = state.comparisonData.filter((city) =>
    selectedIDs.has(String(city.geoname_id ?? city.city_id ?? city.id ?? "")),
  );

  if (state.comparisonData.length === 0) {
    state.countryComparisonData = [];
    state.costBreakdownData = [];
    state.breakdownSort = null;
    state.indicesSort = null;
    state.countryIndicesSort = null;
    state.legatumIndicesSort = null;
    renderCostBreakdownTable();
    renderNumbeoIndicesTable();
    renderCountryNumbeoIndicesTable();
    renderLegatumIndicesTable();
    renderClimateChart();
    return;
  }

  syncCountryComparisonDataToSelectedCities();
  state.costBreakdownData = buildCostBreakdownDataset(state.comparisonData);
  renderCostBreakdownTable();
  renderNumbeoIndicesTable();
  renderCountryNumbeoIndicesTable();
  renderLegatumIndicesTable();
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

function sortComparisonDataByIndexKey(key, direction) {
  const multiplier = direction === "asc" ? 1 : -1;

  state.comparisonData = state.comparisonData
    .map((city, index) => ({
      city,
      index,
      value: toNumber(city.numbeo_indices?.[key]),
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

function sortCountryComparisonDataByIndexKey(key, direction) {
  const multiplier = direction === "asc" ? 1 : -1;

  state.countryComparisonData = state.countryComparisonData
    .map((country, index) => ({
      country,
      index,
      value: toNumber(country.numbeo_indices?.[key]),
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
    .map((entry) => entry.country);
}

function sortCountryComparisonDataByLegatumKey(key, direction) {
  const multiplier = direction === "asc" ? 1 : -1;
  const metricKey =
    state.legatumMetricMode === "rank" ? "rank_2023" : "score_2023";

  state.countryComparisonData = state.countryComparisonData
    .map((country, index) => ({
      country,
      index,
      value: toNumber(country.legatum_indices?.[key]?.[metricKey]),
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
    .map((entry) => entry.country);
}

function syncLegatumToggleButtons() {
  const isScoreMode = state.legatumMetricMode === "score";
  els.legatumScoreToggle.classList.toggle("is-active", isScoreMode);
  els.legatumRankToggle.classList.toggle("is-active", !isScoreMode);
}

function getCostParamValue(city, param) {
  const prices = city.numbeo_cost?.prices;
  if (!Array.isArray(prices)) return null;

  const entry = prices.find((price) => price?.param === param);
  return toNumber(entry?.cost);
}

async function loadCountryComparisonData(cities) {
  const countryCodes = Array.from(
    new Set(
      cities
        .map((city) => city.country_code)
        .filter(Boolean),
    ),
  );

  if (countryCodes.length === 0) return [];

  const qs = new URLSearchParams();
  qs.set("country_codes", countryCodes.join(","));
  qs.set("include", "numbeo_indices,legatum_indices");

  const res = await fetch(`/countries?${qs.toString()}`);
  if (!res.ok) {
    throw new Error(`Failed to load countries: ${res.status}`);
  }

  const data = await res.json();
  const countries = Array.isArray(data.countries)
    ? data.countries
    : Array.isArray(data.data)
      ? data.data
      : [];

  return sortCountriesBySelectedCityCountryOrder(countries, countryCodes);
}

function syncCountryComparisonDataToSelectedCities() {
  const selectedCountryCodes = Array.from(
    new Set(
      state.comparisonData
        .map((city) => city.country_code)
        .filter(Boolean),
    ),
  );
  const selectedCountryCodeSet = new Set(selectedCountryCodes);

  state.countryComparisonData = sortCountriesBySelectedCityCountryOrder(
    state.countryComparisonData.filter((country) =>
      selectedCountryCodeSet.has(country.country_code),
    ),
    selectedCountryCodes,
  );
}

function sortCountriesBySelectedCityCountryOrder(countries, orderedCountryCodes) {
  const positionByCode = new Map(
    orderedCountryCodes.map((code, index) => [code, index]),
  );

  return [...countries].sort((a, b) => {
    const aPos = positionByCode.get(a.country_code);
    const bPos = positionByCode.get(b.country_code);
    if (aPos != null && bPos != null) return aPos - bPos;
    if (aPos != null) return -1;
    if (bPos != null) return 1;
    return String(a.country ?? a.country_code ?? "").localeCompare(
      String(b.country ?? b.country_code ?? ""),
    );
  });
}

function getDisplayCountryName(country) {
  const raw = String(country ?? "");
  return countryDisplayNames[raw] ?? raw;
}

function formatFilterCityName(city) {
  const cityName = city.city ?? city.name ?? "";
  const countryCode = city.country_code ?? "";
  const stateCode = city.state_code ?? "";

  if (countryCode !== "USA" || !stateCode.startsWith("US-")) {
    return escapeHtml(cityName);
  }

  const shortStateCode = stateCode.slice(3);
  if (!shortStateCode) {
    return escapeHtml(cityName);
  }

  return `${escapeHtml(cityName)} <span class="selection-title-meta">${escapeHtml(shortStateCode)}</span>`;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
