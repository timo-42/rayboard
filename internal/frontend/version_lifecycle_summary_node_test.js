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
  versionLifecycleSummary,
  versionLifecycleSummaryItems
} = require("./static/app.js");

const versions = [
  { id: "v1", state: "planned" },
  { id: "v2", state: "released" },
  { id: "v3", state: "archived" },
  { id: "v4", state: "staged" },
  { id: "v5", state: "staged" },
  { id: "v6", state: "paused" },
  { id: "v7", state: "" },
  { id: "v8" }
];

assert.deepStrictEqual(versionLifecycleSummary(versions), {
  total: 8,
  planned: 3,
  released: 1,
  archived: 1,
  unmodeled: 3,
  unmodeled_items: [
    { state: "staged", count: 2 },
    { state: "paused", count: 1 }
  ]
});

assert.deepStrictEqual(
  versionLifecycleSummaryItems(versionLifecycleSummary(versions)),
  [
    "versions: 8",
    "planned: 3",
    "released: 1",
    "archived: 1",
    "unmodeled: 3",
    "unmodeled staged: 2",
    "unmodeled paused: 1"
  ]
);

assert.deepStrictEqual(versionLifecycleSummary(null), {
  total: 0,
  planned: 0,
  released: 0,
  archived: 0,
  unmodeled: 0,
  unmodeled_items: []
});
assert.deepStrictEqual(versionLifecycleSummaryItems(null), [
  "versions: 0",
  "planned: 0",
  "released: 0",
  "archived: 0",
  "unmodeled: 0"
]);
