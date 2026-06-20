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

const { versionReportComponents } = require("./static/app.js");

assert.deepStrictEqual(versionReportComponents([
  { component_id: "component_api", status: "done", story_points: 3 },
  { component_id: "", status: "todo", story_points: null },
  { component_id: "component_web", status: "todo", story_points: 5 },
  { component_id: "component_api", status: "in_progress", story_points: "" },
  { component_id: "", status: "done", story_points: 2 }
]), [
  {
    id: "component_api",
    name: "component_api",
    total: 2,
    done: 1,
    story_points_total: 3,
    story_points_done: 3,
    unestimated: 1
  },
  {
    id: "component_web",
    name: "component_web",
    total: 1,
    done: 0,
    story_points_total: 5,
    story_points_done: 0,
    unestimated: 0
  },
  {
    id: "",
    name: "No component",
    total: 2,
    done: 1,
    story_points_total: 2,
    story_points_done: 2,
    unestimated: 1
  }
]);

assert.deepStrictEqual(versionReportComponents([]), []);
assert.deepStrictEqual(versionReportComponents(null), []);
assert.deepStrictEqual(versionReportComponents({}), []);
