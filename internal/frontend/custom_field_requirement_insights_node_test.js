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
  customFieldOptionUsageSummary,
  customFieldRequirementInsightItems,
  customFieldRequirementInsights,
  customFieldUsageSummary,
  customFieldValuePresent
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

assert.strictEqual(customFieldValuePresent(" High "), true);
assert.strictEqual(customFieldValuePresent("   "), false);
assert.strictEqual(customFieldValuePresent(["", "EU"]), true);
assert.strictEqual(customFieldValuePresent([]), false);
assert.strictEqual(customFieldValuePresent(false), true);
assert.strictEqual(customFieldValuePresent(null), false);

assert.deepStrictEqual(customFieldUsageSummary(fields, [
  {
    custom_fields: {
      severity: "High",
      impact: 3,
      region: ["EU", ""],
      customer: "Acme"
    }
  },
  {
    custom_fields: {
      severity: "",
      impact: null,
      region: [],
      channel: "Email",
      customer: "   "
    }
  },
  {
    custom_fields: {
      severity: "Low",
      impact: 0,
      region: ["EU", "APAC"],
      customer: "Northwind"
    }
  }
]), [
  {
    key: "severity",
    label: "severity",
    field_type: "single_select",
    required: true,
    tickets: 3,
    populated: 2,
    empty: 1,
    required_missing: 1,
    option_usage: [
      { option: "High", count: 1 },
      { option: "Low", count: 1 }
    ]
  },
  {
    key: "impact",
    label: "impact",
    field_type: "number",
    required: false,
    tickets: 3,
    populated: 2,
    empty: 1,
    required_missing: 0,
    option_usage: []
  },
  {
    key: "region",
    label: "region",
    field_type: "multi_select",
    required: true,
    tickets: 3,
    populated: 2,
    empty: 1,
    required_missing: 1,
    option_usage: [
      { option: "EU", count: 2 },
      { option: "APAC", count: 1 }
    ]
  },
  {
    key: "channel",
    label: "channel",
    field_type: "single_select",
    required: false,
    tickets: 3,
    populated: 1,
    empty: 2,
    required_missing: 0,
    option_usage: [
      { option: "Email", count: 1 }
    ]
  },
  {
    key: "customer",
    label: "customer",
    field_type: "text",
    required: false,
    tickets: 3,
    populated: 2,
    empty: 1,
    required_missing: 0,
    option_usage: []
  }
]);

assert.deepStrictEqual(customFieldUsageSummary(null, null), []);

assert.deepStrictEqual(customFieldOptionUsageSummary(fields, [
  {
    custom_fields: {
      severity: "High",
      region: ["EU", ""]
    }
  },
  {
    custom_fields: {
      severity: "",
      channel: "Email"
    }
  },
  {
    custom_fields: {
      severity: "Low",
      region: ["EU", "APAC"]
    }
  }
]), [
  {
    key: "severity",
    label: "severity",
    field_type: "single_select",
    configured_options: [
      { option: "Low", count: 1 },
      { option: "High", count: 1 }
    ],
    unconfigured_options: []
  },
  {
    key: "region",
    label: "region",
    field_type: "multi_select",
    configured_options: [
      { option: "EU", count: 2 }
    ],
    unconfigured_options: [
      { option: "APAC", count: 1 }
    ]
  },
  {
    key: "channel",
    label: "channel",
    field_type: "single_select",
    configured_options: [],
    unconfigured_options: [
      { option: "Email", count: 1 }
    ]
  }
]);

assert.deepStrictEqual(customFieldOptionUsageSummary([], []), []);
assert.deepStrictEqual(customFieldOptionUsageSummary(null, null), []);
