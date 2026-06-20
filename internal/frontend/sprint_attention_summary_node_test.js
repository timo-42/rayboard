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

const { sprintReportAttentionSummary } = require("./static/app.js");

assert.deepStrictEqual(sprintReportAttentionSummary([
  { status: "blocked", priority: "high", story_points: "", updated_at: "2026-06-01T08:00:00Z" },
  { status: "todo", priority: "critical", story_points: 0, updated_at: "2026-06-20T08:00:00Z" },
  { status: "waiting", priority: "medium", story_points: null, updated_at: "2026-06-10T08:00:00Z" },
  { status: "DONE", priority: "urgent", story_points: "", updated_at: "2026-05-01T08:00:00Z" },
  { status: "in_progress", priority: "Urgent", story_points: "   ", updated_at: "2026-06-11T08:00:00Z" },
  { status: "todo", priority: "low", story_points: "", updated_at: "2026-05-01T08:00:00Z" }
], "2026-06-20"), [
  { key: "blocked_open", label: "Blocked open", count: 2 },
  { key: "high_priority_open", label: "High-priority open", count: 3 },
  { key: "unestimated_high", label: "Unestimated high priority", count: 2 },
  { key: "stale_high", label: "Stale high priority", count: 2 }
]);

assert.deepStrictEqual(sprintReportAttentionSummary([], "2026-06-20"), []);
assert.deepStrictEqual(sprintReportAttentionSummary(null, "2026-06-20"), []);
