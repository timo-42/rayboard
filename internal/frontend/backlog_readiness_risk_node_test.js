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
  backlogReadinessSummary,
  backlogRiskSummary
} = require("./static/app.js");

const tickets = [
  {
    status: "todo",
    priority: "High",
    assignee_id: "",
    story_points: null,
    start_date: "",
    due_date: "2026-06-19",
    updated_at: "2026-06-01T09:00:00Z"
  },
  {
    status: "blocked",
    priority: "Medium",
    assignee_id: "user_1",
    story_points: 3,
    start_date: "2026-06-18",
    due_date: "2026-06-25",
    updated_at: "2026-06-20T09:00:00Z"
  },
  {
    status: "done",
    priority: "Critical",
    assignee_id: "",
    story_points: "",
    start_date: "",
    due_date: "2026-06-01",
    updated_at: "2026-05-01T09:00:00Z"
  },
  {
    status: "todo",
    priority: "Low",
    assignee_id: "   ",
    story_points: "bad",
    start_date: "not-a-date",
    due_date: "",
    updated_at: "2026-06-10T09:00:00Z"
  }
];

assert.deepStrictEqual(backlogReadinessSummary(tickets), [
  { key: "ready", label: "Ready", count: 1 },
  { key: "missing_assignee", label: "Missing assignee", count: 3 },
  { key: "missing_estimate", label: "Missing estimate", count: 3 },
  { key: "missing_start", label: "Missing start date", count: 3 },
  { key: "missing_due", label: "Missing due date", count: 1 }
]);

assert.deepStrictEqual(backlogRiskSummary(tickets, "2026-06-20"), [
  { key: "overdue_open", label: "Open overdue", count: 1 },
  { key: "stale_open", label: "Stale open", count: 2 },
  { key: "unassigned_open", label: "Unassigned open", count: 2 },
  { key: "unscheduled_open", label: "Unscheduled open", count: 2 },
  { key: "blocked_open", label: "Blocked open", count: 1 },
  { key: "high_priority_open", label: "High-priority open", count: 1 }
]);

assert.deepStrictEqual(backlogReadinessSummary([]), []);
assert.deepStrictEqual(backlogReadinessSummary(null), []);
assert.deepStrictEqual(backlogRiskSummary([], "2026-06-20"), []);
assert.deepStrictEqual(backlogRiskSummary(null, "2026-06-20"), []);
