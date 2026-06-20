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

const {
  componentOwnerMetadataItems,
  componentOwnershipSummary,
  componentOwnershipSummaryItems
} = require("./static/app.js");

const components = [
  { id: "component_api", owner_user_id: "alice", default_assignee_id: "bob" },
  { id: "component_web", owner_user_id: "", default_assignee_id: "carol" },
  { id: "component_ops", owner_user_id: "drew", default_assignee_id: "" },
  { id: "component_docs", owner_user_id: "", default_assignee_id: "" }
];

assert.deepStrictEqual(componentOwnershipSummary(components), {
  total: 4,
  owners: 2,
  default_assignees: 2,
  fully_covered: 1,
  missing_owner: 2,
  missing_default_assignee: 2
});

assert.deepStrictEqual(
  componentOwnershipSummaryItems(componentOwnershipSummary(components)),
  [
    "components: 4",
    "owners: 2",
    "default assignees: 2",
    "fully covered: 1",
    "missing owner: 2",
    "missing default: 2"
  ]
);

assert.deepStrictEqual(componentOwnerMetadataItems(components[0]), [
  "owner: alice",
  "default assignee: bob"
]);
assert.deepStrictEqual(componentOwnerMetadataItems(components[3]), [
  "owner: none",
  "default assignee: none"
]);
assert.deepStrictEqual(componentOwnershipSummary(null), {
  total: 0,
  owners: 0,
  default_assignees: 0,
  fully_covered: 0,
  missing_owner: 0,
  missing_default_assignee: 0
});
assert.deepStrictEqual(componentOwnershipSummaryItems(null), [
  "components: 0",
  "owners: 0",
  "default assignees: 0",
  "fully covered: 0",
  "missing owner: 0",
  "missing default: 0"
]);
