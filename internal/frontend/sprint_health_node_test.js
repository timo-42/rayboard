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

const { sprintReportHealth } = require("./static/app.js");

assert.deepStrictEqual(
  sprintReportHealth({ state: "active", end_date: "2026-06-19" }, "2026-06-19"),
  { state: "due-soon", label: "Due soon", detail: "Sprint ends today" }
);

assert.deepStrictEqual(
  sprintReportHealth({ state: "active", end_date: "2026-06-22" }, "2026-06-19"),
  { state: "due-soon", label: "Due soon", detail: "3 days remaining" }
);

assert.deepStrictEqual(
  sprintReportHealth({ state: "active", end_date: "2026-06-18" }, "2026-06-19"),
  { state: "overdue", label: "Overdue", detail: "1 day past end date" }
);

assert.deepStrictEqual(
  sprintReportHealth({ state: "planned", start_date: "2026-06-21", end_date: "2026-06-30" }, "2026-06-19"),
  { state: "scheduled", label: "Scheduled", detail: "2 days to start" }
);
