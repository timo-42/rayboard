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

const { versionReportPriorityBreakdown } = require("./static/app.js");

assert.deepStrictEqual(
  versionReportPriorityBreakdown([
    { priority: "high" },
    { priority: "" },
    { priority: "low" },
    { priority: "high" },
    { priority: "low" },
    { priority: "medium" },
    { priority: "" }
  ]),
  [
    { label: "high", count: 2 },
    { label: "low", count: 2 },
    { label: "No priority", count: 2 },
    { label: "medium", count: 1 }
  ]
);
