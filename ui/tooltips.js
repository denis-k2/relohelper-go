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

Object.assign(window, {
  bindHeaderInfoTooltips,
  renderClimateTooltip,
  bindLegatumTooltipEvents,
});
