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

const { versionReportTimelineItems } = require("./static/app.js");

assert.deepStrictEqual(
  versionReportTimelineItems({ state: "released", target_date: "2026-06-20", release_date: "" }),
  ["target 2026-06-20", "release date missing", "state released"]
);

assert.deepStrictEqual(
  versionReportTimelineItems({ state: "archived", target_date: "", release_date: "" }),
  ["no target date", "release date missing", "state archived"]
);

assert.deepStrictEqual(
  versionReportTimelineItems({ state: "released", target_date: "2026-06-20", release_date: "2026-06-20" }),
  ["target 2026-06-20", "release 2026-06-20", "released on target", "state released"]
);

assert.deepStrictEqual(
  versionReportTimelineItems({ state: "released", target_date: "2026-06-20", release_date: "2026-06-18" }),
  ["target 2026-06-20", "release 2026-06-18", "2 days early", "state released"]
);

assert.deepStrictEqual(
  versionReportTimelineItems({ state: "released", target_date: "2026-06-20", release_date: "2026-06-23" }),
  ["target 2026-06-20", "release 2026-06-23", "3 days late", "state released"]
);
