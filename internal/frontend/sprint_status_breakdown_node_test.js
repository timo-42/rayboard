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

const { sprintReportStatusBreakdown } = require("./static/app.js");

assert.deepStrictEqual(sprintReportStatusBreakdown({
  by_status: {
    done: 2,
    review: 1,
    todo: 3,
    "": 1,
    blocked: 0
  }
}), [
  { label: "Todo", count: 3 },
  { label: "Done", count: 2 },
  { label: "No status", count: 1 },
  { label: "review", count: 1 }
]);

assert.deepStrictEqual(sprintReportStatusBreakdown({ by_status: {} }), []);
assert.deepStrictEqual(sprintReportStatusBreakdown({}), []);
assert.deepStrictEqual(sprintReportStatusBreakdown(null), []);
