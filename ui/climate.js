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

function getBaseDatasetColor(dataset) {
  return dataset._baseColor ?? "#2563eb";
}

Object.assign(window, {
  renderClimateChart,
  destroyClimateCharts,
});
