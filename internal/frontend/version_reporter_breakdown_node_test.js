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

const { versionReportReporterBreakdown } = require("./static/app.js");

assert.deepStrictEqual(versionReportReporterBreakdown([
  { reporter_id: "user_1", story_points: 3 },
  { reporter_id: "   ", story_points: null },
  { reporter_id: "user_2", story_points: 5 },
  { reporter_id: "user_1", story_points: "" },
  { reporter_id: "", story_points: 2 },
  { reporter_id: "user_3", story_points: "bad" }
]), [
  {
    key: "user_1",
    label: "reporter user_1",
    tickets: 2,
    story_points: 3,
    has_story_points: true
  },
  {
    key: "user_2",
    label: "reporter user_2",
    tickets: 1,
    story_points: 5,
    has_story_points: true
  },
  {
    key: "user_3",
    label: "reporter user_3",
    tickets: 1,
    story_points: 0,
    has_story_points: false
  },
  {
    key: "",
    label: "No reporter",
    tickets: 2,
    story_points: 2,
    has_story_points: true
  }
]);

assert.deepStrictEqual(versionReportReporterBreakdown([]), []);
assert.deepStrictEqual(versionReportReporterBreakdown(null), []);
assert.deepStrictEqual(versionReportReporterBreakdown({}), []);
