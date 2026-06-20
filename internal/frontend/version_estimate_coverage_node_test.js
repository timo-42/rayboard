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

const { versionReportEstimateCoverage } = require("./static/app.js");

assert.deepStrictEqual(versionReportEstimateCoverage({ total: 5, story_points_unestimated: 2 }), {
  total: 5,
  items: [
    { label: "Estimated", count: 3 },
    { label: "Unestimated", count: 2 }
  ]
});

assert.deepStrictEqual(versionReportEstimateCoverage({ total: 3, story_points_unestimated: 9 }), {
  total: 3,
  items: [
    { label: "Estimated", count: 0 },
    { label: "Unestimated", count: 3 }
  ]
});

assert.deepStrictEqual(versionReportEstimateCoverage({}), {
  total: 0,
  items: [
    { label: "Estimated", count: 0 },
    { label: "Unestimated", count: 0 }
  ]
});

assert.deepStrictEqual(versionReportEstimateCoverage({ total: "not-a-number", story_points_unestimated: "bad" }), {
  total: 0,
  items: [
    { label: "Estimated", count: 0 },
    { label: "Unestimated", count: 0 }
  ]
});
