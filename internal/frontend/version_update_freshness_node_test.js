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

const { versionReportUpdateFreshness } = require("./static/app.js");

assert.deepStrictEqual(versionReportUpdateFreshness([
  { updated_at: "2026-06-20T10:00:00Z" },
  { updated_at: "2026-06-14T00:00:00Z" },
  { updated_at: "2026-06-01T00:00:00Z" },
  { updated_at: "2026-05-01T00:00:00Z" },
  { updated_at: "" },
  { updated_at: "not-a-date" }
], "2026-06-20"), [
  { key: "today", label: "Updated today", count: 1 },
  { key: "week", label: "Updated this week", count: 1 },
  { key: "stale", label: "Stale (8-30 days)", count: 1 },
  { key: "dormant", label: "Dormant (31+ days)", count: 1 },
  { key: "none", label: "No update date", count: 2 }
]);

assert.deepStrictEqual(versionReportUpdateFreshness([], "2026-06-20"), []);
assert.deepStrictEqual(versionReportUpdateFreshness(null, "2026-06-20"), []);
assert.deepStrictEqual(versionReportUpdateFreshness({}, "2026-06-20"), []);
