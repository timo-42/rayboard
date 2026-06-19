package cronjobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	lua "github.com/yuin/gopher-lua"
)

func (s *Service) registerLuaHelpers(ctx context.Context, sandbox *luasandbox.Sandbox, job Job, logs *[]string) {
	rayboard := sandbox.L.GetGlobal("rayboard")
	rayboardTable, ok := rayboard.(*lua.LTable)
	if !ok {
		rayboardTable = sandbox.L.NewTable()
		sandbox.L.SetGlobal("rayboard", rayboardTable)
	}

	sandbox.L.SetField(rayboardTable, "log", sandbox.L.NewFunction(func(L *lua.LState) int {
		if len(*logs) < 100 {
			*logs = append(*logs, L.CheckString(1))
		}
		return 0
	}))
	sandbox.L.SetField(rayboardTable, "search", sandbox.L.NewFunction(s.luaSearch(ctx, sandbox, job)))
	sandbox.L.SetField(rayboardTable, "get_ticket", sandbox.L.NewFunction(s.luaGetTicket(ctx, sandbox, job)))
	sandbox.L.SetField(rayboardTable, "create_ticket", sandbox.L.NewFunction(s.luaCreateTicket(ctx, sandbox, job)))
	sandbox.L.SetField(rayboardTable, "update_ticket", sandbox.L.NewFunction(s.luaUpdateTicket(ctx, sandbox, job)))
	sandbox.L.SetField(rayboardTable, "comment", sandbox.L.NewFunction(s.luaComment(ctx, sandbox, job)))
}

func (s *Service) luaSearch(ctx context.Context, sandbox *luasandbox.Sandbox, job Job) lua.LGFunction {
	return func(L *lua.LState) int {
		if s.search == nil {
			return pushLuaError(L, sandbox, "rayboard.search is not configured")
		}
		input, ok := tableArg(L, sandbox, 1)
		if !ok {
			return pushLuaError(L, sandbox, "rayboard.search expects a table")
		}
		result, err := s.search.SearchTickets(ctx, cronPrincipal(job), search.SearchTicketsInput{
			ProjectID: stringValue(input, "project_id"),
			Filter:    stringValue(input, "filter"),
			Text:      stringValue(input, "text"),
			Sort:      sortSpecs(input["sort"]),
			Limit:     intValue(input, "limit"),
			Cursor:    stringValue(input, "cursor"),
		})
		return pushLuaResult(L, sandbox, result, err)
	}
}

func (s *Service) luaGetTicket(ctx context.Context, sandbox *luasandbox.Sandbox, job Job) lua.LGFunction {
	return func(L *lua.LState) int {
		if s.tracker == nil {
			return pushLuaError(L, sandbox, "rayboard.get_ticket is not configured")
		}
		input, ok := tableArg(L, sandbox, 1)
		if !ok {
			return pushLuaError(L, sandbox, "rayboard.get_ticket expects a table")
		}
		ticket, err := s.tracker.GetTicket(ctx, cronPrincipal(job), ticketIDValue(input))
		return pushLuaResult(L, sandbox, ticket, err)
	}
}

func (s *Service) luaCreateTicket(ctx context.Context, sandbox *luasandbox.Sandbox, job Job) lua.LGFunction {
	return func(L *lua.LState) int {
		if s.tracker == nil {
			return pushLuaError(L, sandbox, "rayboard.create_ticket is not configured")
		}
		input, ok := tableArg(L, sandbox, 1)
		if !ok {
			return pushLuaError(L, sandbox, "rayboard.create_ticket expects a table")
		}
		customFields, ok := customFieldsValue(input)
		if !ok {
			return pushLuaError(L, sandbox, "custom_fields must be a table")
		}
		labels, ok := stringSliceValue(input, "labels")
		if !ok {
			return pushLuaError(L, sandbox, "labels must be an array of strings")
		}
		ticket, err := s.tracker.CreateTicket(ctx, cronPrincipal(job), tracker.CreateTicketInput{
			ProjectID:      stringValue(input, "project_id"),
			Title:          stringValue(input, "title"),
			Description:    stringValue(input, "description"),
			Status:         stringValue(input, "status"),
			Priority:       stringValue(input, "priority"),
			Type:           stringValue(input, "type"),
			ReporterID:     stringValue(input, "reporter_id"),
			AssigneeID:     stringValue(input, "assignee_id"),
			ParentTicketID: stringValue(input, "parent_ticket_id"),
			SprintID:       stringValue(input, "sprint_id"),
			ComponentID:    stringValue(input, "component_id"),
			VersionID:      stringValue(input, "version_id"),
			Rank:           stringValue(input, "rank"),
			StoryPoints:    storyPointsValue(input, "story_points"),
			Labels:         labels,
			CustomFields:   customFields,
		})
		return pushLuaResult(L, sandbox, ticket, err)
	}
}

func (s *Service) luaUpdateTicket(ctx context.Context, sandbox *luasandbox.Sandbox, job Job) lua.LGFunction {
	return func(L *lua.LState) int {
		if s.tracker == nil {
			return pushLuaError(L, sandbox, "rayboard.update_ticket is not configured")
		}
		input, ok := tableArg(L, sandbox, 1)
		if !ok {
			return pushLuaError(L, sandbox, "rayboard.update_ticket expects a table")
		}
		customFields, hasCustomFields, ok := optionalCustomFieldsValue(input)
		if !ok {
			return pushLuaError(L, sandbox, "custom_fields must be a table")
		}
		labels, hasLabels, ok := optionalStringSliceValue(input, "labels")
		if !ok {
			return pushLuaError(L, sandbox, "labels must be an array of strings")
		}
		update := tracker.UpdateTicketInput{
			Title:          optionalString(input, "title"),
			Description:    optionalString(input, "description"),
			Status:         optionalString(input, "status"),
			Priority:       optionalString(input, "priority"),
			Type:           optionalString(input, "type"),
			AssigneeID:     optionalString(input, "assignee_id"),
			ParentTicketID: optionalString(input, "parent_ticket_id"),
			SprintID:       optionalString(input, "sprint_id"),
			ComponentID:    optionalString(input, "component_id"),
			VersionID:      optionalString(input, "version_id"),
			Rank:           optionalString(input, "rank"),
			StoryPoints:    storyPointsValue(input, "story_points"),
			StoryPointsSet: hasKey(input, "story_points"),
		}
		if hasCustomFields {
			update.CustomFields = &customFields
		}
		if hasLabels {
			update.Labels = &labels
		}
		ticket, err := s.tracker.UpdateTicket(ctx, cronPrincipal(job), ticketIDValue(input), update)
		return pushLuaResult(L, sandbox, ticket, err)
	}
}

func (s *Service) luaComment(ctx context.Context, sandbox *luasandbox.Sandbox, job Job) lua.LGFunction {
	return func(L *lua.LState) int {
		if s.comments == nil {
			return pushLuaError(L, sandbox, "rayboard.comment is not configured")
		}
		input, ok := tableArg(L, sandbox, 1)
		if !ok {
			return pushLuaError(L, sandbox, "rayboard.comment expects a table")
		}
		comment, err := s.comments.Create(ctx, cronPrincipal(job), comments.CreateInput{
			TicketID: stringValue(input, "ticket_id"),
			Body:     stringValue(input, "body"),
		})
		return pushLuaResult(L, sandbox, comment, err)
	}
}

func cronPrincipal(job Job) authz.Principal {
	return authz.Principal{
		UserID:      job.OwnerUserID,
		ActorUserID: job.OwnerUserID,
		AuthKind:    authz.AuthKindCron,
	}
}

func tableArg(L *lua.LState, sandbox *luasandbox.Sandbox, index int) (map[string]any, bool) {
	table, ok := L.Get(index).(*lua.LTable)
	if !ok {
		return nil, false
	}
	value, err := sandbox.JSON.ToGo(table)
	if err != nil {
		return nil, false
	}
	input, ok := value.(map[string]any)
	return input, ok
}

func pushLuaResult(L *lua.LState, sandbox *luasandbox.Sandbox, value any, err error) int {
	if err != nil {
		return pushLuaError(L, sandbox, err.Error())
	}
	luaValue, err := toLuaValue(sandbox, value)
	if err != nil {
		return pushLuaError(L, sandbox, err.Error())
	}
	L.Push(luaValue)
	L.Push(lua.LNil)
	return 2
}

func pushLuaError(L *lua.LState, sandbox *luasandbox.Sandbox, message string) int {
	errorValue, err := sandbox.JSON.FromGo(map[string]any{"message": message})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(message))
		return 2
	}
	L.Push(lua.LNil)
	L.Push(errorValue)
	return 2
}

func toLuaValue(sandbox *luasandbox.Sandbox, value any) (lua.LValue, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return lua.LNil, fmt.Errorf("encode Lua helper result: %w", err)
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return lua.LNil, fmt.Errorf("decode Lua helper result: %w", err)
	}
	return sandbox.JSON.FromGo(decoded)
}

func stringValue(input map[string]any, key string) string {
	value, ok := input[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func ticketIDValue(input map[string]any) string {
	if ticketID := stringValue(input, "ticket_id"); ticketID != "" {
		return ticketID
	}
	return stringValue(input, "id")
}

func customFieldsValue(input map[string]any) (map[string]any, bool) {
	customFields, hasCustomFields, ok := optionalCustomFieldsValue(input)
	if !ok || !hasCustomFields {
		return nil, ok
	}
	return customFields, true
}

func optionalCustomFieldsValue(input map[string]any) (map[string]any, bool, bool) {
	value, ok := input["custom_fields"]
	if !ok || value == nil {
		return nil, false, true
	}
	customFields, ok := value.(map[string]any)
	return customFields, true, ok
}

func optionalString(input map[string]any, key string) *string {
	if _, ok := input[key]; !ok {
		return nil
	}
	value := stringValue(input, key)
	return &value
}

func hasKey(input map[string]any, key string) bool {
	_, ok := input[key]
	return ok
}

func storyPointsValue(input map[string]any, key string) *float64 {
	value, ok := input[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case float64:
		return &typed
	case int:
		result := float64(typed)
		return &result
	case int64:
		result := float64(typed)
		return &result
	default:
		return nil
	}
}

func stringSliceValue(input map[string]any, key string) ([]string, bool) {
	values, _, ok := optionalStringSliceValue(input, key)
	return values, ok
}

func optionalStringSliceValue(input map[string]any, key string) ([]string, bool, bool) {
	value, exists := input[key]
	if !exists || value == nil {
		return nil, false, true
	}
	items, ok := value.([]any)
	if !ok {
		return nil, true, false
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, true, false
		}
		values = append(values, text)
	}
	return values, true, true
}

func intValue(input map[string]any, key string) int {
	value, ok := input[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func sortSpecs(value any) []search.SortSpec {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	specs := make([]search.SortSpec, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		specs = append(specs, search.SortSpec{
			Field:     stringValue(itemMap, "field"),
			Direction: stringValue(itemMap, "direction"),
		})
	}
	return specs
}
