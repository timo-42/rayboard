const assert = require("assert");

const noop = () => {};
const element = {
  addEventListener: noop,
  append: noop,
  replaceChildren: noop,
  querySelector: () => element,
  querySelectorAll: () => [],
  closest: () => null,
  classList: { add: noop, remove: noop, toggle: noop },
  dataset: {},
  elements: {},
  style: {},
  setAttribute: noop
};

global.document = {
  cookie: "",
  addEventListener: noop,
  createElement: () => ({ ...element, classList: { add: noop, remove: noop, toggle: noop }, dataset: {}, style: {} }),
  querySelector: () => element,
  querySelectorAll: () => [],
  body: { dataset: {} }
};
global.window = {
  addEventListener: noop,
  history: { pushState: noop },
  location: { pathname: "/", search: "" }
};

const { versionReportStartDateBreakdown } = require("./static/app.js");

assert.deepStrictEqual(versionReportStartDateBreakdown([
  { start_date: "2026-06-19" },
  { start_date: "2026-06-20" },
  { start_date: "2026-06-22" },
  { start_date: "2026-06-30" },
  { start_date: "" },
  { start_date: "not-a-date" }
], "2026-06-20"), [
  { key: "started", label: "Started", count: 1 },
  { key: "today", label: "Starts today", count: 1 },
  { key: "soon", label: "Starts soon", count: 1 },
  { key: "future", label: "Future start", count: 1 },
  { key: "none", label: "No start date", count: 2 }
]);

assert.deepStrictEqual(versionReportStartDateBreakdown([], "2026-06-20"), []);
assert.deepStrictEqual(versionReportStartDateBreakdown(null, "2026-06-20"), []);
assert.deepStrictEqual(versionReportStartDateBreakdown({}, "2026-06-20"), []);
