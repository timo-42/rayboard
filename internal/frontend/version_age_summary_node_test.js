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
  versionAgeSummary,
  versionAgeSummaryItems
} = require("./static/app.js");

const tickets = [
  { id: "a", version_id: "version_2026_7", created_at: "2026-06-20T10:00:00Z" },
  { id: "b", version_id: "version_2026_7", created_at: "2026-06-10T00:00:00Z" },
  { id: "c", version_id: "version_2026_7", created_at: "2026-05-10T00:00:00Z" },
  { id: "d", version_id: "version_2026_7", created_at: "2026-03-01T00:00:00Z" },
  { id: "e", version_id: "version_2026_7", created_at: "" },
  { id: "f", version_id: "version_2026_7", created_at: "not-a-date" },
  { id: "g", version_id: "version_2026_8", created_at: "2026-03-01T00:00:00Z" },
  { id: "h", version_id: "", created_at: "2026-03-01T00:00:00Z" }
];

assert.deepStrictEqual(versionAgeSummary(tickets, "version_2026_7", "2026-06-20"), {
  total: 6,
  new: 1,
  recent: 1,
  aging: 1,
  old: 1,
  unknown_age: 2
});

assert.deepStrictEqual(
  versionAgeSummaryItems(versionAgeSummary(tickets, "version_2026_7", "2026-06-20")),
  [
    "total: 6",
    "new: 1",
    "recent: 1",
    "aging: 1",
    "old: 1",
    "unknown age: 2"
  ]
);

assert.deepStrictEqual(versionAgeSummary(tickets, "version_missing", "2026-06-20"), {
  total: 0,
  new: 0,
  recent: 0,
  aging: 0,
  old: 0,
  unknown_age: 0
});
assert.deepStrictEqual(versionAgeSummary(null, "version_2026_7", "2026-06-20"), {
  total: 0,
  new: 0,
  recent: 0,
  aging: 0,
  old: 0,
  unknown_age: 0
});
assert.deepStrictEqual(versionAgeSummaryItems(null), [
  "total: 0",
  "new: 0",
  "recent: 0",
  "aging: 0",
  "old: 0",
  "unknown age: 0"
]);
