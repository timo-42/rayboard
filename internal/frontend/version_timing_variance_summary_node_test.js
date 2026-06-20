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

const { versionTimingVarianceSummary } = require("./static/app.js");

assert.deepStrictEqual(
  versionTimingVarianceSummary([
    { state: "released", target_date: "2026-06-20", release_date: "2026-06-18" },
    { state: "released", target_date: "2026-06-20", release_date: "2026-06-20" },
    { state: "released", target_date: "2026-06-20", release_date: "2026-06-23" },
    { state: "released", target_date: "", release_date: "2026-06-23" },
    { state: "archived", target_date: "2026-06-20", release_date: "" },
    { state: "planned", target_date: "2026-07-10", release_date: "" }
  ]),
  [
    { key: "early", label: "Released early", count: 1 },
    { key: "on-target", label: "Released on target", count: 1 },
    { key: "late", label: "Released late", count: 1 },
    { key: "no-target", label: "Released without target date", count: 1 },
    { key: "no-release-date", label: "Released without release date", count: 1 },
    { key: "not-released", label: "Not released", count: 1 }
  ]
);

assert.deepStrictEqual(
  versionTimingVarianceSummary([]),
  [
    { key: "early", label: "Released early", count: 0 },
    { key: "on-target", label: "Released on target", count: 0 },
    { key: "late", label: "Released late", count: 0 },
    { key: "no-target", label: "Released without target date", count: 0 },
    { key: "no-release-date", label: "Released without release date", count: 0 },
    { key: "not-released", label: "Not released", count: 0 }
  ]
);

assert.deepStrictEqual(versionTimingVarianceSummary(null).map((item) => item.count), [0, 0, 0, 0, 0, 0]);
