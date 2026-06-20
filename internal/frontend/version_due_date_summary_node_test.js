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
  versionDueDateSummary,
  versionDueDateSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", due_date: "2026-06-19", status: "todo" },
  { id: "b", version_id: "version_2026_7", due_date: "2026-06-20", status: "todo" },
  { id: "c", version_id: "version_2026_7", due_date: "2026-06-21", status: "done" },
  { id: "d", version_id: "version_2026_7", due_date: "", status: "todo" },
  { id: "e", version_id: "version_2026_7", due_date: "2026-06-18", status: "done" },
  { id: "f", version_id: "version_2026_8", due_date: "2026-06-19", status: "todo" },
  { id: "g", version_id: "", due_date: "2026-06-19", status: "todo" }
];

assert.deepStrictEqual(versionDueDateSummary(tickets, "version_2026_7", "2026-06-20"), {
  total: 5,
  with_due_date: 4,
  missing_due_date: 1,
  open_overdue: 1,
  due_today: 1,
  future_due: 1
});

assert.deepStrictEqual(
  versionDueDateSummaryItems(versionDueDateSummary(tickets, "version_2026_7", "2026-06-20")),
  [
    "total: 5",
    "with due date: 4",
    "missing due date: 1",
    "open overdue: 1",
    "due today: 1",
    "future due: 1"
  ]
);

assert.deepStrictEqual(versionDueDateSummary(tickets, "version_missing", "2026-06-20"), {
  total: 0,
  with_due_date: 0,
  missing_due_date: 0,
  open_overdue: 0,
  due_today: 0,
  future_due: 0
});
assert.deepStrictEqual(versionDueDateSummary(null, "version_2026_7", "2026-06-20"), {
  total: 0,
  with_due_date: 0,
  missing_due_date: 0,
  open_overdue: 0,
  due_today: 0,
  future_due: 0
});
assert.deepStrictEqual(versionDueDateSummaryItems(null), [
  "total: 0",
  "with due date: 0",
  "missing due date: 0",
  "open overdue: 0",
  "due today: 0",
  "future due: 0"
]);
