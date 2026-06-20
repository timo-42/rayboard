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
  versionIssueTypeSummary,
  versionIssueTypeSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", type: "Bug" },
  { id: "b", version_id: "version_2026_7", type: "Story" },
  { id: "c", version_id: "version_2026_7", type: "Bug" },
  { id: "d", version_id: "version_2026_7", type: "" },
  { id: "e", version_id: "version_2026_8", type: "Task" },
  { id: "f", version_id: "", type: "Bug" }
];

assert.deepStrictEqual(versionIssueTypeSummary(tickets, "version_2026_7"), {
  total: 4,
  types: [
    { type: "Bug", count: 2 },
    { type: "No issue type", count: 1 },
    { type: "Story", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionIssueTypeSummaryItems(versionIssueTypeSummary(tickets, "version_2026_7")),
  ["total: 4", "Bug: 2", "No issue type: 1", "Story: 1"]
);

assert.deepStrictEqual(versionIssueTypeSummary(tickets, "version_missing"), {
  total: 0,
  types: []
});
assert.deepStrictEqual(versionIssueTypeSummary(null, "version_2026_7"), {
  total: 0,
  types: []
});
assert.deepStrictEqual(versionIssueTypeSummaryItems(null), ["total: 0"]);
