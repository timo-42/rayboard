package search

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

const (
	defaultListLimit   = 50
	maxListLimit       = 200
	maxNameLength      = 200
	maxFilterLength    = 2000
	maxTextLength      = 500
	maxColumnCount     = 20
	maxSortCount       = 4
	maxSearchTermCount = 16
)

var columnPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

var ticketSortFields = map[string]string{
	"created_at": "t.created_at",
	"due_date":   "t.due_date",
	"updated_at": "t.updated_at",
	"key":        "t.key",
	"title":      "t.title",
	"status":     "t.status",
	"priority":   "t.priority",
	"start_date": "t.start_date",
}

var allowedColumns = map[string]struct{}{
	"key":              {},
	"title":            {},
	"description":      {},
	"status":           {},
	"priority":         {},
	"type":             {},
	"reporter_id":      {},
	"assignee_id":      {},
	"parent_ticket_id": {},
	"sprint_id":        {},
	"component_id":     {},
	"version_id":       {},
	"labels":           {},
	"start_date":       {},
	"due_date":         {},
	"rank":             {},
	"created_at":       {},
	"updated_at":       {},
}

type Service struct {
	db         *sql.DB
	repo       *Repository
	authorizer authz.Evaluator
	now        func() time.Time
}

type Option func(*Service)

func NewService(db *sql.DB, authorizer authz.Evaluator, options ...Option) *Service {
	service := &Service{
		db:         db,
		repo:       NewRepository(db),
		authorizer: authorizer,
		now:        func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithNow(now func() time.Time) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func (s *Service) CreateSavedView(ctx context.Context, principal authz.Principal, input CreateSavedViewInput) (SavedView, error) {
	view, err := s.buildSavedView(ctx, principal, input)
	if err != nil {
		return SavedView{}, err
	}
	if err := s.requireManageSavedView(ctx, principal, view); err != nil {
		return SavedView{}, err
	}
	if err := s.repo.CreateSavedView(ctx, view); err != nil {
		return SavedView{}, err
	}
	return view, nil
}

func (s *Service) GetSavedView(ctx context.Context, principal authz.Principal, id string) (SavedView, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return SavedView{}, validationFailed(map[string]string{"id": "Required"})
	}
	view, err := s.repo.GetSavedView(ctx, id)
	if err != nil {
		return SavedView{}, err
	}
	if err := s.requireReadSavedView(principal, view); err != nil {
		return SavedView{}, err
	}
	return view, nil
}

func (s *Service) ListSavedViews(ctx context.Context, principal authz.Principal, input ListSavedViewsInput) ([]SavedView, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	limit, offset, err := normalizeListWindow(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	if principal.UserID == "" || principal.Disabled {
		return nil, authz.ErrForbidden
	}

	var visibleProjectIDs []string
	if input.ProjectID != "" {
		if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(input.ProjectID)); err != nil {
			return nil, err
		}
		if err := s.requireProject(ctx, input.ProjectID); err != nil {
			return nil, err
		}
		visibleProjectIDs = []string{input.ProjectID}
	} else {
		visibleProjectIDs, err = s.visibleProjectIDs(ctx, principal)
		if err != nil {
			return nil, err
		}
	}

	return s.repo.ListSavedViews(ctx, savedViewListQuery{
		OwnerUserID:       principal.UserID,
		ProjectID:         input.ProjectID,
		VisibleProjectIDs: visibleProjectIDs,
		PinnedOnly:        input.Pinned,
		Limit:             limit,
		Offset:            offset,
	})
}

func (s *Service) UpdateSavedView(ctx context.Context, principal authz.Principal, id string, input UpdateSavedViewInput) (SavedView, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return SavedView{}, validationFailed(map[string]string{"id": "Required"})
	}
	current, err := s.repo.GetSavedView(ctx, id)
	if err != nil {
		return SavedView{}, err
	}
	if err := s.requireManageSavedView(ctx, principal, current); err != nil {
		return SavedView{}, err
	}

	next := current
	if input.Name != nil {
		next.Name = strings.TrimSpace(*input.Name)
	}
	if input.Query != nil {
		next.Query = normalizeSavedViewQuery(*input.Query)
	}
	if input.Sort != nil {
		next.Sort = cloneSortSpecs(*input.Sort)
	}
	if input.Columns != nil {
		next.Columns = cloneStrings(*input.Columns)
	}
	if input.DisplayMode != nil {
		next.DisplayMode = strings.TrimSpace(*input.DisplayMode)
	}
	if input.GroupBy != nil {
		next.GroupBy = strings.TrimSpace(*input.GroupBy)
	}
	if input.Pinned != nil {
		next.Pinned = *input.Pinned
	}
	next, err = s.normalizeAndValidateSavedViewConfig(principal, next)
	if err != nil {
		return SavedView{}, err
	}
	next.UpdatedAt = s.now().UTC()
	if err := s.repo.UpdateSavedView(ctx, next); err != nil {
		return SavedView{}, err
	}
	return next, nil
}

func (s *Service) DeleteSavedView(ctx context.Context, principal authz.Principal, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return validationFailed(map[string]string{"id": "Required"})
	}
	view, err := s.repo.GetSavedView(ctx, id)
	if err != nil {
		return err
	}
	if err := s.requireManageSavedView(ctx, principal, view); err != nil {
		return err
	}
	return s.repo.DeleteSavedView(ctx, id)
}

func (s *Service) SearchTickets(ctx context.Context, principal authz.Principal, input SearchTicketsInput) (SearchTicketsResult, error) {
	query, limit, offset, err := s.buildTicketSearchQuery(ctx, principal, input)
	if err != nil {
		return SearchTicketsResult{}, err
	}
	if len(query.ProjectIDs) == 0 {
		return SearchTicketsResult{}, nil
	}

	if query.FTSQuery != "" {
		if err := s.repo.RefreshFTSIndex(ctx); err != nil {
			return SearchTicketsResult{}, err
		}
	}

	query.Limit = limit + 1
	query.Offset = offset
	tickets, err := s.repo.SearchTickets(ctx, query)
	if err != nil {
		return SearchTicketsResult{}, err
	}

	nextCursor := ""
	if len(tickets) > limit {
		tickets = tickets[:limit]
		nextCursor = encodeCursor(offset + limit)
	}
	return SearchTicketsResult{Tickets: tickets, NextCursor: nextCursor}, nil
}

func (s *Service) buildSavedView(ctx context.Context, principal authz.Principal, input CreateSavedViewInput) (SavedView, error) {
	scopeType := strings.ToLower(strings.TrimSpace(input.ScopeType))
	if scopeType == "" {
		scopeType = SavedViewScopeUser
	}
	ownerUserID := strings.TrimSpace(input.OwnerUserID)
	if ownerUserID == "" {
		ownerUserID = principal.UserID
	}
	view := SavedView{
		OwnerUserID: ownerUserID,
		ProjectID:   strings.TrimSpace(input.ProjectID),
		ScopeType:   scopeType,
		Name:        strings.TrimSpace(input.Name),
		Query:       normalizeSavedViewQuery(input.Query),
		Sort:        cloneSortSpecs(input.Sort),
		Columns:     cloneStrings(input.Columns),
		DisplayMode: strings.TrimSpace(input.DisplayMode),
		GroupBy:     strings.TrimSpace(input.GroupBy),
		Pinned:      input.Pinned,
	}
	if err := s.validateSavedViewScope(ctx, principal, view); err != nil {
		return SavedView{}, err
	}
	var err error
	view, err = s.normalizeAndValidateSavedViewConfig(principal, view)
	if err != nil {
		return SavedView{}, err
	}

	id, err := newID("view")
	if err != nil {
		return SavedView{}, err
	}
	now := s.now().UTC()
	view.ID = id
	view.CreatedAt = now
	view.UpdatedAt = now
	return view, nil
}

func (s *Service) validateSavedViewScope(ctx context.Context, principal authz.Principal, view SavedView) error {
	fields := map[string]string{}
	if principal.UserID == "" || principal.Disabled {
		return authz.ErrForbidden
	}
	if view.OwnerUserID == "" {
		fields["owner_user_id"] = "Required"
	} else if view.OwnerUserID != principal.UserID {
		fields["owner_user_id"] = "Must be the current user"
	}
	switch view.ScopeType {
	case SavedViewScopeUser:
	case SavedViewScopeProject:
		if view.ProjectID == "" {
			fields["project_id"] = "Required for project views"
		}
	case SavedViewScopeGlobal:
		if view.ProjectID != "" {
			fields["project_id"] = "Must be empty for global views"
		}
	default:
		fields["scope_type"] = "Must be user, project, or global"
	}
	if len(fields) > 0 {
		return validationFailed(fields)
	}
	if view.ProjectID != "" {
		if err := s.requireProject(ctx, view.ProjectID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) normalizeAndValidateSavedViewConfig(principal authz.Principal, view SavedView) (SavedView, error) {
	fields := map[string]string{}
	view.Query = normalizeSavedViewQuery(view.Query)
	if view.Name == "" {
		fields["name"] = "Required"
	} else if len(view.Name) > maxNameLength {
		fields["name"] = fmt.Sprintf("Must be %d characters or fewer", maxNameLength)
	}
	if err := s.validateSavedViewQuery(view.Query, principal); err != nil {
		fields["query"] = err.Error()
	}
	view.DisplayMode = strings.ToLower(strings.TrimSpace(view.DisplayMode))
	if view.DisplayMode == "" {
		view.DisplayMode = SavedViewDisplayList
	}
	if !validSavedViewDisplayMode(view.DisplayMode) {
		fields["display_mode"] = "Must be list, board, or backlog"
	}
	view.GroupBy = strings.ToLower(strings.TrimSpace(view.GroupBy))
	if view.GroupBy != "" && !validSavedViewGroupBy(view.GroupBy) {
		fields["group_by"] = "Unsupported grouping field"
	}
	if view.Pinned && view.ScopeType != SavedViewScopeProject {
		fields["pinned"] = "Only project views can be pinned"
	}
	sort, err := normalizeSort(view.Sort)
	if err != nil {
		fields["sort"] = err.Error()
	} else {
		view.Sort = sort
	}
	columns, err := normalizeColumns(view.Columns)
	if err != nil {
		fields["columns"] = err.Error()
	} else {
		view.Columns = columns
	}
	if len(fields) > 0 {
		return SavedView{}, validationFailed(fields)
	}
	return view, nil
}

func validSavedViewDisplayMode(displayMode string) bool {
	switch displayMode {
	case SavedViewDisplayList, SavedViewDisplayBoard, SavedViewDisplayBacklog:
		return true
	default:
		return false
	}
}

func validSavedViewGroupBy(groupBy string) bool {
	switch groupBy {
	case "status", "assignee_id", "sprint_id", "component_id", "version_id", "priority", "type":
		return true
	default:
		return false
	}
}

func (s *Service) validateSavedViewQuery(query SavedViewQuery, principal authz.Principal) error {
	if len(query.Filter) > maxFilterLength {
		return fmt.Errorf("filter must be %d characters or fewer", maxFilterLength)
	}
	if strings.TrimSpace(query.Filter) != "" {
		if _, err := compileTicketFilter(query.Filter, principal, s.now().UTC()); err != nil {
			return err
		}
	}
	if len(query.Text) > maxTextLength {
		return fmt.Errorf("text must be %d characters or fewer", maxTextLength)
	}
	if strings.TrimSpace(query.Text) != "" {
		if _, err := buildFTSQuery(query.Text); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) buildTicketSearchQuery(ctx context.Context, principal authz.Principal, input SearchTicketsInput) (ticketSearchQuery, int, int, error) {
	fields := map[string]string{}
	projectID := strings.TrimSpace(input.ProjectID)
	filterText := strings.TrimSpace(input.Filter)
	text := strings.TrimSpace(input.Text)
	if len(filterText) > maxFilterLength {
		fields["filter"] = fmt.Sprintf("Must be %d characters or fewer", maxFilterLength)
	}
	if len(text) > maxTextLength {
		fields["text"] = fmt.Sprintf("Must be %d characters or fewer", maxTextLength)
	}

	filter, err := compileTicketFilter(filterText, principal, s.now().UTC())
	if err != nil {
		fields["filter"] = err.Error()
	}
	ftsQuery := ""
	if text != "" {
		ftsQuery, err = buildFTSQuery(text)
		if err != nil {
			fields["text"] = err.Error()
		}
	}
	sort, err := normalizeSort(input.Sort)
	if err != nil {
		fields["sort"] = err.Error()
	}
	limit, offset, err := decodeSearchWindow(input.Limit, input.Cursor)
	if err != nil {
		fields["cursor"] = err.Error()
	}
	if len(fields) > 0 {
		return ticketSearchQuery{}, 0, 0, validationFailed(fields)
	}

	projectIDs, err := s.authorizedSearchProjectIDs(ctx, principal, projectID)
	if err != nil {
		return ticketSearchQuery{}, 0, 0, err
	}
	return ticketSearchQuery{
		ProjectIDs: projectIDs,
		Filter:     filter,
		FTSQuery:   ftsQuery,
		Sort:       sort,
	}, limit, offset, nil
}

func (s *Service) authorizedSearchProjectIDs(ctx context.Context, principal authz.Principal, projectID string) ([]string, error) {
	if projectID != "" {
		if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
			return nil, err
		}
		if err := s.requireProject(ctx, projectID); err != nil {
			return nil, err
		}
		return []string{projectID}, nil
	}
	return s.visibleProjectIDs(ctx, principal)
}

func (s *Service) visibleProjectIDs(ctx context.Context, principal authz.Principal) ([]string, error) {
	if s == nil || s.authorizer == nil {
		return nil, errors.New("search: authorization evaluator is required")
	}
	projectIDs, err := s.repo.ListActiveProjectIDs(ctx)
	if err != nil {
		return nil, err
	}
	visible := make([]string, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		if s.authorizer.Can(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)) {
			visible = append(visible, projectID)
		}
	}
	return visible, nil
}

func (s *Service) requireReadSavedView(principal authz.Principal, view SavedView) error {
	if principal.UserID == "" || principal.Disabled {
		return authz.ErrForbidden
	}
	switch view.ScopeType {
	case SavedViewScopeUser:
		if view.OwnerUserID == principal.UserID {
			return nil
		}
		return authz.ErrForbidden
	case SavedViewScopeProject:
		return s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(view.ProjectID))
	case SavedViewScopeGlobal:
		return nil
	default:
		return authz.ErrForbidden
	}
}

func (s *Service) requireManageSavedView(ctx context.Context, principal authz.Principal, view SavedView) error {
	if principal.UserID == "" || principal.Disabled {
		return authz.ErrForbidden
	}
	switch view.ScopeType {
	case SavedViewScopeUser:
		if view.OwnerUserID == principal.UserID {
			if view.ProjectID != "" {
				if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(view.ProjectID)); err != nil {
					return err
				}
			}
			return nil
		}
		return authz.ErrForbidden
	case SavedViewScopeProject:
		if err := s.require(principal, authz.PermissionViewsManage, authz.ProjectScope(view.ProjectID)); err != nil {
			return err
		}
		return s.requireProject(ctx, view.ProjectID)
	case SavedViewScopeGlobal:
		return s.require(principal, authz.PermissionViewsManage, authz.GlobalScope())
	default:
		return validationFailed(map[string]string{"scope_type": "Must be user, project, or global"})
	}
}

func (s *Service) requireProject(ctx context.Context, projectID string) error {
	exists, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return notFound("project", projectID)
	}
	return nil
}

func (s *Service) require(principal authz.Principal, permission authz.Permission, scope authz.Scope) error {
	if s == nil || s.authorizer == nil {
		return errors.New("search: authorization evaluator is required")
	}
	return s.authorizer.Require(principal, permission, scope)
}

func normalizeListWindow(limit int, offset int) (int, int, error) {
	fields := map[string]string{}
	if limit < 0 {
		fields["limit"] = "Must be non-negative"
	}
	if limit > maxListLimit {
		fields["limit"] = fmt.Sprintf("Must be %d or fewer", maxListLimit)
	}
	if offset < 0 {
		fields["offset"] = "Must be non-negative"
	}
	if len(fields) > 0 {
		return 0, 0, validationFailed(fields)
	}
	if limit == 0 {
		limit = defaultListLimit
	}
	return limit, offset, nil
}

func decodeSearchWindow(limit int, cursor string) (int, int, error) {
	if limit < 0 {
		return 0, 0, fmt.Errorf("limit must be non-negative")
	}
	if limit > maxListLimit {
		return 0, 0, fmt.Errorf("limit must be %d or fewer", maxListLimit)
	}
	if limit == 0 {
		limit = defaultListLimit
	}
	offset, err := decodeCursor(cursor)
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

func normalizeSavedViewQuery(query SavedViewQuery) SavedViewQuery {
	return SavedViewQuery{
		Filter: strings.TrimSpace(query.Filter),
		Text:   strings.TrimSpace(query.Text),
	}
}

func normalizeSort(input []SortSpec) ([]SortSpec, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if len(input) > maxSortCount {
		return nil, fmt.Errorf("must include %d fields or fewer", maxSortCount)
	}
	normalized := make([]SortSpec, 0, len(input))
	seen := map[string]struct{}{}
	for _, spec := range input {
		field := strings.ToLower(strings.TrimSpace(spec.Field))
		if _, ok := ticketSortFields[field]; !ok {
			return nil, fmt.Errorf("unsupported sort field %q", spec.Field)
		}
		if _, ok := seen[field]; ok {
			return nil, fmt.Errorf("duplicate sort field %q", field)
		}
		seen[field] = struct{}{}
		direction := strings.ToLower(strings.TrimSpace(spec.Direction))
		if direction == "" {
			direction = SortDirectionAsc
		}
		if direction != SortDirectionAsc && direction != SortDirectionDesc {
			return nil, fmt.Errorf("sort direction for %q must be asc or desc", field)
		}
		normalized = append(normalized, SortSpec{Field: field, Direction: direction})
	}
	return normalized, nil
}

func normalizeColumns(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if len(input) > maxColumnCount {
		return nil, fmt.Errorf("must include %d columns or fewer", maxColumnCount)
	}
	normalized := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, column := range input {
		column = strings.ToLower(strings.TrimSpace(column))
		if column == "" {
			return nil, fmt.Errorf("column names must be non-empty")
		}
		if !columnPattern.MatchString(column) {
			return nil, fmt.Errorf("invalid column %q", column)
		}
		if _, ok := allowedColumns[column]; !ok {
			return nil, fmt.Errorf("unsupported column %q", column)
		}
		if _, ok := seen[column]; ok {
			continue
		}
		seen[column] = struct{}{}
		normalized = append(normalized, column)
	}
	return normalized, nil
}

func buildFTSQuery(text string) (string, error) {
	if len(text) > maxTextLength {
		return "", fmt.Errorf("must be %d characters or fewer", maxTextLength)
	}
	var tokens []string
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(current.String()))
		current.Reset()
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			current.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	if len(tokens) == 0 {
		return "", fmt.Errorf("must contain searchable terms")
	}
	if len(tokens) > maxSearchTermCount {
		tokens = tokens[:maxSearchTermCount]
	}
	quoted := make([]string, len(tokens))
	for i, token := range tokens {
		quoted[i] = `"` + token + `"`
	}
	return strings.Join(quoted, " "), nil
}

func encodeCursor(offset int) string {
	if offset <= 0 {
		return ""
	}
	data, _ := json.Marshal(struct {
		Offset int `json:"offset"`
	}{Offset: offset})
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(cursor string) (int, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return 0, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor")
	}
	var payload struct {
		Offset int `json:"offset"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, fmt.Errorf("invalid cursor")
	}
	if payload.Offset < 0 {
		return 0, fmt.Errorf("invalid cursor")
	}
	return payload.Offset, nil
}

func cloneSortSpecs(input []SortSpec) []SortSpec {
	if input == nil {
		return nil
	}
	cloned := make([]SortSpec, len(input))
	copy(cloned, input)
	return cloned
}

func cloneStrings(input []string) []string {
	if input == nil {
		return nil
	}
	cloned := make([]string, len(input))
	copy(cloned, input)
	return cloned
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
