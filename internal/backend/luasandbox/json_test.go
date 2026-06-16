package luasandbox

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestJSONEncodeDecodeFromLua(t *testing.T) {
	sandbox := New(JSONLimits{})
	defer sandbox.Close()

	if err := sandbox.L.DoString(`
		local payload = {
			name = "Ticket",
			open = true,
			count = 3,
			empty = json.null,
			labels = {"backend", "api"},
			nested = { status = "todo" }
		}
		encoded = json.encode(payload)
		decoded = rayboard.json.decode(encoded)
	`); err != nil {
		t.Fatalf("run lua: %v", err)
	}

	encoded := sandbox.L.GetGlobal("encoded").String()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatalf("decode encoded json: %v", err)
	}
	if decoded["name"] != "Ticket" || decoded["open"] != true || decoded["empty"] != nil {
		t.Fatalf("unexpected encoded json: %#v", decoded)
	}

	luaDecoded := sandbox.L.GetGlobal("decoded").(*lua.LTable)
	if got := luaDecoded.RawGetString("empty"); got != sandbox.JSON.Null {
		t.Fatalf("expected json null sentinel, got %#v", got)
	}
	labels := luaDecoded.RawGetString("labels").(*lua.LTable)
	if labels.RawGetInt(1).String() != "backend" || labels.RawGetInt(2).String() != "api" {
		t.Fatalf("unexpected labels: %#v", labels)
	}
}

func TestJSONGoLuaConversions(t *testing.T) {
	sandbox := New(JSONLimits{})
	defer sandbox.Close()

	value, err := sandbox.JSON.FromGo(map[string]any{
		"title":       "Bug",
		"assignee_id": nil,
		"labels":      []any{"p1", "backend"},
	})
	if err != nil {
		t.Fatalf("from go: %v", err)
	}
	table := value.(*lua.LTable)
	if table.RawGetString("assignee_id") != sandbox.JSON.Null {
		t.Fatalf("expected null sentinel")
	}

	got, err := sandbox.JSON.ToGo(table)
	if err != nil {
		t.Fatalf("to go: %v", err)
	}
	want := map[string]any{
		"title":       "Bug",
		"assignee_id": nil,
		"labels":      []any{"p1", "backend"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected conversion:\nwant %#v\ngot  %#v", want, got)
	}
}

func TestJSONRejectsUnsupportedTables(t *testing.T) {
	tests := map[string]string{
		"mixed":     `return json.encode({ "first", name = "mixed" })`,
		"sparse":    `local t = {}; t[2] = "hole"; return json.encode(t)`,
		"recursive": `local t = {}; t.self = t; return json.encode(t)`,
		"function":  `return json.encode({ f = function() end })`,
		"nil":       `return json.encode(nil)`,
	}

	for name, script := range tests {
		t.Run(name, func(t *testing.T) {
			sandbox := New(JSONLimits{})
			defer sandbox.Close()

			err := sandbox.L.DoString(script)
			if err == nil {
				t.Fatal("expected script to fail")
			}
			if !strings.Contains(err.Error(), ErrUnsupportedValue.Error()) {
				t.Fatalf("expected unsupported value error, got %v", err)
			}
		})
	}
}

func TestJSONLimits(t *testing.T) {
	sandbox := New(JSONLimits{
		MaxInputBytes:  8,
		MaxOutputBytes: 8,
		MaxDepth:       1,
	})
	defer sandbox.Close()

	if _, err := sandbox.JSON.Decode(`{"too":"long"}`); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected input limit error, got %v", err)
	}

	wide := sandbox.L.NewTable()
	wide.RawSetString("value", lua.LString("too long"))
	if _, err := sandbox.JSON.Encode(wide); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected output limit error, got %v", err)
	}

	nested := sandbox.L.NewTable()
	child := sandbox.L.NewTable()
	grandchild := sandbox.L.NewTable()
	child.RawSetString("child", grandchild)
	nested.RawSetString("nested", child)
	if _, err := sandbox.JSON.Encode(nested); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected depth limit error, got %v", err)
	}
}

func TestJSONDecodeRejectsTrailingData(t *testing.T) {
	sandbox := New(JSONLimits{})
	defer sandbox.Close()

	if _, err := sandbox.JSON.Decode(`{} {}`); err == nil || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("expected trailing data error, got %v", err)
	}
}
