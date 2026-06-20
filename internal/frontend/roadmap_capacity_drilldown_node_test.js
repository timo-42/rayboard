const assert = require("assert");

global.document = {
  createElement(tag) {
    return {
      tagName: tag.toUpperCase(),
      className: "",
      dataset: {},
      style: {},
      children: [],
      textContent: "",
      type: "",
      append(...children) {
        this.children.push(...children);
      },
      setAttribute(name, value) {
        this[name] = value;
      },
      classList: {
        add() {}
      }
    };
  },
  querySelector() {
    return null;
  },
  addEventListener() {}
};

global.window = { addEventListener() {} };

const {
  roadmapCapacityDrilldown
} = require("./static/app.js");

const items = [
  {
    epic: {
      id: "epic-low",
      key: "CORE-2",
      title: "Small follow-up",
      start_date: "2026-07-01",
      due_date: "2026-07-05"
    },
    progress: { total: 3, done: 2 }
  },
  {
    epic: {
      id: "epic-risk",
      key: "CORE-1",
      title: "Large launch",
      start_date: "2026-07-08",
      due_date: "2026-07-31"
    },
    progress: { total: 10, done: 1 }
  },
  {
    epic: {
      id: "epic-other",
      key: "CORE-3",
      title: "August launch",
      start_date: "2026-08-01",
      due_date: "2026-08-15"
    },
    progress: { total: 5, done: 5 }
  },
  {
    epic: {
      id: "epic-points",
      key: "CORE-4",
      title: "Point heavy",
      start_date: "2026-07-03",
      due_date: "2026-07-20"
    },
    progress: { total: 4, done: 1 }
  }
];

const july = roadmapCapacityDrilldown(items, "2026-07");

assert.deepStrictEqual(
  july.map((row) => ({
    id: row.id,
    label: row.label,
    open: row.openChildren,
    done: row.doneChildren,
    total: row.childTickets,
    atRisk: row.atRisk,
    range: row.date_range
  })),
  [
    {
      id: "epic-risk",
      label: "CORE-1 Large launch",
      open: 9,
      done: 1,
      total: 10,
      atRisk: true,
      range: "2026-07-08 to 2026-07-31"
    },
    {
      id: "epic-points",
      label: "CORE-4 Point heavy",
      open: 3,
      done: 1,
      total: 4,
      atRisk: true,
      range: "2026-07-03 to 2026-07-20"
    },
    {
      id: "epic-low",
      label: "CORE-2 Small follow-up",
      open: 1,
      done: 2,
      total: 3,
      atRisk: false,
      range: "2026-07-01 to 2026-07-05"
    }
  ]
);

assert.deepStrictEqual(roadmapCapacityDrilldown(items, "2026-09"), []);
assert.deepStrictEqual(roadmapCapacityDrilldown(null, "2026-07"), []);
assert.deepStrictEqual(roadmapCapacityDrilldown(items, ""), []);
