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

const { sprintReportRiskSummary } = require("./static/app.js");

assert.deepStrictEqual(sprintReportRiskSummary([
  { status: "todo", due_date: "2026-06-19", updated_at: "2026-06-01T08:00:00Z", assignee_id: "", start_date: "" },
  { status: "in_progress", due_date: "2026-06-22", updated_at: "2026-06-12T08:00:00Z", assignee_id: "user_1", start_date: "2026-06-20" },
  { status: "todo", due_date: "", updated_at: "2026-06-20T08:00:00Z", assignee_id: "   ", start_date: "2026-06-20" },
  { status: "DONE", due_date: "2026-06-10", updated_at: "2026-05-01T08:00:00Z", assignee_id: "", start_date: "" },
  { status: "todo", due_date: "not-a-date", updated_at: "not-a-date", assignee_id: "user_2", start_date: "" }
], "2026-06-20"), [
  { key: "overdue_open", label: "Open overdue", count: 1 },
  { key: "stale_open", label: "Stale open", count: 2 },
  { key: "unassigned_open", label: "Unassigned open", count: 2 },
  { key: "unscheduled_open", label: "Unscheduled open", count: 3 }
]);

assert.deepStrictEqual(sprintReportRiskSummary([], "2026-06-20"), []);
assert.deepStrictEqual(sprintReportRiskSummary(null, "2026-06-20"), []);
