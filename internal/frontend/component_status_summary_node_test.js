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
  componentStatusSummary,
  componentStatusSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", component_id: "component_api", status: "todo" },
  { id: "b", component_id: "component_api", status: "done" },
  { id: "c", component_id: "component_api", status: "todo" },
  { id: "d", component_id: "component_api", status: "" },
  { id: "e", component_id: "component_web", status: "blocked" },
  { id: "f", component_id: "", status: "done" }
];

assert.deepStrictEqual(componentStatusSummary(tickets, "component_api"), {
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
  componentStatusSummaryItems(componentStatusSummary(tickets, "component_api")),
  ["total: 4", "open: 3", "done: 1", "todo: 2", "done: 1", "No status: 1"]
);

assert.deepStrictEqual(componentStatusSummary(tickets, "component_missing"), {
  total: 0,
  done: 0,
  open: 0,
  statuses: []
});
assert.deepStrictEqual(componentStatusSummary(null, "component_api"), {
  total: 0,
  done: 0,
  open: 0,
  statuses: []
});
assert.deepStrictEqual(componentStatusSummaryItems(null), ["total: 0", "open: 0", "done: 0"]);
