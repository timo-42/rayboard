package luasandbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"

	lua "github.com/yuin/gopher-lua"
)

const (
	defaultMaxJSONInputBytes  = 1 << 20
	defaultMaxJSONOutputBytes = 1 << 20
	defaultMaxJSONDepth       = 64
)

var (
	ErrUnsupportedValue = errors.New("luasandbox: unsupported JSON value")
	ErrLimitExceeded    = errors.New("luasandbox: JSON limit exceeded")
)

type JSONLimits struct {
	MaxInputBytes  int
	MaxOutputBytes int
	MaxDepth       int
}

func DefaultJSONLimits() JSONLimits {
	return JSONLimits{
		MaxInputBytes:  defaultMaxJSONInputBytes,
		MaxOutputBytes: defaultMaxJSONOutputBytes,
		MaxDepth:       defaultMaxJSONDepth,
	}
}

type Sandbox struct {
	L    *lua.LState
	JSON *JSONModule
}

func New(limits JSONLimits) *Sandbox {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})
	lua.OpenBase(L)
	lua.OpenTable(L)
	lua.OpenString(L)
	lua.OpenMath(L)

	return &Sandbox{
		L:    L,
		JSON: RegisterJSON(L, limits),
	}
}

func (s *Sandbox) Close() {
	if s == nil || s.L == nil {
		return
	}
	s.L.Close()
}

type JSONModule struct {
	L      *lua.LState
	Limits JSONLimits
	Null   *lua.LUserData
}

func RegisterJSON(L *lua.LState, limits JSONLimits) *JSONModule {
	limits = normalizeJSONLimits(limits)
	module := &JSONModule{
		L:      L,
		Limits: limits,
		Null:   newJSONNull(L),
	}

	table := L.NewTable()
	L.SetField(table, "null", module.Null)
	L.SetField(table, "encode", L.NewFunction(module.luaEncode))
	L.SetField(table, "decode", L.NewFunction(module.luaDecode))
	L.SetGlobal("json", table)

	rayboard := L.GetGlobal("rayboard")
	rayboardTable, ok := rayboard.(*lua.LTable)
	if !ok {
		rayboardTable = L.NewTable()
		L.SetGlobal("rayboard", rayboardTable)
	}
	L.SetField(rayboardTable, "json", table)

	return module
}

func (m *JSONModule) Encode(value lua.LValue) (string, error) {
	converted, err := m.ToGo(value)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(converted)
	if err != nil {
		return "", fmt.Errorf("encode JSON: %w", err)
	}
	if len(encoded) > m.Limits.MaxOutputBytes {
		return "", fmt.Errorf("%w: encoded JSON exceeds %d bytes", ErrLimitExceeded, m.Limits.MaxOutputBytes)
	}
	return string(encoded), nil
}

func (m *JSONModule) Decode(text string) (lua.LValue, error) {
	if len(text) > m.Limits.MaxInputBytes {
		return lua.LNil, fmt.Errorf("%w: JSON input exceeds %d bytes", ErrLimitExceeded, m.Limits.MaxInputBytes)
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(text)))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return lua.LNil, fmt.Errorf("decode JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return lua.LNil, errors.New("decode JSON: trailing data")
	}
	return m.FromGo(value)
}

func (m *JSONModule) ToGo(value lua.LValue) (any, error) {
	return m.toGo(value, 0, map[*lua.LTable]struct{}{})
}

func (m *JSONModule) FromGo(value any) (lua.LValue, error) {
	return m.fromGo(value, 0)
}

func (m *JSONModule) luaEncode(L *lua.LState) int {
	encoded, err := m.Encode(L.Get(1))
	if err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}
	L.Push(lua.LString(encoded))
	return 1
}

func (m *JSONModule) luaDecode(L *lua.LState) int {
	value, err := m.Decode(L.CheckString(1))
	if err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}
	L.Push(value)
	return 1
}

func (m *JSONModule) toGo(value lua.LValue, depth int, seen map[*lua.LTable]struct{}) (any, error) {
	if depth > m.Limits.MaxDepth {
		return nil, fmt.Errorf("%w: JSON depth exceeds %d", ErrLimitExceeded, m.Limits.MaxDepth)
	}
	switch typed := value.(type) {
	case lua.LBool:
		return bool(typed), nil
	case lua.LNumber:
		number := float64(typed)
		if math.IsInf(number, 0) || math.IsNaN(number) {
			return nil, fmt.Errorf("%w: non-finite number", ErrUnsupportedValue)
		}
		return number, nil
	case lua.LString:
		return string(typed), nil
	case *lua.LTable:
		return m.tableToGo(typed, depth, seen)
	case *lua.LUserData:
		if typed == m.Null {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: userdata cannot be encoded as JSON", ErrUnsupportedValue)
	case *lua.LNilType:
		return nil, fmt.Errorf("%w: use json.null for JSON null", ErrUnsupportedValue)
	default:
		return nil, fmt.Errorf("%w: %s cannot be encoded as JSON", ErrUnsupportedValue, value.Type().String())
	}
}

func (m *JSONModule) tableToGo(table *lua.LTable, depth int, seen map[*lua.LTable]struct{}) (any, error) {
	if _, ok := seen[table]; ok {
		return nil, fmt.Errorf("%w: recursive table", ErrUnsupportedValue)
	}
	seen[table] = struct{}{}
	defer delete(seen, table)

	stringKeys := []string{}
	arrayValues := map[int]lua.LValue{}
	maxIndex := 0

	var keyErr error
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if keyErr != nil {
			return
		}
		switch typed := key.(type) {
		case lua.LString:
			stringKeys = append(stringKeys, string(typed))
		case lua.LNumber:
			index := int(typed)
			if float64(index) != float64(typed) || index <= 0 {
				keyErr = fmt.Errorf("%w: array keys must be positive integers", ErrUnsupportedValue)
				return
			}
			arrayValues[index] = value
			if index > maxIndex {
				maxIndex = index
			}
		default:
			keyErr = fmt.Errorf("%w: table keys must be strings or array indexes", ErrUnsupportedValue)
		}
	})
	if keyErr != nil {
		return nil, keyErr
	}

	if len(stringKeys) > 0 && len(arrayValues) > 0 {
		return nil, fmt.Errorf("%w: mixed string and array keys", ErrUnsupportedValue)
	}
	if len(arrayValues) > 0 {
		return m.arrayTableToGo(arrayValues, maxIndex, depth, seen)
	}

	object := make(map[string]any, len(stringKeys))
	sort.Strings(stringKeys)
	for _, key := range stringKeys {
		converted, err := m.toGo(table.RawGetString(key), depth+1, seen)
		if err != nil {
			return nil, err
		}
		object[key] = converted
	}
	return object, nil
}

func (m *JSONModule) arrayTableToGo(values map[int]lua.LValue, maxIndex int, depth int, seen map[*lua.LTable]struct{}) ([]any, error) {
	array := make([]any, maxIndex)
	for index := 1; index <= maxIndex; index++ {
		value, ok := values[index]
		if !ok {
			return nil, fmt.Errorf("%w: array table has a hole at index %d", ErrUnsupportedValue, index)
		}
		converted, err := m.toGo(value, depth+1, seen)
		if err != nil {
			return nil, err
		}
		array[index-1] = converted
	}
	return array, nil
}

func (m *JSONModule) fromGo(value any, depth int) (lua.LValue, error) {
	if depth > m.Limits.MaxDepth {
		return lua.LNil, fmt.Errorf("%w: JSON depth exceeds %d", ErrLimitExceeded, m.Limits.MaxDepth)
	}
	switch typed := value.(type) {
	case nil:
		return m.Null, nil
	case bool:
		return lua.LBool(typed), nil
	case string:
		return lua.LString(typed), nil
	case json.Number:
		number, err := typed.Float64()
		if err != nil {
			return lua.LNil, fmt.Errorf("decode JSON number: %w", err)
		}
		if math.IsInf(number, 0) || math.IsNaN(number) {
			return lua.LNil, fmt.Errorf("%w: non-finite number", ErrUnsupportedValue)
		}
		return lua.LNumber(number), nil
	case int:
		return lua.LNumber(typed), nil
	case int64:
		return lua.LNumber(typed), nil
	case float64:
		if math.IsInf(typed, 0) || math.IsNaN(typed) {
			return lua.LNil, fmt.Errorf("%w: non-finite number", ErrUnsupportedValue)
		}
		return lua.LNumber(typed), nil
	case []any:
		return m.sliceToLua(typed, depth)
	case map[string]any:
		return m.mapToLua(typed, depth)
	default:
		return lua.LNil, fmt.Errorf("%w: unsupported Go value %T", ErrUnsupportedValue, value)
	}
}

func (m *JSONModule) sliceToLua(values []any, depth int) (lua.LValue, error) {
	table := m.L.NewTable()
	for index, item := range values {
		converted, err := m.fromGo(item, depth+1)
		if err != nil {
			return lua.LNil, err
		}
		table.RawSetInt(index+1, converted)
	}
	return table, nil
}

func (m *JSONModule) mapToLua(values map[string]any, depth int) (lua.LValue, error) {
	table := m.L.NewTable()
	for key, item := range values {
		converted, err := m.fromGo(item, depth+1)
		if err != nil {
			return lua.LNil, err
		}
		table.RawSetString(key, converted)
	}
	return table, nil
}

func newJSONNull(L *lua.LState) *lua.LUserData {
	userData := L.NewUserData()
	userData.Value = "json.null"
	meta := L.NewTable()
	L.SetField(meta, "__tostring", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("json.null"))
		return 1
	}))
	L.SetMetatable(userData, meta)
	return userData
}

func normalizeJSONLimits(limits JSONLimits) JSONLimits {
	defaults := DefaultJSONLimits()
	if limits.MaxInputBytes <= 0 {
		limits.MaxInputBytes = defaults.MaxInputBytes
	}
	if limits.MaxOutputBytes <= 0 {
		limits.MaxOutputBytes = defaults.MaxOutputBytes
	}
	if limits.MaxDepth <= 0 {
		limits.MaxDepth = defaults.MaxDepth
	}
	return limits
}
