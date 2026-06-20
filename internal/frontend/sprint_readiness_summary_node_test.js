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

const { sprintReportReadinessSummary } = require("./static/app.js");

assert.deepStrictEqual(sprintReportReadinessSummary([
  { assignee_id: "user_1", story_points: 0, start_date: "2026-06-20", due_date: "2026-06-30" },
  { assignee_id: "", story_points: null, start_date: "", due_date: "" },
  { assignee_id: "   ", story_points: "   ", start_date: "not-a-date", due_date: "also-bad" },
  { assignee_id: "user_2", story_points: "bad", start_date: "2026-06-21", due_date: "" }
]), [
  { key: "ready", label: "Ready", count: 1 },
  { key: "missing_assignee", label: "Missing assignee", count: 2 },
  { key: "missing_estimate", label: "Missing estimate", count: 3 },
  { key: "missing_start", label: "Missing start date", count: 2 },
  { key: "missing_due", label: "Missing due date", count: 3 }
]);

assert.deepStrictEqual(sprintReportReadinessSummary([], "2026-06-20"), []);
assert.deepStrictEqual(sprintReportReadinessSummary(null, "2026-06-20"), []);
