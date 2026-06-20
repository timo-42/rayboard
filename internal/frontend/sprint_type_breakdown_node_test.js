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

const { sprintReportTypeBreakdown } = require("./static/app.js");

assert.deepStrictEqual(
  sprintReportTypeBreakdown([
    { type: "bug" },
    { type: "story" },
    { type: "" },
    { type: "task" },
    { type: "bug" },
    { type: "story" },
    { type: "" }
  ]),
  [
    { label: "bug", count: 2 },
    { label: "No issue type", count: 2 },
    { label: "story", count: 2 },
    { label: "task", count: 1 }
  ]
);

assert.deepStrictEqual(sprintReportTypeBreakdown([]), []);
assert.deepStrictEqual(sprintReportTypeBreakdown(null), []);
