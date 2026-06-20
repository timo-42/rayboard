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

const { versionReportSprints } = require("./static/app.js");

assert.deepStrictEqual(versionReportSprints([
  { sprint_id: "sprint_1", status: "done", story_points: 3 },
  { sprint_id: "", status: "todo", story_points: null },
  { sprint_id: "sprint_2", status: "todo", story_points: 5 },
  { sprint_id: "sprint_1", status: "in_progress", story_points: "" },
  { sprint_id: "", status: "done", story_points: 2 }
]), [
  {
    id: "sprint_1",
    name: "sprint_1",
    total: 2,
    done: 1,
    story_points_total: 3,
    story_points_done: 3,
    unestimated: 1
  },
  {
    id: "sprint_2",
    name: "sprint_2",
    total: 1,
    done: 0,
    story_points_total: 5,
    story_points_done: 0,
    unestimated: 0
  },
  {
    id: "",
    name: "No sprint",
    total: 2,
    done: 1,
    story_points_total: 2,
    story_points_done: 2,
    unestimated: 1
  }
]);

assert.deepStrictEqual(versionReportSprints([]), []);
assert.deepStrictEqual(versionReportSprints(null), []);
assert.deepStrictEqual(versionReportSprints({}), []);
