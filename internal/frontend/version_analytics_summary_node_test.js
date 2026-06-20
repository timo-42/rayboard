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

const { versionReportAnalyticsSummary } = require("./static/app.js");

assert.deepStrictEqual(versionReportAnalyticsSummary({
  velocity: { completed: 8, unit: "points" },
  burndown: [
    { date: "2026-06-19", remaining: 5 },
    { date: "2026-06-20", remaining: 2 }
  ],
  burnup: [
    { date: "2026-06-19", total: 10, done: 5 },
    { date: "2026-06-20", total: 10, done: 8 }
  ]
}), {
  velocity: { completed: 8, unit: "points" },
  remaining: 2,
  burnup: { done: 8, total: 10 }
});

assert.deepStrictEqual(versionReportAnalyticsSummary(null), {
  velocity: { completed: 0, unit: "tickets" },
  remaining: 0,
  burnup: { done: 0, total: 0 }
});
