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

function renderTableHeaderRow({
  headRowEl,
  headLabel,
  data,
  getColumnLabel,
  getColumnMeta,
  getHeaderDataAttrs = null,
}) {
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
}

function buildHeatmapScale(values, totalCount) {
  const numericValues = values
    .map((value) => toNumber(value))
    .filter((value) => value != null);
  const minValue =
    numericValues.length > 0 ? Math.min(...numericValues) : null;
  const maxValue =
    numericValues.length > 0 ? Math.max(...numericValues) : null;
  const shouldApplySecondary =
    totalCount - numericValues.length < 5 &&
    numericValues.length >= 3;
  const sortedUniqueValues = shouldApplySecondary
    ? Array.from(new Set(numericValues)).sort((a, b) => a - b)
    : [];

  return {
    minValue,
    maxValue,
    shouldApplySecondary,
    secondMinValue:
      sortedUniqueValues.length >= 3 ? sortedUniqueValues[1] : null,
    secondMaxValue:
      sortedUniqueValues.length >= 3
        ? sortedUniqueValues[sortedUniqueValues.length - 2]
        : null,
  };
}

function getHeatmapClass(value, scale, better = "low") {
  const numericValue = toNumber(value);
  if (
    numericValue == null ||
    scale.minValue == null ||
    scale.maxValue == null ||
    scale.minValue === scale.maxValue
  ) {
    return "";
  }

  const lowerIsBetter = better === "low";

  if (lowerIsBetter) {
    if (numericValue === scale.minValue) return " cost-value-low";
    if (numericValue === scale.maxValue) return " cost-value-high";
    if (
      scale.shouldApplySecondary &&
      scale.secondMinValue != null &&
      numericValue === scale.secondMinValue
    ) {
      return " cost-value-low-soft";
    }
    if (
      scale.shouldApplySecondary &&
      scale.secondMaxValue != null &&
      numericValue === scale.secondMaxValue
    ) {
      return " cost-value-high-soft";
    }
    return "";
  }

  if (numericValue === scale.minValue) return " cost-value-high";
  if (numericValue === scale.maxValue) return " cost-value-low";
  if (
    scale.shouldApplySecondary &&
    scale.secondMinValue != null &&
    numericValue === scale.secondMinValue
  ) {
    return " cost-value-high-soft";
  }
  if (
    scale.shouldApplySecondary &&
    scale.secondMaxValue != null &&
    numericValue === scale.secondMaxValue
  ) {
    return " cost-value-low-soft";
  }
  return "";
}

function sortItemsByNumericValue(items, getValue, direction) {
  const multiplier = direction === "asc" ? 1 : -1;

  return items
    .map((item, index) => ({
      item,
      index,
      value: toNumber(getValue(item)),
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
    .map((entry) => entry.item);
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

function renderCostBreakdownTable() {
  if (!state.costBreakdownData.length || !state.comparisonData.length) {
    els.costBreakdownEmptyState.classList.remove("hidden");
    els.costBreakdownWrapper.classList.add("hidden");
    return;
  }

  els.costBreakdownEmptyState.classList.add("hidden");
  els.costBreakdownWrapper.classList.remove("hidden");

  renderTableHeaderRow({
    headRowEl: els.costBreakdownHeadRow,
    headLabel: "Item",
    data: state.comparisonData,
    getColumnLabel: (city) => city.city ?? city.name ?? "City",
    getColumnMeta: (city) =>
      getDisplayCountryName(
        city.country ?? city.country_name ?? city.country_code ?? "",
      ),
    getHeaderDataAttrs: (city) =>
      `data-header-kind="city" data-city-id="${escapeHtml(String(city.geoname_id ?? city.city_id ?? city.id ?? ""))}"`,
  });

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
          const scale = buildHeatmapScale(
            item.values.map((value) => value?.cost),
            item.values.length,
          );
          const cells = item.values
            .map((value) => {
              let cellClass = "cost-value";
              cellClass += getHeatmapClass(
                value?.cost,
                scale,
                isSalaryRow ? "high" : "low",
              );

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
      const scale = buildHeatmapScale(
        data.map((item) => getValue(item, row.key)),
        data.length,
      );

      const cells = data
        .map((item) => {
          const value = getValue(item, row.key);
          const cellDataAttrs = getCellDataAttrs
            ? getCellDataAttrs(item, row)
            : "";
          let cellClass = "indices-value";
          cellClass += getHeatmapClass(
            value,
            scale,
            isHigherBetter ? "high" : "low",
          );

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
  state.comparisonData = sortItemsByNumericValue(
    state.comparisonData,
    (city) => getCostParamValue(city, param),
    direction,
  );
}

function sortComparisonDataByIndexKey(key, direction) {
  state.comparisonData = sortItemsByNumericValue(
    state.comparisonData,
    (city) => city.numbeo_indices?.[key],
    direction,
  );
}

function sortCountryComparisonDataByIndexKey(key, direction) {
  state.countryComparisonData = sortItemsByNumericValue(
    state.countryComparisonData,
    (country) => country.numbeo_indices?.[key],
    direction,
  );
}

function sortCountryComparisonDataByLegatumKey(key, direction) {
  const metricKey =
    state.legatumMetricMode === "rank" ? "rank_2023" : "score_2023";

  state.countryComparisonData = sortItemsByNumericValue(
    state.countryComparisonData,
    (country) => country.legatum_indices?.[key]?.[metricKey],
    direction,
  );
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
