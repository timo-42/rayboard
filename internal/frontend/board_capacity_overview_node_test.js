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
  boardCapacityOverview,
  boardCapacityOverviewLabel,
  boardColumnTicketCount,
  boardSummaryMetrics
} = require("./static/app.js");

const columns = [
  {
    slug: "todo",
    name: "To Do",
    ticket_count: 2,
    wip_limit: 5,
    tickets: [{ id: "ignored" }]
  },
  {
    slug: "doing",
    name: "Doing",
    tickets: [{ id: "a" }, { id: "b" }, { id: "c" }, { id: "d" }],
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
  }
];

assert.strictEqual(boardColumnTicketCount(columns[0]), 2);
assert.strictEqual(boardColumnTicketCount(columns[1]), 4);
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
    ticket_count: 4,
    wip_limit: 3,
    status: "over_limit",
    remaining: 0,
    overage: 1
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
  }
]);

assert.deepStrictEqual(
  boardSummaryMetrics({ filtered_by_saved_view: true, columns }, []),
  {
    total_tickets: 16,
    column_count: 4,
    wip_warnings: 1,
    saved_view_filter: "filtered",
    capacity: boardCapacityOverview(columns)
  }
);

assert.deepStrictEqual(
  boardSummaryMetrics(null, [{ slug: "todo", name: "To Do", tickets: [{ id: "a" }], wip_limit: null }]),
  {
    total_tickets: 1,
    column_count: 1,
    wip_warnings: 0,
    saved_view_filter: "all tickets",
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
    ]
  }
);

assert.strictEqual(boardCapacityOverviewLabel(boardCapacityOverview(columns)[0]), "To Do: 2/5, 3 remaining");
assert.strictEqual(boardCapacityOverviewLabel(boardCapacityOverview(columns)[1]), "Doing: 4/3, 1 over limit");
assert.strictEqual(boardCapacityOverviewLabel(boardCapacityOverview(columns)[2]), "Done: 7, unlimited");
