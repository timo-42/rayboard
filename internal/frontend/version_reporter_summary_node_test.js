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
  versionReporterSummary,
  versionReporterSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", reporter_id: "user_ada" },
  { id: "b", version_id: "version_2026_7", reporter_id: "user_ada" },
  { id: "c", version_id: "version_2026_7", reporter_id: "user_grace" },
  { id: "d", version_id: "version_2026_7", reporter_id: "" },
  { id: "e", version_id: "version_2026_8", reporter_id: "user_ada" },
  { id: "f", version_id: "", reporter_id: "user_grace" }
];

assert.deepStrictEqual(versionReporterSummary(tickets, "version_2026_7"), {
  total: 4,
  reporters: [
    { label: "reporter user_ada", count: 2 },
    { label: "reporter user_grace", count: 1 },
    { label: "No reporter", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionReporterSummaryItems(versionReporterSummary(tickets, "version_2026_7")),
  ["total: 4", "reporter user_ada: 2", "reporter user_grace: 1", "No reporter: 1"]
);

assert.deepStrictEqual(versionReporterSummary(tickets, "version_missing"), {
  total: 0,
  reporters: []
});
assert.deepStrictEqual(versionReporterSummary(null, "version_2026_7"), {
  total: 0,
  reporters: []
});
assert.deepStrictEqual(versionReporterSummaryItems(null), ["total: 0"]);
