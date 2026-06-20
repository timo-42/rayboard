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
  createPageLayoutBuilderFieldItem,
  createPageLayoutBuilderFieldType,
  createPageLayoutBuilderItemKind,
  createPageLayoutBuilderTextKind,
  mutateCreatePageLayoutBuilderItems,
  parseCreatePageLayoutBuilderJSON
} = require("./static/app.js");

assert.deepStrictEqual(
  parseCreatePageLayoutBuilderJSON('[{"key":"title"}]'),
  [{ key: "title" }]
);
assert.strictEqual(parseCreatePageLayoutBuilderJSON('{"key":"title"}'), null);
assert.strictEqual(parseCreatePageLayoutBuilderJSON("not json"), null);

assert.strictEqual(createPageLayoutBuilderFieldType("description"), "textarea");
assert.strictEqual(createPageLayoutBuilderFieldType("story_points"), "number");
assert.strictEqual(createPageLayoutBuilderFieldType("start_date"), "date");
assert.strictEqual(createPageLayoutBuilderFieldType("custom_fields.severity"), "text");

assert.strictEqual(createPageLayoutBuilderItemKind({ key: "title" }), "field");
assert.strictEqual(createPageLayoutBuilderItemKind({ kind: "heading", text: "Bug details" }), "text");
assert.strictEqual(createPageLayoutBuilderItemKind({ html: "<b>bad</b>" }), "unsupported");
assert.strictEqual(createPageLayoutBuilderItemKind({ fields: [{ key: "title" }] }), "unsupported");

assert.strictEqual(createPageLayoutBuilderTextKind({ kind: "heading" }), "heading");
assert.strictEqual(createPageLayoutBuilderTextKind({ kind: "help" }), "help");
assert.strictEqual(createPageLayoutBuilderTextKind({ type: "paragraph" }), "help");

assert.deepStrictEqual(createPageLayoutBuilderFieldItem("title"), {
  key: "title",
  label: "Title",
  type: "text",
  required: true
});
assert.deepStrictEqual(createPageLayoutBuilderFieldItem("due_date"), {
  key: "due_date",
  label: "Due Date",
  type: "date",
  required: false
});

const layout = [{ key: "title" }, { key: "priority" }, { kind: "help", text: "Help" }];
mutateCreatePageLayoutBuilderItems(layout, 1, "up");
assert.deepStrictEqual(layout.map((item) => item.key || item.kind), ["priority", "title", "help"]);
mutateCreatePageLayoutBuilderItems(layout, 0, "down");
assert.deepStrictEqual(layout.map((item) => item.key || item.kind), ["title", "priority", "help"]);
mutateCreatePageLayoutBuilderItems(layout, 1, "remove");
assert.deepStrictEqual(layout.map((item) => item.key || item.kind), ["title", "help"]);
