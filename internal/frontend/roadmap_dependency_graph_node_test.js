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
  setAttribute: noop,
  insertAdjacentElement: noop
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
  roadmapDependencyGraph,
  roadmapDependencyGraphNodeLabel
} = require("./static/app.js");

const graph = roadmapDependencyGraph([
  {
    source_epic_id: "epic-1",
    target_epic_id: "epic-2",
    link: {
      id: "link-1",
      link_type: "blocks",
      source: { id: "epic-1", key: "CORE-1", title: "Auth Epic", status: "todo" },
      target: { id: "ticket-2", key: "CORE-2", title: "Login UI", status: "in_progress" }
    }
  },
  {
    source_epic_id: "epic-1",
    target_epic_id: "epic-1",
    link: {
      id: "link-2",
      link_type: "relates_to",
      source: { id: "ticket-3", key: "CORE-3", title: "Session API", status: "todo" },
      target: { id: "ticket-4", key: "CORE-4", title: "Token API", status: "done" }
    }
  },
  {
    source_epic_id: "epic-1",
    target_epic_id: "epic-1",
    link: {
      id: "",
      link_type: "blocks",
      source: { id: "ticket-5", key: "CORE-5", title: "Broken edge" },
      target: null
    }
  }
]);

assert.deepStrictEqual(
  graph.nodes.map((node) => ({
    id: node.id,
    label: roadmapDependencyGraphNodeLabel(node),
    type: node.node_type
  })),
  [
    { id: "epic-1", label: "CORE-1 Auth Epic", type: "epic" },
    { id: "ticket-2", label: "CORE-2 Login UI", type: "issue" },
    { id: "ticket-3", label: "CORE-3 Session API", type: "issue" },
    { id: "ticket-4", label: "CORE-4 Token API", type: "issue" },
    { id: "ticket-5", label: "CORE-5 Broken edge", type: "issue" }
  ]
);

assert.deepStrictEqual(
  graph.edges.map((edge) => ({
    id: edge.id,
    source: edge.source_label,
    label: edge.label,
    target: edge.target_label,
    scope: edge.scope
  })),
  [
    {
      id: "link-1",
      source: "CORE-1 Auth Epic",
      label: "blocks",
      target: "CORE-2 Login UI",
      scope: "cross_epic"
    },
    {
      id: "link-2",
      source: "CORE-3 Session API",
      label: "relates to",
      target: "CORE-4 Token API",
      scope: "same_epic"
    }
  ]
);

assert.strictEqual(graph.incomplete, 1);
assert.deepStrictEqual(roadmapDependencyGraph(null), { nodes: [], edges: [], incomplete: 0 });
assert.strictEqual(roadmapDependencyGraphNodeLabel(null), "issue");
