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

const { sprintReportEstimateCoverage } = require("./static/app.js");

assert.deepStrictEqual(sprintReportEstimateCoverage({ total: 6, story_points_unestimated: 2 }), {
  total: 6,
  items: [
    { label: "Estimated", count: 4 },
    { label: "Unestimated", count: 2 }
  ]
});

assert.deepStrictEqual(sprintReportEstimateCoverage({ total: 3, story_points_unestimated: 9 }), {
  total: 3,
  items: [
    { label: "Estimated", count: 0 },
    { label: "Unestimated", count: 3 }
  ]
});

assert.deepStrictEqual(sprintReportEstimateCoverage({}), {
  total: 0,
  items: [
    { label: "Estimated", count: 0 },
    { label: "Unestimated", count: 0 }
  ]
});

assert.deepStrictEqual(sprintReportEstimateCoverage({ total: "bad", story_points_unestimated: "also-bad" }), {
  total: 0,
  items: [
    { label: "Estimated", count: 0 },
    { label: "Unestimated", count: 0 }
  ]
});
