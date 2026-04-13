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

function toNumber(value) {
  if (value == null) return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
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
    return formatCurrency(value, row.currencyCode ?? "USD");
  }
  return String(value);
}

function fmt(value) {
  const n = toNumber(value);
  return n == null ? "—" : n.toFixed(1);
}

const currencyLocales = {
  USD: "en-US",
  EUR: "de-DE",
  GBP: "en-GB",
  RUB: "ru-RU",
  CAD: "en-CA",
  AUD: "en-AU",
  JPY: "ja-JP",
  CNY: "zh-CN",
  INR: "en-IN",
  BRL: "pt-BR",
};

const currencyFractionDigits = {
  USD: 1,
  EUR: 1,
  GBP: 1,
  CAD: 1,
  AUD: 1,
  CNY: 1,
  BRL: 1,
  RUB: 0,
  JPY: 0,
  INR: 0,
};

function formatCurrency(value, currencyCode) {
  const amount = toNumber(value);
  if (amount == null) return "—";

  const code = String(currencyCode || "USD").toUpperCase();
  const locale = currencyLocales[code] ?? currencyLocales.USD;
  const fractionDigits =
    currencyFractionDigits[code] ?? currencyFractionDigits.USD;

  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency: code,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(amount);
}

function formatCostValue(value, options = {}) {
  if (!value) return "—";
  const n = toNumber(value.cost);
  if (n == null) return "—";

  if (options.isMortgageRate) {
    return `${n.toFixed(2)}%`;
  }

  const amount = toNumber(options.amount ?? n);
  if (amount == null) return "—";
  return formatCurrency(amount, options.currencyCode ?? value.currency ?? "USD");
}

function formatLegatumAxisMidValue(value) {
  if (!Number.isFinite(value)) return "—";
  return String(Math.round(value));
}

function convertUSDAmount(amountUSD, rate) {
  const amount = toNumber(amountUSD);
  const fxRate = toNumber(rate);
  if (amount == null || fxRate == null) return null;
  return amount * fxRate;
}

Object.assign(window, {
  getDisplayCountryName,
  escapeHtml,
  toNumber,
  withAlpha,
  formatCompactPopulation,
  formatCompactArea,
  formatUTCOffset,
  formatTooltipValue,
  formatTooltipValueParts,
  formatIndexValue,
  fmt,
  formatCurrency,
  formatCostValue,
  formatLegatumAxisMidValue,
  convertUSDAmount,
});
