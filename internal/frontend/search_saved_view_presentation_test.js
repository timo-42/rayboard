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
  groupedSearchResults,
  savedViewSearchPresentation,
  searchResultColumnLabel,
  searchResultColumnValue,
  searchResultColumns,
  searchResultFallbackMetadata
} = require("./static/app.js");

assert.deepStrictEqual(
  searchResultColumns(["key", "title", "status", "status", "not_a_column", "story_points"]),
  ["key", "title", "status", "story_points"]
);
assert.deepStrictEqual(searchResultColumns([]), []);
assert.deepStrictEqual(searchResultColumns(null), []);

assert.strictEqual(searchResultColumnLabel("story_points"), "Story points");
assert.strictEqual(searchResultColumnLabel("assignee_id"), "Assignee");
assert.strictEqual(searchResultColumnLabel("custom_name"), "Custom Name");

const ticket = {
  id: "ticket-1",
  key: "CORE-1",
  title: "Fix login",
  description: "",
  status: "todo",
  priority: "High",
  type: "Bug",
  assignee_id: "",
  sprint_id: "sprint-1",
  component_id: "",
  version_id: "",
  labels: ["backend", "auth"],
  story_points: 3,
  created_at: "2026-06-20T05:00:00Z",
  updated_at: ""
};

assert.strictEqual(searchResultColumnValue(ticket, "key"), "CORE-1");
assert.strictEqual(searchResultColumnValue(ticket, "labels"), "backend, auth");
assert.strictEqual(searchResultColumnValue(ticket, "assignee_id"), "Unassigned");
assert.strictEqual(searchResultColumnValue(ticket, "story_points"), "3 pt");
assert.strictEqual(searchResultColumnValue(ticket, "description"), "No description");
assert.strictEqual(searchResultFallbackMetadata(ticket), "todo / Bug / High / 3 pt");

assert.deepStrictEqual(
  groupedSearchResults([
    { key: "CORE-1", priority: "High" },
    { key: "CORE-2", priority: "" },
    { key: "CORE-3", priority: "Low" },
    { key: "CORE-4", priority: "High" }
  ], "priority").map((group) => ({
    label: group.label,
    keys: group.tickets.map((item) => item.key)
  })),
  [
    { label: "High", keys: ["CORE-1", "CORE-4"] },
    { label: "Low", keys: ["CORE-3"] },
    { label: "No priority", keys: ["CORE-2"] }
  ]
);
assert.deepStrictEqual(groupedSearchResults([{ key: "CORE-1", priority: "High" }], "unsupported"), []);

assert.deepStrictEqual(
  savedViewSearchPresentation({
    id: "view-1",
    name: "Open bugs",
    columns: ["key", "priority", "bogus"],
    group_by: "priority"
  }),
  {
    view_id: "view-1",
    name: "Open bugs",
    columns: ["key", "priority"],
    group_by: "priority"
  }
);

assert.deepStrictEqual(
  savedViewSearchPresentation({ columns: [], group_by: "bogus" }),
  { view_id: "", name: "", columns: [], group_by: "" }
);
