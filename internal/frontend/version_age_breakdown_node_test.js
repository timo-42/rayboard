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

const { versionReportAgeBreakdown } = require("./static/app.js");

assert.deepStrictEqual(versionReportAgeBreakdown([
  { created_at: "2026-06-20T10:00:00Z" },
  { created_at: "2026-06-14T00:00:00Z" },
  { created_at: "2026-06-01T00:00:00Z" },
  { created_at: "2026-05-01T00:00:00Z" },
  { created_at: "" },
  { created_at: "not-a-date" }
], "2026-06-20"), [
  { key: "new", label: "New (0-7 days)", count: 2 },
  { key: "recent", label: "Recent (8-30 days)", count: 1 },
  { key: "older", label: "Older (31+ days)", count: 1 },
  { key: "none", label: "No created date", count: 2 }
]);

assert.deepStrictEqual(versionReportAgeBreakdown([], "2026-06-20"), []);
assert.deepStrictEqual(versionReportAgeBreakdown(null, "2026-06-20"), []);
assert.deepStrictEqual(versionReportAgeBreakdown({}, "2026-06-20"), []);
