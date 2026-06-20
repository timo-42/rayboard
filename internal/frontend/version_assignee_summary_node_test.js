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
  versionAssigneeSummary,
  versionAssigneeSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", assignee_id: "user_ada" },
  { id: "b", version_id: "version_2026_7", assignee_id: "user_ada" },
  { id: "c", version_id: "version_2026_7", assignee_id: "user_grace" },
  { id: "d", version_id: "version_2026_7", assignee_id: "" },
  { id: "e", version_id: "version_2026_8", assignee_id: "user_ada" },
  { id: "f", version_id: "", assignee_id: "user_grace" }
];

assert.deepStrictEqual(versionAssigneeSummary(tickets, "version_2026_7"), {
  total: 4,
  assignees: [
    { label: "assignee user_ada", count: 2 },
    { label: "assignee user_grace", count: 1 },
    { label: "Unassigned", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionAssigneeSummaryItems(versionAssigneeSummary(tickets, "version_2026_7")),
  ["total: 4", "assignee user_ada: 2", "assignee user_grace: 1", "Unassigned: 1"]
);

assert.deepStrictEqual(versionAssigneeSummary(tickets, "version_missing"), {
  total: 0,
  assignees: []
});
assert.deepStrictEqual(versionAssigneeSummary(null, "version_2026_7"), {
  total: 0,
  assignees: []
});
assert.deepStrictEqual(versionAssigneeSummaryItems(null), ["total: 0"]);
