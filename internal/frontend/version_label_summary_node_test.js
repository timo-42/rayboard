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

const {
  versionLabelSummary,
  versionLabelSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", labels: ["backend", "api"] },
  { id: "b", version_id: "version_2026_7", labels: ["backend"] },
  { id: "c", version_id: "version_2026_7", labels: [] },
  { id: "d", version_id: "version_2026_7", labels: ["frontend", "api"] },
  { id: "e", version_id: "version_2026_8", labels: ["api"] },
  { id: "f", version_id: "", labels: ["backend"] }
];

assert.deepStrictEqual(versionLabelSummary(tickets, "version_2026_7"), {
  total: 4,
  labels: [
    { label: "api", count: 2 },
    { label: "backend", count: 2 },
    { label: "frontend", count: 1 },
    { label: "No labels", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionLabelSummaryItems(versionLabelSummary(tickets, "version_2026_7")),
  ["total: 4", "api: 2", "backend: 2", "frontend: 1", "No labels: 1"]
);

assert.deepStrictEqual(versionLabelSummary(tickets, "version_missing"), {
  total: 0,
  labels: []
});
assert.deepStrictEqual(versionLabelSummary(null, "version_2026_7"), {
  total: 0,
  labels: []
});
assert.deepStrictEqual(versionLabelSummaryItems(null), ["total: 0"]);
