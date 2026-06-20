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

const { versionReportAssigneeWorkloads } = require("./static/app.js");

assert.deepStrictEqual(
  versionReportAssigneeWorkloads([
    { assignee_id: "u2", status: "todo", story_points: "" },
    { assignee_id: "u1", status: "done", story_points: 3 },
    { assignee_id: "", status: "todo", story_points: null },
    { assignee_id: "u1", status: "todo", story_points: 2 },
    { assignee_id: "u2", status: "done", story_points: 5 },
    { assignee_id: "u2", status: "todo", story_points: "bad" }
  ]),
  [
    {
      key: "u2",
      label: "assignee u2",
      total: 3,
      done: 1,
      story_points_total: 5,
      story_points_done: 5,
      unestimated: 2
    },
    {
      key: "u1",
      label: "assignee u1",
      total: 2,
      done: 1,
      story_points_total: 5,
      story_points_done: 3,
      unestimated: 0
    },
    {
      key: "",
      label: "Unassigned",
      total: 1,
      done: 0,
      story_points_total: 0,
      story_points_done: 0,
      unestimated: 1
    }
  ]
);
