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
  versionStartDateSummary,
  versionStartDateSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", start_date: "2026-06-19", status: "todo" },
  { id: "b", version_id: "version_2026_7", start_date: "2026-06-20", status: "todo" },
  { id: "c", version_id: "version_2026_7", start_date: "2026-06-21", status: "done" },
  { id: "d", version_id: "version_2026_7", start_date: "", status: "todo" },
  { id: "e", version_id: "version_2026_7", start_date: "2026-06-18", status: "done" },
  { id: "f", version_id: "version_2026_8", start_date: "2026-06-19", status: "todo" },
  { id: "g", version_id: "", start_date: "2026-06-19", status: "todo" }
];

assert.deepStrictEqual(versionStartDateSummary(tickets, "version_2026_7", "2026-06-20"), {
  total: 5,
  with_start_date: 4,
  missing_start_date: 1,
  started_open: 1,
  starts_today: 1,
  future_start: 1
});

assert.deepStrictEqual(
  versionStartDateSummaryItems(versionStartDateSummary(tickets, "version_2026_7", "2026-06-20")),
  [
    "total: 5",
    "with start date: 4",
    "missing start date: 1",
    "started open: 1",
    "starts today: 1",
    "future start: 1"
  ]
);

assert.deepStrictEqual(versionStartDateSummary(tickets, "version_missing", "2026-06-20"), {
  total: 0,
  with_start_date: 0,
  missing_start_date: 0,
  started_open: 0,
  starts_today: 0,
  future_start: 0
});
assert.deepStrictEqual(versionStartDateSummary(null, "version_2026_7", "2026-06-20"), {
  total: 0,
  with_start_date: 0,
  missing_start_date: 0,
  started_open: 0,
  starts_today: 0,
  future_start: 0
});
assert.deepStrictEqual(versionStartDateSummaryItems(null), [
  "total: 0",
  "with start date: 0",
  "missing start date: 0",
  "started open: 0",
  "starts today: 0",
  "future start: 0"
]);
