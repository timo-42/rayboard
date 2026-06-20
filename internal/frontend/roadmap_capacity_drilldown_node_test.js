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
  roadmapCapacityBucketTargetLabel,
  roadmapCapacityBucketTargetStatus,
  roadmapCapacityDrilldown,
  roadmapCapacityTargetForBucket,
  roadmapCapacityTargetMap,
  roadmapCapacityTargetValue
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

assert.strictEqual(roadmapCapacityTargetValue("12.5"), 12.5);
assert.strictEqual(roadmapCapacityTargetValue(""), 0);
assert.strictEqual(roadmapCapacityTargetValue("-4"), 0);
assert.strictEqual(roadmapCapacityTargetValue("not-a-number"), 0);
assert.strictEqual(roadmapCapacityTargetForBucket("2026-07", { "2026-07": "12.5", "2026-08": "8" }), "12.5");
assert.strictEqual(roadmapCapacityTargetForBucket("2026-09", { "2026-07": "12.5" }), "");
assert.strictEqual(roadmapCapacityTargetForBucket("2026-07", "14"), "14");
assert.deepStrictEqual(
  roadmapCapacityTargetMap([
    { metadata: { project_id: "project_1", month: "2026-07" }, spec: { target_points: 12.5 }, status: {} },
    { metadata: { project_id: "project_1", month: "2026-08" }, spec: { target_points: 0 }, status: {} },
    { metadata: { project_id: "project_1", month: "2026-09" }, spec: { target_points: 9 }, status: { deleted: true } }
  ]),
  { "2026-07": "12.5" }
);

assert.deepStrictEqual(
  roadmapCapacityBucketTargetStatus({ storyPointsRemaining: 8 }, 12),
  {
    hasTarget: true,
    target: 12,
    remaining: 4,
    over: 0,
    overTarget: false,
    label: "4 pts target room"
  }
);

assert.deepStrictEqual(
  roadmapCapacityBucketTargetStatus({ storyPointsRemaining: 15.5 }, "12"),
  {
    hasTarget: true,
    target: 12,
    remaining: 0,
    over: 3.5,
    overTarget: true,
    label: "over target by 3.5 pts"
  }
);

assert.deepStrictEqual(
  roadmapCapacityBucketTargetStatus({ storyPointsRemaining: 15.5 }, ""),
  {
    hasTarget: false,
    target: 0,
    remaining: 0,
    over: 0,
    overTarget: false,
    label: ""
  }
);

assert.strictEqual(roadmapCapacityBucketTargetLabel({ storyPointsRemaining: 15.5 }, "12"), "over target by 3.5 pts");
