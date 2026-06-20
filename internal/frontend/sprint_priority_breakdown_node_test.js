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

const { sprintReportPriorityBreakdown } = require("./static/app.js");

assert.deepStrictEqual(
  sprintReportPriorityBreakdown([
    { priority: "High" },
    { priority: "" },
    { priority: "Low" },
    { priority: "High" },
    { priority: "Medium" },
    { priority: "" },
    { priority: "Low" }
  ]),
  [
    { label: "High", count: 2 },
    { label: "Low", count: 2 },
    { label: "No priority", count: 2 },
    { label: "Medium", count: 1 }
  ]
);

assert.deepStrictEqual(sprintReportPriorityBreakdown([]), []);
assert.deepStrictEqual(sprintReportPriorityBreakdown(null), []);
