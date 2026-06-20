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
  versionDateCoverageSummary,
  versionDateCoverageSummaryItems
} = require("./static/app.js");

const versions = [
  { id: "v1", target_date: "2026-07-01", release_date: "2026-07-02" },
  { id: "v2", target_date: "2026-08-01", release_date: "" },
  { id: "v3", target_date: "", release_date: "2026-09-01" },
  { id: "v4", target_date: "", release_date: "" },
  { id: "v5" }
];

assert.deepStrictEqual(versionDateCoverageSummary(versions), {
  total: 5,
  target_dates: 2,
  release_dates: 2,
  both_dates: 1,
  target_only: 1,
  release_only: 1,
  missing_dates: 2
});

assert.deepStrictEqual(
  versionDateCoverageSummaryItems(versionDateCoverageSummary(versions)),
  [
    "versions: 5",
    "target dates: 2",
    "release dates: 2",
    "both dates: 1",
    "target only: 1",
    "release only: 1",
    "missing dates: 2"
  ]
);

assert.deepStrictEqual(versionDateCoverageSummary(null), {
  total: 0,
  target_dates: 0,
  release_dates: 0,
  both_dates: 0,
  target_only: 0,
  release_only: 0,
  missing_dates: 0
});
assert.deepStrictEqual(versionDateCoverageSummaryItems(null), [
  "versions: 0",
  "target dates: 0",
  "release dates: 0",
  "both dates: 0",
  "target only: 0",
  "release only: 0",
  "missing dates: 0"
]);
