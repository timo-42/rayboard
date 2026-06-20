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
  boardAssigneeWorkloads,
  boardCapacityOverview,
  boardCapacityOverviewLabel,
  boardColumnTicketCount,
  boardDueDateBreakdown,
  boardFlowBalance,
  boardFlowBalanceItems,
  boardIssueTypeBreakdown,
  boardLabelBreakdown,
  boardPriorityBreakdown,
  boardRiskOverview,
  boardRiskOverviewLabel,
  boardSummaryMetrics
} = require("./static/app.js");

const columns = [
  {
    slug: "todo",
    name: "To Do",
    ticket_count: 2,
    wip_limit: 5,
    tickets: [
      { id: "a", status: "todo", assignee_id: "alice", story_points: 3, type: "Bug", priority: "High", labels: ["backend", "urgent"], due_date: "2026-06-18", updated_at: "2026-06-10T09:00:00Z" },
      { id: "b", status: "todo", assignee_id: "bob", story_points: 5, type: "Story", priority: "Low", labels: ["frontend"], due_date: "2026-06-25", updated_at: "2026-06-20T09:00:00Z" }
    ]
  },
  {
    slug: "doing",
    name: "Doing",
    tickets: [
      { id: "c", status: "blocked", assignee_id: "alice", story_points: 2, type: "Bug", priority: "Medium", labels: ["backend"], due_date: "2026-06-30", updated_at: "2026-06-01T09:00:00Z" },
      { id: "d", status: "in_progress", assignee_id: "bob", story_points: "", type: "Task", priority: "Critical", labels: ["urgent"], due_date: "2026-06-19", updated_at: "2026-06-20T09:00:00Z" },
      { id: "e", status: "in_progress", assignee_id: "", story_points: 1, type: "", priority: "Low", labels: [], due_date: "", updated_at: "2026-06-20T09:00:00Z" },
      { id: "f", status: "done", assignee_id: "alice", story_points: 8, type: "Bug", priority: "Critical", labels: ["backend", "frontend"], due_date: "2026-06-01", updated_at: "2026-05-01T09:00:00Z" },
      { id: "g", status: "in_progress", assignee_id: "", story_points: null, type: "Story", priority: "", labels: [], due_date: "", updated_at: "2026-06-20T09:00:00Z" }
    ],
    wip_limit: 3
  },
  {
    slug: "done",
    name: "Done",
    ticket_count: 7,
    wip_limit: 0,
    over_wip_limit: true
  },
  {
    slug: "review",
    name: "Review",
    ticket_count: 3,
    wip_limit: 3
  },
  {
    slug: "qa",
    name: "QA",
    ticket_count: 0,
    wip_limit: 2
  }
];

assert.strictEqual(boardColumnTicketCount(columns[0]), 2);
assert.strictEqual(boardColumnTicketCount(columns[1]), 5);
assert.strictEqual(boardColumnTicketCount({}), 0);

assert.deepStrictEqual(boardCapacityOverview(columns), [
  {
    slug: "todo",
    name: "To Do",
    ticket_count: 2,
    wip_limit: 5,
    status: "limited",
    remaining: 3,
    overage: 0
  },
  {
    slug: "doing",
    name: "Doing",
    ticket_count: 5,
    wip_limit: 3,
    status: "over_limit",
    remaining: 0,
    overage: 2
  },
  {
    slug: "done",
    name: "Done",
    ticket_count: 7,
    wip_limit: null,
    status: "unlimited",
    remaining: null,
    overage: 0
  },
  {
    slug: "review",
    name: "Review",
    ticket_count: 3,
    wip_limit: 3,
    status: "limited",
    remaining: 0,
    overage: 0
  },
  {
    slug: "qa",
    name: "QA",
    ticket_count: 0,
    wip_limit: 2,
    status: "limited",
    remaining: 2,
    overage: 0
  }
]);

assert.deepStrictEqual(
  boardSummaryMetrics({ filtered_by_saved_view: true, columns }, []),
  {
    total_tickets: 17,
    column_count: 5,
    wip_warnings: 1,
    saved_view_filter: "filtered",
    flow_balance: boardFlowBalance(columns),
    priorities: boardPriorityBreakdown(columns),
    due_dates: boardDueDateBreakdown(columns),
    issue_types: boardIssueTypeBreakdown(columns),
    labels: boardLabelBreakdown(columns),
    assignee_workloads: boardAssigneeWorkloads(columns),
    capacity: boardCapacityOverview(columns),
    risks: boardRiskOverview(columns)
  }
);

assert.deepStrictEqual(
  boardSummaryMetrics(null, [{ slug: "todo", name: "To Do", tickets: [{ id: "a" }], wip_limit: null }]),
  {
    total_tickets: 1,
    column_count: 1,
    wip_warnings: 0,
    saved_view_filter: "all tickets",
    flow_balance: {
      empty_columns: 0,
      at_limit_columns: 0,
      over_limit_columns: 0,
      bottleneck: {
        slug: "todo",
        name: "To Do",
        ticket_count: 1,
        wip_limit: null,
        at_limit: false,
        over_limit: false
      }
    },
    priorities: [
      { label: "No priority", count: 1 }
    ],
    due_dates: [
      { key: "none", label: "No due date", count: 1 }
    ],
    issue_types: [
      { label: "No issue type", count: 1 }
    ],
    labels: [
      { label: "No labels", count: 1 }
    ],
    assignee_workloads: [
      { key: "", label: "Unassigned", tickets: 1, story_points: 0, has_story_points: false }
    ],
    capacity: [
      {
        slug: "todo",
        name: "To Do",
        ticket_count: 1,
        wip_limit: null,
        status: "unlimited",
        remaining: null,
        overage: 0
      }
    ],
    risks: []
  }
);

assert.strictEqual(boardCapacityOverviewLabel(boardCapacityOverview(columns)[0]), "To Do: 2/5, 3 remaining");
assert.strictEqual(boardCapacityOverviewLabel(boardCapacityOverview(columns)[1]), "Doing: 5/3, 2 over limit");
assert.strictEqual(boardCapacityOverviewLabel(boardCapacityOverview(columns)[2]), "Done: 7, unlimited");

assert.deepStrictEqual(boardFlowBalance(columns), {
  empty_columns: 1,
  at_limit_columns: 1,
  over_limit_columns: 1,
  bottleneck: {
    slug: "done",
    name: "Done",
    ticket_count: 7,
    wip_limit: null,
    at_limit: false,
    over_limit: false
  }
});

assert.deepStrictEqual(boardFlowBalanceItems(boardFlowBalance(columns)), [
  "empty columns: 1",
  "at WIP limit: 1",
  "over WIP limit: 1",
  "bottleneck: Done (7)"
]);
assert.deepStrictEqual(boardFlowBalanceItems(boardFlowBalance(null)), [
  "empty columns: 0",
  "at WIP limit: 0",
  "over WIP limit: 0",
  "bottleneck: none"
]);

assert.deepStrictEqual(boardPriorityBreakdown(columns), [
  { label: "Critical", count: 2 },
  { label: "Low", count: 2 },
  { label: "High", count: 1 },
  { label: "Medium", count: 1 },
  { label: "No priority", count: 1 }
]);
assert.deepStrictEqual(boardPriorityBreakdown(null), []);

assert.deepStrictEqual(boardDueDateBreakdown(columns, "2026-06-20"), [
  { key: "overdue", label: "Overdue", count: 3 },
  { key: "later", label: "Later", count: 2 },
  { key: "none", label: "No due date", count: 2 }
]);
assert.deepStrictEqual(boardDueDateBreakdown(null, "2026-06-20"), []);

assert.deepStrictEqual(boardIssueTypeBreakdown(columns), [
  { label: "Bug", count: 3 },
  { label: "Story", count: 2 },
  { label: "No issue type", count: 1 },
  { label: "Task", count: 1 }
]);
assert.deepStrictEqual(boardIssueTypeBreakdown(null), []);

assert.deepStrictEqual(boardLabelBreakdown(columns), [
  { label: "backend", count: 3 },
  { label: "frontend", count: 2 },
  { label: "urgent", count: 2 },
  { label: "No labels", count: 2 }
]);
assert.deepStrictEqual(boardLabelBreakdown(null), []);

assert.deepStrictEqual(boardAssigneeWorkloads(columns), [
  { key: "alice", label: "assignee alice", tickets: 3, story_points: 13, has_story_points: true },
  { key: "bob", label: "assignee bob", tickets: 2, story_points: 5, has_story_points: true },
  { key: "", label: "Unassigned", tickets: 2, story_points: 1, has_story_points: true }
]);
assert.deepStrictEqual(boardAssigneeWorkloads(null), []);

assert.deepStrictEqual(boardRiskOverview(columns, "2026-06-20"), [
  {
    slug: "doing",
    name: "Doing",
    blocked: 1,
    overdue: 1,
    stale: 1,
    high_priority: 1,
    total: 4
  },
  {
    slug: "todo",
    name: "To Do",
    blocked: 0,
    overdue: 1,
    stale: 1,
    high_priority: 1,
    total: 3
  }
]);

assert.deepStrictEqual(boardRiskOverview(null, "2026-06-20"), []);
assert.strictEqual(boardRiskOverviewLabel(boardRiskOverview(columns, "2026-06-20")[0]), "Doing: 1 blocked, 1 overdue, 1 stale, 1 high priority");
