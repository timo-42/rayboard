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
  versionPrioritySummary,
  versionPrioritySummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", priority: "High" },
  { id: "b", version_id: "version_2026_7", priority: "Low" },
  { id: "c", version_id: "version_2026_7", priority: "High" },
  { id: "d", version_id: "version_2026_7", priority: "" },
  { id: "e", version_id: "version_2026_8", priority: "Medium" },
  { id: "f", version_id: "", priority: "High" }
];

assert.deepStrictEqual(versionPrioritySummary(tickets, "version_2026_7"), {
  total: 4,
  priorities: [
    { priority: "High", count: 2 },
    { priority: "Low", count: 1 },
    { priority: "No priority", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionPrioritySummaryItems(versionPrioritySummary(tickets, "version_2026_7")),
  ["total: 4", "High: 2", "Low: 1", "No priority: 1"]
);

assert.deepStrictEqual(versionPrioritySummary(tickets, "version_missing"), {
  total: 0,
  priorities: []
});
assert.deepStrictEqual(versionPrioritySummary(null, "version_2026_7"), {
  total: 0,
  priorities: []
});
assert.deepStrictEqual(versionPrioritySummaryItems(null), ["total: 0"]);
