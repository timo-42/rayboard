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

const { sprintReportVersions } = require("./static/app.js");

assert.deepStrictEqual(sprintReportVersions([
  { version_id: "version_1", status: "done", story_points: 3 },
  { version_id: "", status: "todo", story_points: null },
  { version_id: "version_2", status: "todo", story_points: 5 },
  { version_id: "version_1", status: "in_progress", story_points: "" },
  { version_id: "", status: "done", story_points: 2 }
]), [
  {
    id: "version_1",
    name: "version_1",
    total: 2,
    done: 1,
    story_points_total: 3,
    story_points_done: 3,
    unestimated: 1
  },
  {
    id: "version_2",
    name: "version_2",
    total: 1,
    done: 0,
    story_points_total: 5,
    story_points_done: 0,
    unestimated: 0
  },
  {
    id: "",
    name: "No version",
    total: 2,
    done: 1,
    story_points_total: 2,
    story_points_done: 2,
    unestimated: 1
  }
]);

assert.deepStrictEqual(sprintReportVersions([]), []);
assert.deepStrictEqual(sprintReportVersions(null), []);
