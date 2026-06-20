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
  createDocumentFragment: () => ({ append: noop }),
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
  notificationHookPreviewSummary,
  notificationHookPreviewSummaryItems
} = require("./static/app.js");

const display = {
  state: "completed",
  suppressed: true,
  plan: {
    destination_ids: ["dest_1"],
    message: "Updated",
    payload: { ticket_id: "ticket_1" }
  },
  output: {
    destination_ids: ["dest_1"],
    message: "Updated",
    payload: { ticket_id: "ticket_1", routed: true }
  },
  error: "",
  run_id: "run_1"
};

assert.deepStrictEqual(notificationHookPreviewSummary(display), {
  loaded: true,
  state: "completed",
  suppressed: true,
  destinations: 0,
  planned_destinations: 1,
  destination_override: true,
  message_override: true,
  payload_override: true,
  has_error: false,
  run_id: "run_1"
});

assert.deepStrictEqual(notificationHookPreviewSummaryItems(notificationHookPreviewSummary(display)), [
  "state: completed",
  "suppressed: yes",
  "destinations: 0",
  "planned destinations: 1",
  "destination override: yes",
  "message override: yes",
  "payload override: yes",
  "error: no",
  "run: run_1"
]);

assert.deepStrictEqual(notificationHookPreviewSummaryItems(notificationHookPreviewSummary({
  state: "failed",
  suppressed: false,
  plan: { destination_ids: [] },
  output: {},
  error: "bad route",
  run_id: ""
})), [
  "state: failed",
  "suppressed: no",
  "destinations: 0",
  "destination override: no",
  "message override: no",
  "payload override: no",
  "error: yes"
]);

assert.deepStrictEqual(notificationHookPreviewSummary(null), {
  loaded: false,
  state: "",
  suppressed: false,
  destinations: 0,
  planned_destinations: 0,
  destination_override: false,
  message_override: false,
  payload_override: false,
  has_error: false,
  run_id: ""
});
assert.deepStrictEqual(notificationHookPreviewSummaryItems(notificationHookPreviewSummary(null)), ["preview: none"]);
