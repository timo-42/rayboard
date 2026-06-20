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

const { ticketLinkDependencySummary } = require("./static/app.js");

assert.deepStrictEqual(
  ticketLinkDependencySummary([
    { link_type: "blocks" },
    { link_type: "is_blocked_by" },
    { link_type: "blocks" },
    { link_type: "relates_to" },
    { link_type: "unknown" }
  ]),
  {
    total: 4,
    items: [
      { label: "Blocks", count: 2 },
      { label: "Blocked by", count: 1 },
      { label: "Related", count: 1 }
    ]
  }
);

assert.deepStrictEqual(ticketLinkDependencySummary([]), {
  total: 0,
  items: [
    { label: "Blocks", count: 0 },
    { label: "Blocked by", count: 0 },
    { label: "Related", count: 0 }
  ]
});

assert.deepStrictEqual(ticketLinkDependencySummary(null), {
  total: 0,
  items: [
    { label: "Blocks", count: 0 },
    { label: "Blocked by", count: 0 },
    { label: "Related", count: 0 }
  ]
});
