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
  versionStatusSummary,
  versionStatusSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", status: "todo" },
  { id: "b", version_id: "version_2026_7", status: "done" },
  { id: "c", version_id: "version_2026_7", status: "todo" },
  { id: "d", version_id: "version_2026_7", status: "" },
  { id: "e", version_id: "version_2026_8", status: "blocked" },
  { id: "f", version_id: "", status: "done" }
];

assert.deepStrictEqual(versionStatusSummary(tickets, "version_2026_7"), {
  total: 4,
  done: 1,
  open: 3,
  statuses: [
    { status: "todo", count: 2 },
    { status: "done", count: 1 },
    { status: "No status", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionStatusSummaryItems(versionStatusSummary(tickets, "version_2026_7")),
  ["total: 4", "open: 3", "done: 1", "todo: 2", "done: 1", "No status: 1"]
);

assert.deepStrictEqual(versionStatusSummary(tickets, "version_missing"), {
  total: 0,
  done: 0,
  open: 0,
  statuses: []
});
assert.deepStrictEqual(versionStatusSummary(null, "version_2026_7"), {
  total: 0,
  done: 0,
  open: 0,
  statuses: []
});
assert.deepStrictEqual(versionStatusSummaryItems(null), ["total: 0", "open: 0", "done: 0"]);
