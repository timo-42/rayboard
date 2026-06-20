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

const { versionReportLabelBreakdown } = require("./static/app.js");

assert.deepStrictEqual(versionReportLabelBreakdown([
  { labels: ["backend", "api"] },
  { labels: ["api"] },
  { labels: [] },
  { labels: ["docs", "backend"] },
  { labels: null }
]), [
  { label: "api", count: 2 },
  { label: "backend", count: 2 },
  { label: "No labels", count: 2 },
  { label: "docs", count: 1 }
]);

assert.deepStrictEqual(versionReportLabelBreakdown([
  { labels: ["zeta"] },
  { labels: ["alpha"] },
  { labels: [] }
]), [
  { label: "alpha", count: 1 },
  { label: "zeta", count: 1 },
  { label: "No labels", count: 1 }
]);

assert.deepStrictEqual(versionReportLabelBreakdown([]), []);
assert.deepStrictEqual(versionReportLabelBreakdown(null), []);
assert.deepStrictEqual(versionReportLabelBreakdown({}), []);
