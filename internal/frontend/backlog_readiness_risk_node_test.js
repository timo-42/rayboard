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
  backlogAgeBreakdown,
  backlogAttentionSummary,
  backlogComponentBreakdown,
  backlogComponentBreakdownLabel,
  backlogDueDateBreakdown,
  backlogLabelBreakdown,
  backlogLabelBreakdownVisibleItems,
  backlogReadinessSummary,
  backlogRiskSummary,
  backlogStartDateBreakdown,
  backlogUpdateFreshness
} = require("./static/app.js");

const tickets = [
  {
    status: "todo",
    priority: "High",
    assignee_id: "",
    component_id: "api",
    story_points: null,
    labels: ["backend", "urgent"],
    created_at: "2026-06-18T09:00:00Z",
    start_date: "",
    due_date: "2026-06-19",
    updated_at: "2026-06-01T09:00:00Z"
  },
  {
    status: "blocked",
    priority: "Medium",
    assignee_id: "user_1",
    component_id: "api",
    story_points: 3,
    labels: ["backend"],
    created_at: "2026-05-25T09:00:00Z",
    start_date: "2026-06-18",
    due_date: "2026-06-25",
    updated_at: "2026-06-20T09:00:00Z"
  },
  {
    status: "done",
    priority: "Critical",
    assignee_id: "",
    component_id: "ui",
    story_points: "",
    labels: [],
    created_at: "2026-04-01T09:00:00Z",
    start_date: "",
    due_date: "2026-06-01",
    updated_at: "2026-05-01T09:00:00Z"
  },
  {
    status: "todo",
    priority: "Low",
    assignee_id: "   ",
    component_id: "",
    story_points: null,
    labels: null,
    created_at: "",
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

assert.deepStrictEqual(backlogAttentionSummary(tickets, "2026-06-20"), [
  { key: "blocked_open", label: "Blocked open", count: 1 },
  { key: "high_priority_open", label: "High-priority open", count: 1 },
  { key: "unestimated_high", label: "Unestimated high priority", count: 1 },
  { key: "stale_high", label: "Stale high priority", count: 1 }
]);

assert.deepStrictEqual(backlogAgeBreakdown(tickets, "2026-06-20"), [
  { key: "new", label: "New (0-7 days)", count: 1 },
  { key: "recent", label: "Recent (8-30 days)", count: 1 },
  { key: "older", label: "Older (31+ days)", count: 1 },
  { key: "none", label: "No created date", count: 1 }
]);

assert.deepStrictEqual(backlogUpdateFreshness(tickets, "2026-06-20"), [
  { key: "today", label: "Updated today", count: 1 },
  { key: "stale", label: "Stale (8-30 days)", count: 2 },
  { key: "dormant", label: "Dormant (31+ days)", count: 1 }
]);

const scheduleTickets = [
  { start_date: "2026-06-01", due_date: "2026-06-19" },
  { start_date: "2026-06-20", due_date: "2026-06-20" },
  { start_date: "2026-06-22", due_date: "2026-06-22" },
  { start_date: "2026-07-01", due_date: "2026-07-01" },
  { start_date: "", due_date: "" },
  { start_date: "bad", due_date: "bad" }
];

assert.deepStrictEqual(backlogStartDateBreakdown(scheduleTickets, "2026-06-20"), [
  { key: "started", label: "Started", count: 1 },
  { key: "today", label: "Starts today", count: 1 },
  { key: "soon", label: "Starts soon", count: 1 },
  { key: "future", label: "Future start", count: 1 },
  { key: "none", label: "No start date", count: 2 }
]);

assert.deepStrictEqual(backlogDueDateBreakdown(scheduleTickets, "2026-06-20"), [
  { key: "overdue", label: "Overdue", count: 1 },
  { key: "today", label: "Due today", count: 1 },
  { key: "soon", label: "Due soon", count: 1 },
  { key: "later", label: "Later", count: 1 },
  { key: "none", label: "No due date", count: 2 }
]);

assert.deepStrictEqual(backlogLabelBreakdown(tickets), [
  { label: "backend", count: 2 },
  { label: "urgent", count: 1 },
  { label: "No labels", count: 2 }
]);

assert.deepStrictEqual(backlogComponentBreakdown(tickets), [
  {
    id: "api",
    label: "api",
    count: 2,
    done: 0,
    story_points_total: 3,
    story_points_done: 0,
    unestimated: 1
  },
  {
    id: "ui",
    label: "ui",
    count: 1,
    done: 1,
    story_points_total: 0,
    story_points_done: 0,
    unestimated: 1
  },
  {
    id: "",
    label: "No component",
    count: 1,
    done: 0,
    story_points_total: 0,
    story_points_done: 0,
    unestimated: 1
  }
]);
assert.strictEqual(
  backlogComponentBreakdownLabel(backlogComponentBreakdown(tickets)[0]),
  "api: 0/2 done / 0/3 pts"
);
assert.strictEqual(
  backlogComponentBreakdownLabel(backlogComponentBreakdown(tickets)[2]),
  "No component: 0/1 done / 1 unestimated"
);

assert.deepStrictEqual(backlogLabelBreakdownVisibleItems([
  { label: "label-1", count: 9 },
  { label: "label-2", count: 8 },
  { label: "label-3", count: 7 },
  { label: "label-4", count: 6 },
  { label: "label-5", count: 5 },
  { label: "label-6", count: 4 },
  { label: "label-7", count: 3 },
  { label: "label-8", count: 2 },
  { label: "label-9", count: 1 },
  { label: "No labels", count: 4 }
]), [
  { label: "label-1", count: 9 },
  { label: "label-2", count: 8 },
  { label: "label-3", count: 7 },
  { label: "label-4", count: 6 },
  { label: "label-5", count: 5 },
  { label: "label-6", count: 4 },
  { label: "label-7", count: 3 },
  { label: "No labels", count: 4 }
]);

assert.deepStrictEqual(backlogReadinessSummary([]), []);
assert.deepStrictEqual(backlogReadinessSummary(null), []);
assert.deepStrictEqual(backlogRiskSummary([], "2026-06-20"), []);
assert.deepStrictEqual(backlogRiskSummary(null, "2026-06-20"), []);
assert.deepStrictEqual(backlogAttentionSummary([], "2026-06-20"), []);
assert.deepStrictEqual(backlogAttentionSummary(null, "2026-06-20"), []);
assert.deepStrictEqual(backlogStartDateBreakdown([], "2026-06-20"), []);
assert.deepStrictEqual(backlogStartDateBreakdown(null, "2026-06-20"), []);
assert.deepStrictEqual(backlogDueDateBreakdown([], "2026-06-20"), []);
assert.deepStrictEqual(backlogDueDateBreakdown(null, "2026-06-20"), []);
assert.deepStrictEqual(backlogAgeBreakdown([], "2026-06-20"), []);
assert.deepStrictEqual(backlogAgeBreakdown(null, "2026-06-20"), []);
assert.deepStrictEqual(backlogUpdateFreshness([], "2026-06-20"), []);
assert.deepStrictEqual(backlogUpdateFreshness(null, "2026-06-20"), []);
assert.deepStrictEqual(backlogLabelBreakdown([]), []);
assert.deepStrictEqual(backlogLabelBreakdown(null), []);
assert.deepStrictEqual(backlogLabelBreakdownVisibleItems(null), []);
assert.deepStrictEqual(backlogComponentBreakdown([]), []);
assert.deepStrictEqual(backlogComponentBreakdown(null), []);
