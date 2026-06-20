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
  automationRunFailureBreakdown,
  automationRunFailureLabel,
  summarizeAutomationRuns
} = require("./static/app.js");

const runs = [
  {
    state: "failed",
    error: "timeout while calling provider",
    trigger_type: "manual",
    created_at: "2026-06-20T09:00:00Z",
    started_at: "2026-06-20T09:00:00Z",
    finished_at: "2026-06-20T09:00:05Z"
  },
  {
    state: "failed",
    error: "timeout while calling provider",
    trigger_type: "scheduled",
    created_at: "2026-06-20T09:01:00Z",
    started_at: "2026-06-20T09:01:00Z",
    finished_at: "2026-06-20T09:01:10Z"
  },
  {
    state: "canceled",
    error: "",
    trigger_type: "manual",
    created_at: "2026-06-20T09:02:00Z"
  },
  {
    state: "completed",
    error: "",
    trigger_type: "scheduled",
    created_at: "2026-06-20T09:03:00Z",
    started_at: "2026-06-20T09:03:00Z",
    finished_at: "2026-06-20T09:03:20Z"
  },
  {
    state: "queued",
    trigger_type: "",
    created_at: "not-a-date"
  }
];

assert.deepStrictEqual(automationRunFailureBreakdown(runs), [
  { label: "timeout while calling provider", count: 2 },
  { label: "canceled", count: 1 }
]);

assert.deepStrictEqual(automationRunFailureBreakdown([
  { state: "failed", error: "b" },
  { state: "failed", error: "a" },
  { state: "failed", error: "c" }
], 2), [
  { label: "a", count: 1 },
  { label: "b", count: 1 }
]);

assert.strictEqual(automationRunFailureLabel({
  state: "failed",
  error: "  line one\n\nline two\tline three  "
}), "line one line two line three");

assert.strictEqual(
  automationRunFailureLabel({ state: "failed", error: "x".repeat(80) }),
  `${"x".repeat(69)}...`
);

assert.deepStrictEqual(automationRunFailureBreakdown(null), []);

const summary = summarizeAutomationRuns(runs);
assert.strictEqual(summary.total, 5);
assert.strictEqual(summary.completed, 1);
assert.strictEqual(summary.failed, 3);
assert.strictEqual(summary.active, 1);
assert.strictEqual(summary.latestFailure, "timeout while calling provider");
assert.strictEqual(summary.completionRateLabel, "20%");
assert.strictEqual(summary.failureRateLabel, "60%");
assert.strictEqual(summary.averageDurationLabel, "12s");
assert.strictEqual(summary.maxDurationLabel, "20s");
assert.deepStrictEqual(summary.triggerCounts, {
  manual: 2,
  scheduled: 2,
  unknown: 1
});
assert.deepStrictEqual(summary.failures, [
  { label: "timeout while calling provider", count: 2 },
  { label: "canceled", count: 1 }
]);
