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
  customFieldLayoutSummary,
  customFieldRequirementInsightItems,
  customFieldRequirementInsights
} = require("./static/app.js");

const fields = [
  { key: "severity", field_type: "single_select", required: true, options: ["Low", "High"] },
  { key: "impact", field_type: "number", required: false },
  { key: "region", field_type: "multi_select", required: true, options: ["", "EU"] },
  { key: "channel", field_type: "single_select", required: false, options: [] },
  { key: "customer", field_type: "text", required: false },
  { key: "", field_type: "boolean", required: false }
];

assert.deepStrictEqual(customFieldRequirementInsights(fields), {
  required: 2,
  optional: 4,
  select_with_options: 2,
  select_without_options: 1,
  search_ready: 4,
  configuration_attention: 1
});

assert.deepStrictEqual(customFieldLayoutSummary(fields), {
  total: 6,
  required: 2,
  optional: 4,
  types: [
    { type: "boolean", count: 1 },
    { type: "multi_select", count: 1 },
    { type: "number", count: 1 },
    { type: "single_select", count: 2 },
    { type: "text", count: 1 }
  ],
  insights: {
    required: 2,
    optional: 4,
    select_with_options: 2,
    select_without_options: 1,
    search_ready: 4,
    configuration_attention: 1
  }
});

assert.deepStrictEqual(customFieldRequirementInsightItems(customFieldRequirementInsights(fields)), [
  "required: 2",
  "optional: 4",
  "selects configured: 2",
  "selects missing options: 1",
  "search-ready: 4",
  "needs attention: 1"
]);

assert.deepStrictEqual(customFieldRequirementInsights(null), {
  required: 0,
  optional: 0,
  select_with_options: 0,
  select_without_options: 0,
  search_ready: 0,
  configuration_attention: 0
});

assert.deepStrictEqual(customFieldRequirementInsightItems(null), [
  "required: 0",
  "optional: 0",
  "selects configured: 0",
  "selects missing options: 0",
  "search-ready: 0",
  "needs attention: 0"
]);
