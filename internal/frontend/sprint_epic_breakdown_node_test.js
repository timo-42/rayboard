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

const { sprintReportEpics } = require("./static/app.js");

assert.deepStrictEqual(sprintReportEpics([
  { parent_ticket_id: "epic_1", status: "done", story_points: 3 },
  { parent_ticket_id: "", status: "todo", story_points: null },
  { parent_ticket_id: "epic_2", status: "todo", story_points: 5 },
  { parent_ticket_id: "epic_1", status: "in_progress", story_points: "" },
  { parent_ticket_id: "", status: "done", story_points: 2 }
]), [
  {
    id: "epic_1",
    name: "epic_1",
    total: 2,
    done: 1,
    story_points_total: 3,
    story_points_done: 3,
    unestimated: 1
  },
  {
    id: "epic_2",
    name: "epic_2",
    total: 1,
    done: 0,
    story_points_total: 5,
    story_points_done: 0,
    unestimated: 0
  },
  {
    id: "",
    name: "No epic",
    total: 2,
    done: 1,
    story_points_total: 2,
    story_points_done: 2,
    unestimated: 1
  }
]);

assert.deepStrictEqual(sprintReportEpics([]), []);
assert.deepStrictEqual(sprintReportEpics(null), []);
