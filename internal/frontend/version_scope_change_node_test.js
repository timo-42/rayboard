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

const { versionReportScopeChangeItems } = require("./static/app.js");

assert.deepStrictEqual(
  versionReportScopeChangeItems({
    current: 7,
    snapshot: 5,
    added: 3,
    removed: 1,
    unchanged: 4
  }),
  [
    "current 7",
    "snapshot 5",
    "added 3",
    "removed 1",
    "unchanged 4"
  ]
);

assert.deepStrictEqual(
  versionReportScopeChangeItems({ current: "2", added: null }),
  [
    "current 2",
    "snapshot 0",
    "added 0",
    "removed 0",
    "unchanged 0"
  ]
);

assert.deepStrictEqual(versionReportScopeChangeItems(null), [
  "current 0",
  "snapshot 0",
  "added 0",
  "removed 0",
  "unchanged 0"
]);
