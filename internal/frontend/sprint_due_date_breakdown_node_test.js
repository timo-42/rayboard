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

const { sprintReportDueDateBreakdown } = require("./static/app.js");

assert.deepStrictEqual(sprintReportDueDateBreakdown([
  { due_date: "2026-06-19" },
  { due_date: "2026-06-20" },
  { due_date: "2026-06-22" },
  { due_date: "2026-06-30" },
  { due_date: "" },
  { due_date: "not-a-date" }
], "2026-06-20"), [
  { key: "overdue", label: "Overdue", count: 1 },
  { key: "today", label: "Due today", count: 1 },
  { key: "soon", label: "Due soon", count: 1 },
  { key: "later", label: "Later", count: 1 },
  { key: "none", label: "No due date", count: 2 }
]);

assert.deepStrictEqual(sprintReportDueDateBreakdown([], "2026-06-20"), []);
assert.deepStrictEqual(sprintReportDueDateBreakdown(null, "2026-06-20"), []);
