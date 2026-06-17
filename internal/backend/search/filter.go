package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/timo-42/rayboard/internal/backend/authz"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type wherePart struct {
	SQL  string
	Args []any
}

type compiledFilter struct {
	Parts []wherePart
}

type ticketFilterField struct {
	Name        string
	SQL         string
	Kind        string
	Normalize   func(string) string
}

type filterValue struct {
	Kind  string
	Value any
}

type filterTranslator struct {
	principal authz.Principal
	now       time.Time
}

const (
	filterKindString = "string"
	filterKindNumber = "number"
	filterKindBool   = "bool"
	filterKindList   = "list"
)

var ticketFilterFields = map[string]ticketFilterField{
	"project":          {Name: "project", SQL: "p.key", Kind: filterKindString, Normalize: strings.ToUpper},
	"project_id":       {Name: "project_id", SQL: "t.project_id", Kind: filterKindString},
	"key":              {Name: "key", SQL: "t.key", Kind: filterKindString, Normalize: strings.ToUpper},
	"title":            {Name: "title", SQL: "t.title", Kind: filterKindString},
	"status":           {Name: "status", SQL: "t.status", Kind: filterKindString, Normalize: strings.ToLower},
	"priority":         {Name: "priority", SQL: "t.priority", Kind: filterKindString, Normalize: strings.ToLower},
	"type":             {Name: "type", SQL: "t.type", Kind: filterKindString, Normalize: strings.ToLower},
	"reporter_id":      {Name: "reporter_id", SQL: "t.reporter_id", Kind: filterKindString},
	"assignee_id":      {Name: "assignee_id", SQL: "t.assignee_id", Kind: filterKindString},
	"parent_ticket_id": {Name: "parent_ticket_id", SQL: "t.parent_ticket_id", Kind: filterKindString},
	"sprint_id":        {Name: "sprint_id", SQL: "t.sprint_id", Kind: filterKindString},
	"component_id":     {Name: "component_id", SQL: "t.component_id", Kind: filterKindString},
	"version_id":       {Name: "version_id", SQL: "t.version_id", Kind: filterKindString},
	"labels":           {Name: "labels", Kind: filterKindList, Normalize: strings.ToLower},
	"start_date":       {Name: "start_date", SQL: "t.start_date", Kind: filterKindString},
	"due_date":         {Name: "due_date", SQL: "t.due_date", Kind: filterKindString},
	"created_at":       {Name: "created_at", SQL: "t.created_at", Kind: filterKindString},
	"updated_at":       {Name: "updated_at", SQL: "t.updated_at", Kind: filterKindString},
}

func compileTicketFilter(input string, principal authz.Principal, now time.Time) (compiledFilter, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return compiledFilter{}, nil
	}
	if len(input) > maxFilterLength {
		return compiledFilter{}, fmt.Errorf("must be %d characters or fewer", maxFilterLength)
	}

	env, err := cel.NewEnv(
		cel.Variable("project", cel.StringType),
		cel.Variable("project_id", cel.StringType),
		cel.Variable("key", cel.StringType),
		cel.Variable("title", cel.StringType),
		cel.Variable("status", cel.StringType),
		cel.Variable("priority", cel.StringType),
		cel.Variable("type", cel.StringType),
		cel.Variable("reporter_id", cel.StringType),
		cel.Variable("assignee_id", cel.StringType),
		cel.Variable("parent_ticket_id", cel.StringType),
		cel.Variable("sprint_id", cel.StringType),
		cel.Variable("component_id", cel.StringType),
		cel.Variable("version_id", cel.StringType),
		cel.Variable("labels", cel.DynType),
		cel.Variable("start_date", cel.StringType),
		cel.Variable("due_date", cel.StringType),
		cel.Variable("created_at", cel.StringType),
		cel.Variable("updated_at", cel.StringType),
		cel.Variable("custom", cel.DynType),
		cel.Function("currentUser", cel.Overload("rayboard_current_user", []*cel.Type{}, cel.StringType)),
		cel.Function("today", cel.Overload("rayboard_today", []*cel.Type{}, cel.StringType)),
		cel.Function("now", cel.Overload("rayboard_now", []*cel.Type{}, cel.StringType)),
	)
	if err != nil {
		return compiledFilter{}, fmt.Errorf("initialize CEL environment: %w", err)
	}
	ast, issues := env.Compile(input)
	if issues != nil && issues.Err() != nil {
		return compiledFilter{}, fmt.Errorf("invalid CEL filter: %w", issues.Err())
	}
	if ast.OutputType() != cel.BoolType {
		return compiledFilter{}, fmt.Errorf("filter expression must return bool")
	}

	translator := filterTranslator{principal: principal, now: now.UTC()}
	part, err := translator.expr(ast.Expr())
	if err != nil {
		return compiledFilter{}, err
	}
	return compiledFilter{Parts: []wherePart{part}}, nil
}

func (t filterTranslator) expr(expr *exprpb.Expr) (wherePart, error) {
	call := expr.GetCallExpr()
	if call == nil {
		return wherePart{}, fmt.Errorf("unsupported filter expression")
	}
	args := call.GetArgs()
	switch call.GetFunction() {
	case operators.LogicalAnd, operators.LogicalOr:
		if len(args) != 2 {
			return wherePart{}, fmt.Errorf("invalid logical expression")
		}
		left, err := t.expr(args[0])
		if err != nil {
			return wherePart{}, err
		}
		right, err := t.expr(args[1])
		if err != nil {
			return wherePart{}, err
		}
		joiner := " AND "
		if call.GetFunction() == operators.LogicalOr {
			joiner = " OR "
		}
		return wherePart{
			SQL:  "(" + left.SQL + joiner + right.SQL + ")",
			Args: append(left.Args, right.Args...),
		}, nil
	case operators.Equals, operators.NotEquals, operators.Less, operators.LessEquals, operators.Greater, operators.GreaterEquals:
		if len(args) != 2 {
			return wherePart{}, fmt.Errorf("invalid comparison expression")
		}
		return t.comparison(call.GetFunction(), args[0], args[1])
	case operators.In, operators.OldIn:
		if len(args) != 2 {
			return wherePart{}, fmt.Errorf("invalid list membership expression")
		}
		return t.inExpression(args[0], args[1])
	case "contains", "startsWith", "endsWith":
		return t.stringHelper(call.GetFunction(), call.GetTarget(), args)
	default:
		return wherePart{}, fmt.Errorf("unsupported function or operator %q", call.GetFunction())
	}
}

func (t filterTranslator) comparison(operator string, leftExpr *exprpb.Expr, rightExpr *exprpb.Expr) (wherePart, error) {
	if field, ok := fieldFromExpr(leftExpr); ok {
		value, err := t.value(rightExpr)
		if err != nil {
			return wherePart{}, err
		}
		return t.compareField(field, operator, value)
	}
	if field, ok := fieldFromExpr(rightExpr); ok {
		value, err := t.value(leftExpr)
		if err != nil {
			return wherePart{}, err
		}
		return t.compareField(field, reverseOperator(operator), value)
	}
	return wherePart{}, fmt.Errorf("comparison must include a supported ticket field")
}

func (t filterTranslator) compareField(field string, operator string, value filterValue) (wherePart, error) {
	if strings.HasPrefix(field, "custom.") {
		return t.compareCustomField(strings.TrimPrefix(field, "custom."), operator, value)
	}

	def, ok := ticketFilterFields[field]
	if !ok {
		return wherePart{}, fmt.Errorf("unsupported field %q", field)
	}
	if value.Kind == filterKindList {
		if operator != operators.Equals && operator != operators.NotEquals {
			return wherePart{}, fmt.Errorf("list comparisons only support == and !=")
		}
		return t.fieldInList(def, value, operator == operators.NotEquals)
	}
	if value.Kind != filterKindString {
		return wherePart{}, fmt.Errorf("field %q requires a string value", field)
	}
	text := value.Value.(string)
	if def.Normalize != nil {
		text = def.Normalize(text)
	}
	if field == "labels" {
		return labelClause(operator, text)
	}
	sqlOperator, err := sqlComparisonOperator(operator)
	if err != nil {
		return wherePart{}, err
	}
	return wherePart{SQL: def.SQL + " " + sqlOperator + " ?", Args: []any{text}}, nil
}

func (t filterTranslator) compareCustomField(key string, operator string, value filterValue) (wherePart, error) {
	if !validCustomFieldKey(key) {
		return wherePart{}, fmt.Errorf("unsupported custom field %q", key)
	}
	if value.Kind == filterKindList {
		if operator != operators.Equals && operator != operators.NotEquals {
			return wherePart{}, fmt.Errorf("custom field list comparisons only support == and !=")
		}
		return t.customInList(key, value, operator == operators.NotEquals)
	}

	sqlOperator, err := sqlComparisonOperator(operator)
	if err != nil {
		return wherePart{}, err
	}
	switch value.Kind {
	case filterKindString:
		return wherePart{
			SQL: `EXISTS (
				SELECT 1
				FROM ticket_custom_field_values cfv
				JOIN custom_field_definitions cfd ON cfd.id = cfv.field_id
				WHERE cfv.ticket_id = t.id
				  AND cfd.project_id = t.project_id
				  AND cfd.key = ?
				  AND cfd.field_type IN ('text', 'date', 'single_select', 'user')
				  AND json_extract(cfv.value_json, '$') ` + sqlOperator + ` ?
			)`,
			Args: []any{key, value.Value.(string)},
		}, nil
	case filterKindNumber:
		return wherePart{
			SQL: `EXISTS (
				SELECT 1
				FROM ticket_custom_field_values cfv
				JOIN custom_field_definitions cfd ON cfd.id = cfv.field_id
				WHERE cfv.ticket_id = t.id
				  AND cfd.project_id = t.project_id
				  AND cfd.key = ?
				  AND cfd.field_type = 'number'
				  AND CAST(json_extract(cfv.value_json, '$') AS REAL) ` + sqlOperator + ` ?
			)`,
			Args: []any{key, value.Value},
		}, nil
	case filterKindBool:
		if operator != operators.Equals && operator != operators.NotEquals {
			return wherePart{}, fmt.Errorf("boolean custom fields only support == and !=")
		}
		boolValue := 0
		if value.Value.(bool) {
			boolValue = 1
		}
		return wherePart{
			SQL: `EXISTS (
				SELECT 1
				FROM ticket_custom_field_values cfv
				JOIN custom_field_definitions cfd ON cfd.id = cfv.field_id
				WHERE cfv.ticket_id = t.id
				  AND cfd.project_id = t.project_id
				  AND cfd.key = ?
				  AND cfd.field_type = 'boolean'
				  AND json_extract(cfv.value_json, '$') ` + sqlOperator + ` ?
			)`,
			Args: []any{key, boolValue},
		}, nil
	default:
		return wherePart{}, fmt.Errorf("unsupported custom field value type")
	}
}

func (t filterTranslator) fieldInList(field ticketFilterField, value filterValue, negated bool) (wherePart, error) {
	values, ok := value.Value.([]filterValue)
	if !ok || len(values) == 0 {
		return wherePart{}, fmt.Errorf("list membership requires a non-empty literal list")
	}
	if field.Name == "labels" {
		parts := make([]string, 0, len(values))
		args := make([]any, 0, len(values))
		for _, item := range values {
			if item.Kind != filterKindString {
				return wherePart{}, fmt.Errorf("labels require string values")
			}
			parts = append(parts, `EXISTS (SELECT 1 FROM ticket_labels tl WHERE tl.ticket_id = t.id AND tl.label = ?)`)
			args = append(args, strings.ToLower(item.Value.(string)))
		}
		sql := "(" + strings.Join(parts, " OR ") + ")"
		if negated {
			sql = "NOT " + sql
		}
		return wherePart{SQL: sql, Args: args}, nil
	}

	placeholders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, item := range values {
		if item.Kind != filterKindString {
			return wherePart{}, fmt.Errorf("field %q requires string list values", field.Name)
		}
		text := item.Value.(string)
		if field.Normalize != nil {
			text = field.Normalize(text)
		}
		placeholders = append(placeholders, "?")
		args = append(args, text)
	}
	sql := field.SQL + " IN (" + strings.Join(placeholders, ", ") + ")"
	if negated {
		sql = field.SQL + " NOT IN (" + strings.Join(placeholders, ", ") + ")"
	}
	return wherePart{SQL: sql, Args: args}, nil
}

func (t filterTranslator) customInList(key string, value filterValue, negated bool) (wherePart, error) {
	if !validCustomFieldKey(key) {
		return wherePart{}, fmt.Errorf("unsupported custom field %q", key)
	}
	values := value.Value.([]filterValue)
	if len(values) == 0 {
		return wherePart{}, fmt.Errorf("custom field list comparisons require a non-empty literal list")
	}
	placeholders := make([]string, 0, len(values))
	args := []any{key}
	fieldTypes := map[string]struct{}{}
	for _, item := range values {
		switch item.Kind {
		case filterKindString:
			fieldTypes["'text'"] = struct{}{}
			fieldTypes["'date'"] = struct{}{}
			fieldTypes["'single_select'"] = struct{}{}
			fieldTypes["'user'"] = struct{}{}
			args = append(args, item.Value.(string))
		case filterKindNumber:
			fieldTypes["'number'"] = struct{}{}
			args = append(args, item.Value)
		default:
			return wherePart{}, fmt.Errorf("custom field list supports string and number values")
		}
		placeholders = append(placeholders, "?")
	}
	types := make([]string, 0, len(fieldTypes))
	for fieldType := range fieldTypes {
		types = append(types, fieldType)
	}
	sql := `EXISTS (
		SELECT 1
		FROM ticket_custom_field_values cfv
		JOIN custom_field_definitions cfd ON cfd.id = cfv.field_id
		WHERE cfv.ticket_id = t.id
		  AND cfd.project_id = t.project_id
		  AND cfd.key = ?
		  AND cfd.field_type IN (` + strings.Join(types, ", ") + `)
		  AND json_extract(cfv.value_json, '$') IN (` + strings.Join(placeholders, ", ") + `)
	)`
	if negated {
		sql = "NOT " + sql
	}
	return wherePart{SQL: sql, Args: args}, nil
}

func (t filterTranslator) inExpression(leftExpr *exprpb.Expr, rightExpr *exprpb.Expr) (wherePart, error) {
	if field, ok := fieldFromExpr(rightExpr); ok && field == "labels" {
		value, err := t.value(leftExpr)
		if err != nil {
			return wherePart{}, err
		}
		if value.Kind != filterKindString {
			return wherePart{}, fmt.Errorf("labels require string values")
		}
		return labelClause(operators.Equals, strings.ToLower(value.Value.(string)))
	}
	if field, ok := fieldFromExpr(leftExpr); ok {
		value, err := t.value(rightExpr)
		if err != nil {
			return wherePart{}, err
		}
		if value.Kind != filterKindList {
			return wherePart{}, fmt.Errorf("right side of in must be a literal list")
		}
		return t.compareField(field, operators.Equals, value)
	}
	return wherePart{}, fmt.Errorf("list membership must include a supported ticket field")
}

func (t filterTranslator) stringHelper(name string, target *exprpb.Expr, args []*exprpb.Expr) (wherePart, error) {
	if target == nil || len(args) != 1 {
		return wherePart{}, fmt.Errorf("%s requires one string argument", name)
	}
	field, ok := fieldFromExpr(target)
	if !ok {
		return wherePart{}, fmt.Errorf("%s target must be a supported ticket field", name)
	}
	def, ok := ticketFilterFields[field]
	if !ok || def.Kind != filterKindString {
		return wherePart{}, fmt.Errorf("%s only supports string ticket fields", name)
	}
	value, err := t.value(args[0])
	if err != nil {
		return wherePart{}, err
	}
	if value.Kind != filterKindString {
		return wherePart{}, fmt.Errorf("%s requires a string argument", name)
	}
	pattern := likePattern(name, value.Value.(string))
	return wherePart{SQL: def.SQL + " LIKE ? ESCAPE '\\'", Args: []any{pattern}}, nil
}

func (t filterTranslator) value(expr *exprpb.Expr) (filterValue, error) {
	if constant := expr.GetConstExpr(); constant != nil {
		switch kind := constant.GetConstantKind().(type) {
		case *exprpb.Constant_StringValue:
			return filterValue{Kind: filterKindString, Value: strings.TrimSpace(kind.StringValue)}, nil
		case *exprpb.Constant_Int64Value:
			return filterValue{Kind: filterKindNumber, Value: float64(kind.Int64Value)}, nil
		case *exprpb.Constant_Uint64Value:
			return filterValue{Kind: filterKindNumber, Value: float64(kind.Uint64Value)}, nil
		case *exprpb.Constant_DoubleValue:
			return filterValue{Kind: filterKindNumber, Value: kind.DoubleValue}, nil
		case *exprpb.Constant_BoolValue:
			return filterValue{Kind: filterKindBool, Value: kind.BoolValue}, nil
		default:
			return filterValue{}, fmt.Errorf("unsupported literal value")
		}
	}
	if list := expr.GetListExpr(); list != nil {
		values := make([]filterValue, 0, len(list.GetElements()))
		for _, element := range list.GetElements() {
			value, err := t.value(element)
			if err != nil {
				return filterValue{}, err
			}
			values = append(values, value)
		}
		return filterValue{Kind: filterKindList, Value: values}, nil
	}
	call := expr.GetCallExpr()
	if call != nil {
		switch call.GetFunction() {
		case "currentUser":
			if len(call.GetArgs()) != 0 {
				return filterValue{}, fmt.Errorf("currentUser() does not accept arguments")
			}
			if t.principal.UserID == "" {
				return filterValue{}, fmt.Errorf("currentUser() requires an authenticated user")
			}
			return filterValue{Kind: filterKindString, Value: t.principal.UserID}, nil
		case "today":
			if len(call.GetArgs()) != 0 {
				return filterValue{}, fmt.Errorf("today() does not accept arguments")
			}
			return filterValue{Kind: filterKindString, Value: t.now.Format("2006-01-02")}, nil
		case "now":
			if len(call.GetArgs()) != 0 {
				return filterValue{}, fmt.Errorf("now() does not accept arguments")
			}
			return filterValue{Kind: filterKindString, Value: t.now.Format(time.RFC3339Nano)}, nil
		}
	}
	return filterValue{}, fmt.Errorf("expected a literal value or approved function")
}

func fieldFromExpr(expr *exprpb.Expr) (string, bool) {
	if ident := expr.GetIdentExpr(); ident != nil {
		name := strings.ToLower(ident.GetName())
		_, ok := ticketFilterFields[name]
		return name, ok
	}
	if selectExpr := expr.GetSelectExpr(); selectExpr != nil {
		operand := selectExpr.GetOperand()
		if ident := operand.GetIdentExpr(); ident != nil && ident.GetName() == "custom" {
			field := strings.ToLower(selectExpr.GetField())
			if validCustomFieldKey(field) {
				return "custom." + field, true
			}
		}
	}
	return "", false
}

func labelClause(operator string, value string) (wherePart, error) {
	switch operator {
	case operators.Equals:
		return wherePart{
			SQL:  "EXISTS (SELECT 1 FROM ticket_labels tl WHERE tl.ticket_id = t.id AND tl.label = ?)",
			Args: []any{value},
		}, nil
	case operators.NotEquals:
		return wherePart{
			SQL:  "NOT EXISTS (SELECT 1 FROM ticket_labels tl WHERE tl.ticket_id = t.id AND tl.label = ?)",
			Args: []any{value},
		}, nil
	default:
		return wherePart{}, fmt.Errorf("labels only support ==, !=, and in")
	}
}

func sqlComparisonOperator(operator string) (string, error) {
	switch operator {
	case operators.Equals:
		return "=", nil
	case operators.NotEquals:
		return "!=", nil
	case operators.Less:
		return "<", nil
	case operators.LessEquals:
		return "<=", nil
	case operators.Greater:
		return ">", nil
	case operators.GreaterEquals:
		return ">=", nil
	default:
		return "", fmt.Errorf("unsupported comparison operator")
	}
}

func reverseOperator(operator string) string {
	switch operator {
	case operators.Less:
		return operators.Greater
	case operators.LessEquals:
		return operators.GreaterEquals
	case operators.Greater:
		return operators.Less
	case operators.GreaterEquals:
		return operators.LessEquals
	default:
		return operator
	}
}

func validCustomFieldKey(key string) bool {
	if key == "" || len(key) > 64 {
		return false
	}
	for index, r := range key {
		if index == 0 {
			if r < 'a' || r > 'z' {
				return false
			}
			continue
		}
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func likePattern(function string, value string) string {
	escaped := escapeLike(value)
	switch function {
	case "startsWith":
		return escaped + "%"
	case "endsWith":
		return "%" + escaped
	default:
		return "%" + escaped + "%"
	}
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

func normalizeFilterValue(field string, value string) string {
	value = strings.TrimSpace(value)
	if def, ok := ticketFilterFields[field]; ok && def.Normalize != nil {
		return def.Normalize(value)
	}
	return value
}
