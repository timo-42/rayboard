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
  notificationDeliveryAnalytics,
  notificationDeliveryAnalyticsLabel
} = require("./static/app.js");

const deliveries = [
  {
    event_type: "ticket_assigned",
    destination_id: "dest_slack",
    destination_name: "Slack alerts",
    destination_service: "slack",
    state: "delivered",
    attempt_count: 1,
    max_attempts: 3
  },
  {
    event_type: "ticket_assigned",
    destination_id: "dest_slack",
    destination_name: "Slack alerts",
    destination_service: "slack",
    state: "failed",
    attempt_count: 3,
    max_attempts: 3
  },
  {
    event_type: "ticket_assigned",
    destination_id: "dest_slack",
    destination_name: "Slack alerts",
    destination_service: "slack",
    state: "queued",
    attempt_count: 0,
    max_attempts: 3
  },
  {
    event_type: "comment_created",
    destination_id: "dest_email",
    destination_name: "Email desk",
    destination_service: "smtp",
    state: "queued",
    attempt_count: 0,
    max_attempts: 2
  },
  {
    event_type: "comment_created",
    destination_id: "dest_email",
    destination_name: "Email desk",
    destination_service: "smtp",
    state: "sending",
    attempt_count: 2,
    max_attempts: 2
  },
  {
    event_type: "",
    destination_id: "",
    destination_name: "",
    destination_service: "",
    state: "canceled",
    attempt_count: 1,
    max_attempts: 1
  }
];

const analytics = notificationDeliveryAnalytics(deliveries);

assert.deepStrictEqual(analytics.by_event, [
  {
    key: "ticket_assigned",
    label: "ticket_assigned",
    total: 3,
    delivered: 1,
    failed: 1,
    queued: 1,
    sending: 0,
    canceled: 0
  },
  {
    key: "comment_created",
    label: "comment_created",
    total: 2,
    delivered: 0,
    failed: 0,
    queued: 1,
    sending: 1,
    canceled: 0
  },
  {
    key: "unknown event",
    label: "unknown event",
    total: 1,
    delivered: 0,
    failed: 0,
    queued: 0,
    sending: 0,
    canceled: 1
  }
]);

assert.deepStrictEqual(analytics.by_destination, [
  {
    key: "dest_slack",
    label: "Slack alerts / slack",
    total: 3,
    delivered: 1,
    failed: 1,
    queued: 1,
    sending: 0,
    canceled: 0
  },
  {
    key: "dest_email",
    label: "Email desk / smtp",
    total: 2,
    delivered: 0,
    failed: 0,
    queued: 1,
    sending: 1,
    canceled: 0
  },
  {
    key: "unknown destination",
    label: "unknown destination",
    total: 1,
    delivered: 0,
    failed: 0,
    queued: 0,
    sending: 0,
    canceled: 1
  }
]);

assert.strictEqual(analytics.retry_pressure, 3);
assert.strictEqual(analytics.exhausted, 3);
assert.strictEqual(notificationDeliveryAnalyticsLabel(analytics.by_event[0]), "ticket_assigned: 3 / 1 delivered / 1 failed / 1 queued");

assert.deepStrictEqual(notificationDeliveryAnalytics([]), {
  by_event: [],
  by_destination: [],
  retry_pressure: 0,
  exhausted: 0
});
assert.deepStrictEqual(notificationDeliveryAnalytics(null), {
  by_event: [],
  by_destination: [],
  retry_pressure: 0,
  exhausted: 0
});
