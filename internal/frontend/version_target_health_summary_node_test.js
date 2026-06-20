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

const { versionTargetHealthSummary } = require("./static/app.js");

assert.deepStrictEqual(
  versionTargetHealthSummary([
    { state: "released", target_date: "2026-06-01", release_date: "2026-06-01" },
    { state: "archived", target_date: "2026-05-01", release_date: "2026-05-02" },
    { state: "planned", target_date: "2026-06-19" },
    { state: "planned", target_date: "2026-06-20" },
    { state: "planned", target_date: "2026-07-03" },
    { state: "planned", target_date: "2026-07-10" },
    { state: "planned", target_date: "" }
  ], "2026-06-20"),
  [
    { key: "released", label: "Released", count: 2 },
    { key: "overdue", label: "Overdue", count: 1 },
    { key: "due-soon", label: "Due soon", count: 2 },
    { key: "scheduled", label: "Scheduled later", count: 1 },
    { key: "unscheduled", label: "Unscheduled", count: 1 }
  ]
);

assert.deepStrictEqual(
  versionTargetHealthSummary([], "2026-06-20"),
  [
    { key: "released", label: "Released", count: 0 },
    { key: "overdue", label: "Overdue", count: 0 },
    { key: "due-soon", label: "Due soon", count: 0 },
    { key: "scheduled", label: "Scheduled later", count: 0 },
    { key: "unscheduled", label: "Unscheduled", count: 0 }
  ]
);

assert.deepStrictEqual(versionTargetHealthSummary(null, "2026-06-20").map((item) => item.count), [0, 0, 0, 0, 0]);
