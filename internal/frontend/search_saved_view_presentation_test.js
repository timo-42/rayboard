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
  savedViewConfigurationInsightItems,
  savedViewFieldUsageInsightItems,
  savedViewFieldUsageSummary,
  savedViewOverviewSummary,
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

assert.deepStrictEqual(
  savedViewOverviewSummary([
    {
      scope_type: "project",
      display_mode: "board",
      pinned: true,
      query: { text: "auth", filter: "status == 'todo'" },
      group_by: "priority",
      columns: ["key", "title"],
      sort: [{ field: "updated_at", direction: "desc" }]
    },
    {
      scope_type: "user",
      display_mode: "backlog",
      query: { text: "", filter: "" },
      columns: [],
      sort: []
    },
    {
      scope_type: "global",
      display_mode: "list",
      query: { filter: "priority == 'High'" },
      columns: ["key"],
      sort: [{ field: "priority", direction: "asc" }]
    }
  ]),
  {
    total: 3,
    pinned: 1,
    scopes: [
      { key: "global", count: 1 },
      { key: "project", count: 1 },
      { key: "user", count: 1 }
    ],
    modes: [
      { key: "backlog", count: 1 },
      { key: "board", count: 1 },
      { key: "list", count: 1 }
    ],
    configuration: {
      text_queries: 1,
      cel_filters: 2,
      grouped: 1,
      column_configured: 2,
      sorted: 2,
      project_scoped: 1,
      pinned: 1,
      board_mode: 1,
      backlog_mode: 1
    },
    field_usage: {
      columns: [
        { field: "key", label: "Key", count: 2 },
        { field: "title", label: "Title", count: 1 }
      ],
      sorts: [
        { field: "priority", label: "Priority", count: 1 },
        { field: "updated_at", label: "Updated", count: 1 }
      ],
      groups: [
        { field: "priority", label: "Priority", count: 1 }
      ]
    }
  }
);

assert.deepStrictEqual(
  savedViewConfigurationInsightItems({
    text_queries: 1,
    cel_filters: 2,
    grouped: 1,
    column_configured: 2,
    sorted: 2,
    project_scoped: 1,
    pinned: 1,
    board_mode: 1,
    backlog_mode: 1
  }),
  [
    "text queries: 1",
    "CEL filters: 2",
    "grouped: 1",
    "columns: 2",
    "sorted: 2",
    "project-scoped: 1",
    "pinned: 1",
    "board mode: 1",
    "backlog mode: 1"
  ]
);

assert.deepStrictEqual(
  savedViewFieldUsageSummary([
    { columns: ["key", "priority", "key"], sort: [{ field: "updated_at" }, { field: "priority" }], group_by: "priority" },
    { columns: ["status"], sort: [{ field: "updated_at" }], group_by: "" },
    { columns: [], sort: [{ field: "" }], group_by: "status" }
  ]),
  {
    columns: [
      { field: "key", label: "Key", count: 2 },
      { field: "priority", label: "Priority", count: 1 },
      { field: "status", label: "Status", count: 1 }
    ],
    sorts: [
      { field: "updated_at", label: "Updated", count: 2 },
      { field: "priority", label: "Priority", count: 1 }
    ],
    groups: [
      { field: "priority", label: "Priority", count: 1 },
      { field: "status", label: "Status", count: 1 }
    ]
  }
);

assert.deepStrictEqual(
  savedViewFieldUsageInsightItems({
    columns: [{ label: "Key", count: 2 }, { label: "Priority", count: 1 }],
    sorts: [{ label: "Updated", count: 2 }],
    groups: []
  }),
  ["columns Key: 2", "columns Priority: 1", "sorts Updated: 2", "groups: none"]
);
assert.deepStrictEqual(savedViewFieldUsageSummary(null), { columns: [], sorts: [], groups: [] });
