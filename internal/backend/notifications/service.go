package notifications

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
)

var (
	ErrNotFound   = errors.New("notifications: not found")
	ErrValidation = errors.New("notifications: validation failed")
	ErrDelivery   = errors.New("notifications: delivery failed")
)

type Notification struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Type        string         `json:"type"`
	SubjectType string         `json:"subject_type,omitempty"`
	SubjectID   string         `json:"subject_id,omitempty"`
	Body        string         `json:"body"`
	Data        map[string]any `json:"data"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type CreateInput struct {
	UserID      string
	Type        string
	SubjectType string
	SubjectID   string
	Body        string
	Data        map[string]any
}

type ListInput struct {
	UnreadOnly bool
	Limit      int
	Offset     int
}

type Service struct {
	db         *sql.DB
	now        func() time.Time
	eventStore *events.Store
	runs       *automation.RunStore
	openrouter *openrouter.Service
}

type Option func(*Service)

func NewService(db *sql.DB, options ...Option) *Service {
	service := &Service{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
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

func WithEventStore(store *events.Store) Option {
	return func(service *Service) {
		service.eventStore = store
	}
}

func WithRunStore(runStore *automation.RunStore) Option {
	return func(service *Service) {
		service.runs = runStore
	}
}

func WithOpenRouterService(openRouterService *openrouter.Service) Option {
	return func(service *Service) {
		service.openrouter = openRouterService
	}
}

func (s *Service) RegisterEventHandlers(bus *events.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe("comment.created", s.handleCommentCreated)
	bus.Subscribe("comment.mentioned", s.handleCommentMentioned)
	bus.Subscribe("ticket.updated", s.handleTicketUpdated)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Notification, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Type = strings.TrimSpace(input.Type)
	input.SubjectType = strings.TrimSpace(input.SubjectType)
	input.SubjectID = strings.TrimSpace(input.SubjectID)
	input.Body = strings.TrimSpace(input.Body)
	fields := map[string]string{}
	if input.UserID == "" {
		fields["user_id"] = "Required"
	}
	if input.Type == "" {
		fields["type"] = "Required"
	}
	if input.Body == "" {
		fields["body"] = "Required"
	}
	if len(input.Body) > 2000 {
		fields["body"] = "Must be 2000 characters or fewer"
	}
	if len(fields) > 0 {
		return Notification{}, fmt.Errorf("%w: invalid notification", ErrValidation)
	}

	id, err := newID("notif")
	if err != nil {
		return Notification{}, err
	}
	now := s.now().UTC()
	dataJSON, err := marshalData(input.Data)
	if err != nil {
		return Notification{}, err
	}
	notification := Notification{
		ID:          id,
		UserID:      input.UserID,
		Type:        input.Type,
		SubjectType: input.SubjectType,
		SubjectID:   input.SubjectID,
		Body:        input.Body,
		Data:        nonNilMap(input.Data),
		CreatedAt:   now,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, subject_type, subject_id, body, data_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, notification.ID, notification.UserID, notification.Type, nullableString(notification.SubjectType), nullableString(notification.SubjectID), notification.Body, dataJSON, formatTime(notification.CreatedAt)); err != nil {
		return Notification{}, fmt.Errorf("insert notification: %w", err)
	}
	return notification, nil
}

func (s *Service) List(ctx context.Context, principal authz.Principal, input ListInput) ([]Notification, error) {
	if principal.UserID == "" || principal.Disabled {
		return nil, authz.ErrForbidden
	}
	limit, offset, err := normalizeList(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	where := []string{"user_id = ?"}
	args := []any{principal.UserID}
	if input.UnreadOnly {
		where = append(where, "read_at IS NULL")
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, type, subject_type, subject_id, body, data_json, read_at, created_at
		FROM notifications
		WHERE `+joinAnd(where)+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return notifications, nil
}

func (s *Service) SetRead(ctx context.Context, principal authz.Principal, notificationID string, read bool) (Notification, error) {
	if principal.UserID == "" || principal.Disabled {
		return Notification{}, authz.ErrForbidden
	}
	var readAt any
	if read {
		readAt = formatTime(s.now().UTC())
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = ?
		WHERE id = ? AND user_id = ?
	`, readAt, notificationID, principal.UserID)
	if err != nil {
		return Notification{}, fmt.Errorf("update notification read state: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Notification{}, fmt.Errorf("check notification update: %w", err)
	}
	if affected == 0 {
		return Notification{}, ErrNotFound
	}
	return s.getForUser(ctx, principal.UserID, notificationID)
}

func (s *Service) MarkAllRead(ctx context.Context, principal authz.Principal) error {
	if principal.UserID == "" || principal.Disabled {
		return authz.ErrForbidden
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, ?)
		WHERE user_id = ? AND read_at IS NULL
	`, formatTime(s.now().UTC()), principal.UserID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

func (s *Service) ProcessPendingDomainEvents(ctx context.Context, limit int) (int, error) {
	if s == nil || s.eventStore == nil {
		return 0, nil
	}
	pending, err := s.eventStore.ListPending(ctx, limit, "comment.created", "comment.mentioned", "ticket.updated")
	if err != nil {
		return 0, err
	}

	processed := 0
	var firstErr error
	for _, stored := range pending {
		event := events.Event{
			Type:        stored.Type,
			ActorID:     stored.ActorID,
			ProjectID:   stored.ProjectID,
			ObjectID:    stored.ObjectID,
			SubjectType: stored.SubjectType,
			SubjectID:   stored.SubjectID,
			RelatedType: stored.RelatedType,
			RelatedID:   stored.RelatedID,
			At:          stored.At,
			Data:        stored.Data,
		}
		if err := s.handleDomainEvent(ctx, event); err != nil {
			if markErr := s.eventStore.MarkFailed(ctx, stored.ID, err); markErr != nil {
				return processed, markErr
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := s.enqueuePolicyDeliveriesForEvent(ctx, stored.ID, event); err != nil {
			if markErr := s.eventStore.MarkFailed(ctx, stored.ID, err); markErr != nil {
				return processed, markErr
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := s.eventStore.MarkProcessed(ctx, stored.ID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, firstErr
}

func (s *Service) handleDomainEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case "comment.created":
		return s.handleCommentCreated(ctx, event)
	case "comment.mentioned":
		return s.handleCommentMentioned(ctx, event)
	case "ticket.updated":
		return s.handleTicketUpdated(ctx, event)
	default:
		return nil
	}
}

func (s *Service) enqueuePolicyDeliveriesForEvent(ctx context.Context, domainEventID string, event events.Event) error {
	plans, err := s.externalNotificationPlans(ctx, event)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		policies, err := s.matchingPolicies(ctx, plan.EventType, plan.ProjectID)
		if err != nil {
			return err
		}
		for _, policy := range policies {
			hooked, err := s.applyNotificationHooks(ctx, policy, plan)
			if err != nil {
				return err
			}
			if hooked.Suppressed {
				continue
			}
			if strings.TrimSpace(hooked.Message) == "" || len(hooked.Message) > 4000 {
				return fmt.Errorf("%w: invalid hooked notification message", ErrValidation)
			}
			if err := s.validatePolicyDestinations(ctx, policy.ScopeType, policy.ProjectID, hooked.DestinationIDs); err != nil {
				return err
			}
			for _, destinationID := range hooked.DestinationIDs {
				if _, err := s.EnqueueDelivery(ctx, EnqueueDeliveryInput{
					DomainEventID:  domainEventID,
					IdempotencyKey: deliveryIdempotencyKey(domainEventID, policy.ID, destinationID, hooked.EventType),
					PolicyID:       policy.ID,
					DestinationID:  destinationID,
					EventType:      hooked.EventType,
					SubjectType:    hooked.SubjectType,
					SubjectID:      hooked.SubjectID,
					Message:        hooked.Message,
					Payload:        hooked.Payload,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type externalNotificationPlan struct {
	EventType   string
	ProjectID   string
	SubjectType string
	SubjectID   string
	Message     string
	Payload     map[string]any
}

func (s *Service) externalNotificationPlans(ctx context.Context, event events.Event) ([]externalNotificationPlan, error) {
	switch event.Type {
	case "comment.created":
		ticketID, _ := event.Data["ticket_id"].(string)
		if ticketID == "" {
			return nil, nil
		}
		ticket, err := s.ticket(ctx, ticketID)
		if err != nil {
			return nil, err
		}
		return []externalNotificationPlan{{
			EventType:   "comment_added",
			ProjectID:   eventProjectID(event, ticket.ProjectID),
			SubjectType: "ticket",
			SubjectID:   ticket.ID,
			Message:     fmt.Sprintf("New comment on %s", ticket.Key),
			Payload: map[string]any{
				"ticket_id":  ticket.ID,
				"ticket_key": ticket.Key,
				"comment_id": event.ObjectID,
			},
		}}, nil
	case "comment.mentioned":
		ticketID, _ := event.Data["ticket_id"].(string)
		if ticketID == "" {
			return nil, nil
		}
		ticket, err := s.ticket(ctx, ticketID)
		if err != nil {
			return nil, err
		}
		mentionedUserID, _ := event.Data["mentioned_user_id"].(string)
		mentionedUsername, _ := event.Data["mentioned_username"].(string)
		return []externalNotificationPlan{{
			EventType:   "comment_mentioned",
			ProjectID:   eventProjectID(event, ticket.ProjectID),
			SubjectType: "ticket",
			SubjectID:   ticket.ID,
			Message:     fmt.Sprintf("%s mentioned %s on %s", displayActor(event.ActorID), displayMention(mentionedUsername, mentionedUserID), ticket.Key),
			Payload: map[string]any{
				"ticket_id":          ticket.ID,
				"ticket_key":         ticket.Key,
				"comment_id":         eventCommentID(event),
				"mentioned_user_id":  mentionedUserID,
				"mentioned_username": mentionedUsername,
				"actor_user_id":      event.ActorID,
			},
		}}, nil
	case "ticket.updated":
		ticket, err := s.ticket(ctx, event.ObjectID)
		if err != nil {
			return nil, err
		}
		projectID := eventProjectID(event, ticket.ProjectID)
		plans := []externalNotificationPlan{}
		if newAssignee := changeNew(event.Data, "assignee_id"); newAssignee != "" {
			plans = append(plans, externalNotificationPlan{
				EventType:   "ticket_assigned",
				ProjectID:   projectID,
				SubjectType: "ticket",
				SubjectID:   ticket.ID,
				Message:     fmt.Sprintf("Assigned to %s", ticket.Key),
				Payload: map[string]any{
					"ticket_id":     ticket.ID,
					"ticket_key":    ticket.Key,
					"assignee_id":   newAssignee,
					"actor_user_id": event.ActorID,
				},
			})
		}
		if status := changeNew(event.Data, "status"); status != "" {
			plans = append(plans, externalNotificationPlan{
				EventType:   "ticket_status_changed",
				ProjectID:   projectID,
				SubjectType: "ticket",
				SubjectID:   ticket.ID,
				Message:     fmt.Sprintf("%s moved to %s", ticket.Key, status),
				Payload: map[string]any{
					"ticket_id":     ticket.ID,
					"ticket_key":    ticket.Key,
					"status":        status,
					"actor_user_id": event.ActorID,
				},
			})
		}
		if sprintID := changeNew(event.Data, "sprint_id"); sprintID != "" {
			plans = append(plans, externalNotificationPlan{
				EventType:   "sprint_changed",
				ProjectID:   projectID,
				SubjectType: "ticket",
				SubjectID:   ticket.ID,
				Message:     fmt.Sprintf("%s sprint changed", ticket.Key),
				Payload: map[string]any{
					"ticket_id":     ticket.ID,
					"ticket_key":    ticket.Key,
					"sprint_id":     sprintID,
					"actor_user_id": event.ActorID,
				},
			})
		}
		if versionID := changeNew(event.Data, "version_id"); versionID != "" {
			plans = append(plans, externalNotificationPlan{
				EventType:   "release_changed",
				ProjectID:   projectID,
				SubjectType: "ticket",
				SubjectID:   ticket.ID,
				Message:     fmt.Sprintf("%s release changed", ticket.Key),
				Payload: map[string]any{
					"ticket_id":     ticket.ID,
					"ticket_key":    ticket.Key,
					"version_id":    versionID,
					"actor_user_id": event.ActorID,
				},
			})
		}
		return plans, nil
	default:
		return nil, nil
	}
}

func (s *Service) matchingPolicies(ctx context.Context, eventType string, projectID string) ([]Policy, error) {
	var matched []Policy
	globalPolicies, err := s.ListPolicies(ctx, ListPoliciesInput{ScopeType: PolicyScopeGlobal})
	if err != nil {
		return nil, err
	}
	matched = appendMatchingPolicies(matched, globalPolicies, eventType)
	if projectID != "" {
		projectPolicies, err := s.ListPolicies(ctx, ListPoliciesInput{ScopeType: PolicyScopeProject, ProjectID: projectID})
		if err != nil {
			return nil, err
		}
		matched = appendMatchingPolicies(matched, projectPolicies, eventType)
	}
	return matched, nil
}

func appendMatchingPolicies(result []Policy, policies []Policy, eventType string) []Policy {
	for _, policy := range policies {
		if policy.Enabled && slices.Contains(policy.EventTypes, eventType) {
			result = append(result, policy)
		}
	}
	return result
}

func (s *Service) handleCommentCreated(ctx context.Context, event events.Event) error {
	ticketID, _ := event.Data["ticket_id"].(string)
	if ticketID == "" {
		return nil
	}
	ticket, err := s.ticket(ctx, ticketID)
	if err != nil {
		return err
	}
	watchers, err := s.ticketWatcherUserIDs(ctx, ticket.ID)
	if err != nil {
		return err
	}
	recipients := recipientSet(event.ActorID, append([]string{ticket.ReporterID, ticket.AssigneeID}, watchers...)...)
	for userID := range recipients {
		allowed, err := s.inAppNotificationAllowed(ctx, userID, ticket.ProjectID, "comment_added")
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}
		if _, err := s.Create(ctx, CreateInput{
			UserID:      userID,
			Type:        "comment_added",
			SubjectType: "ticket",
			SubjectID:   ticket.ID,
			Body:        fmt.Sprintf("New comment on %s", ticket.Key),
			Data: map[string]any{
				"ticket_id":  ticket.ID,
				"ticket_key": ticket.Key,
				"comment_id": event.ObjectID,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) handleCommentMentioned(ctx context.Context, event events.Event) error {
	ticketID, _ := event.Data["ticket_id"].(string)
	mentionedUserID, _ := event.Data["mentioned_user_id"].(string)
	if ticketID == "" || mentionedUserID == "" || mentionedUserID == event.ActorID {
		return nil
	}
	ticket, err := s.ticket(ctx, ticketID)
	if err != nil {
		return err
	}
	if ok, err := s.activeUser(ctx, mentionedUserID); err != nil {
		return err
	} else if !ok {
		return nil
	}
	allowed, err := s.inAppNotificationAllowed(ctx, mentionedUserID, ticket.ProjectID, "comment_mentioned")
	if err != nil {
		return err
	}
	if !allowed {
		return nil
	}
	if _, err := s.Create(ctx, CreateInput{
		UserID:      mentionedUserID,
		Type:        "comment_mentioned",
		SubjectType: "ticket",
		SubjectID:   ticket.ID,
		Body:        fmt.Sprintf("You were mentioned on %s", ticket.Key),
		Data: map[string]any{
			"ticket_id":           ticket.ID,
			"ticket_key":          ticket.Key,
			"comment_id":          eventCommentID(event),
			"mentioned_username":  event.Data["mentioned_username"],
			"mentioned_user_name": event.Data["mentioned_user_name"],
		},
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) handleTicketUpdated(ctx context.Context, event events.Event) error {
	ticket, err := s.ticket(ctx, event.ObjectID)
	if err != nil {
		return err
	}
	if newAssignee := changeNew(event.Data, "assignee_id"); newAssignee != "" && newAssignee != event.ActorID {
		allowed, err := s.inAppNotificationAllowed(ctx, newAssignee, ticket.ProjectID, "ticket_assigned")
		if err != nil {
			return err
		}
		if allowed {
			if _, err := s.Create(ctx, CreateInput{
				UserID:      newAssignee,
				Type:        "ticket_assigned",
				SubjectType: "ticket",
				SubjectID:   ticket.ID,
				Body:        fmt.Sprintf("Assigned to %s", ticket.Key),
				Data: map[string]any{
					"ticket_id":  ticket.ID,
					"ticket_key": ticket.Key,
				},
			}); err != nil {
				return err
			}
		}
	}
	if status := changeNew(event.Data, "status"); status != "" {
		if err := s.createTicketChangeNotifications(ctx, event.ActorID, ticket, "ticket_status_changed", fmt.Sprintf("%s moved to %s", ticket.Key, status), map[string]any{"status": status}); err != nil {
			return err
		}
	}
	if sprintID := changeNew(event.Data, "sprint_id"); sprintID != "" {
		if err := s.createTicketChangeNotifications(ctx, event.ActorID, ticket, "sprint_changed", fmt.Sprintf("%s sprint changed", ticket.Key), map[string]any{"sprint_id": sprintID}); err != nil {
			return err
		}
	}
	if versionID := changeNew(event.Data, "version_id"); versionID != "" {
		if err := s.createTicketChangeNotifications(ctx, event.ActorID, ticket, "release_changed", fmt.Sprintf("%s release changed", ticket.Key), map[string]any{"version_id": versionID}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) createTicketChangeNotifications(ctx context.Context, actorID string, ticket eventTicket, notificationType string, body string, data map[string]any) error {
	watchers, err := s.ticketWatcherUserIDs(ctx, ticket.ID)
	if err != nil {
		return err
	}
	recipients := recipientSet(actorID, append([]string{ticket.ReporterID, ticket.AssigneeID}, watchers...)...)
	for userID := range recipients {
		allowed, err := s.inAppNotificationAllowed(ctx, userID, ticket.ProjectID, notificationType)
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}
		payload := map[string]any{
			"ticket_id":  ticket.ID,
			"ticket_key": ticket.Key,
		}
		for key, value := range data {
			payload[key] = value
		}
		if _, err := s.Create(ctx, CreateInput{
			UserID:      userID,
			Type:        notificationType,
			SubjectType: "ticket",
			SubjectID:   ticket.ID,
			Body:        body,
			Data:        payload,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) inAppNotificationAllowed(ctx context.Context, userID string, projectID string, notificationType string) (bool, error) {
	userPreferences, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		return false, err
	}
	if !preferencesAllowInApp(userPreferences, notificationType) {
		return false, nil
	}
	projectPreferences, err := s.GetProjectPreferences(ctx, projectID)
	if err != nil {
		return false, err
	}
	return preferencesAllowInApp(projectPreferences, notificationType), nil
}

func preferencesAllowInApp(preferences Preferences, notificationType string) bool {
	if !preferences.InAppEnabled {
		return false
	}
	switch notificationType {
	case "ticket_assigned":
		return preferences.AssignmentEnabled
	case "comment_added", "comment_mentioned":
		return preferences.CommentEnabled
	case "ticket_status_changed":
		return preferences.StatusChangeEnabled
	case "sprint_changed":
		return preferences.SprintChangeEnabled
	case "release_changed":
		return preferences.ReleaseChangeEnabled
	case "automation_failed":
		return preferences.AutomationFailureEnabled
	default:
		return true
	}
}

type eventTicket struct {
	ID         string
	ProjectID  string
	Key        string
	ReporterID string
	AssigneeID string
}

func (s *Service) ticket(ctx context.Context, ticketID string) (eventTicket, error) {
	var ticket eventTicket
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, key, COALESCE(reporter_id, ''), COALESCE(assignee_id, '')
		FROM tickets
		WHERE id = ? AND deleted_at IS NULL
	`, ticketID).Scan(&ticket.ID, &ticket.ProjectID, &ticket.Key, &ticket.ReporterID, &ticket.AssigneeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return eventTicket{}, ErrNotFound
		}
		return eventTicket{}, fmt.Errorf("get notification ticket: %w", err)
	}
	return ticket, nil
}

func (s *Service) ticketWatcherUserIDs(ctx context.Context, ticketID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tw.user_id
		FROM ticket_watchers tw
		JOIN users u ON u.id = tw.user_id
		WHERE tw.ticket_id = ? AND u.deleted_at IS NULL AND u.is_disabled = 0
		ORDER BY tw.user_id ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list notification ticket watchers: %w", err)
	}
	defer rows.Close()
	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan notification ticket watcher: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification ticket watchers: %w", err)
	}
	return userIDs, nil
}

func (s *Service) activeUser(ctx context.Context, userID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return false, nil
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE id = ? AND deleted_at IS NULL AND is_disabled = 0
	`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check notification user: %w", err)
	}
	return exists == 1, nil
}

func displayActor(actorID string) string {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return "Someone"
	}
	return actorID
}

func displayMention(username string, userID string) string {
	username = strings.TrimSpace(username)
	if username != "" {
		return "@" + username
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "someone"
	}
	return userID
}

func eventCommentID(event events.Event) string {
	if commentID, _ := event.Data["comment_id"].(string); strings.TrimSpace(commentID) != "" {
		return strings.TrimSpace(commentID)
	}
	return strings.TrimSpace(event.ObjectID)
}

func eventProjectID(event events.Event, fallback string) string {
	if strings.TrimSpace(event.ProjectID) != "" {
		return strings.TrimSpace(event.ProjectID)
	}
	return strings.TrimSpace(fallback)
}

func deliveryIdempotencyKey(domainEventID string, policyID string, destinationID string, eventType string) string {
	return strings.Join([]string{domainEventID, policyID, destinationID, eventType}, ":")
}

func (s *Service) getForUser(ctx context.Context, userID string, notificationID string) (Notification, error) {
	notification, err := scanNotification(s.db.QueryRowContext(ctx, `
		SELECT id, user_id, type, subject_type, subject_id, body, data_json, read_at, created_at
		FROM notifications
		WHERE id = ? AND user_id = ?
	`, notificationID, userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Notification{}, ErrNotFound
		}
		return Notification{}, fmt.Errorf("get notification: %w", err)
	}
	return notification, nil
}

func scanNotification(scanner interface{ Scan(...any) error }) (Notification, error) {
	var notification Notification
	var subjectType sql.NullString
	var subjectID sql.NullString
	var dataJSON string
	var readAt sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&subjectType,
		&subjectID,
		&notification.Body,
		&dataJSON,
		&readAt,
		&createdAt,
	); err != nil {
		return Notification{}, err
	}
	notification.SubjectType = nullString(subjectType)
	notification.SubjectID = nullString(subjectID)
	notification.Data = map[string]any{}
	if dataJSON != "" {
		if err := json.Unmarshal([]byte(dataJSON), &notification.Data); err != nil {
			return Notification{}, fmt.Errorf("decode notification data: %w", err)
		}
	}
	notification.ReadAt = parseNullableTime(readAt)
	created, err := parseTime(createdAt)
	if err != nil {
		return Notification{}, err
	}
	notification.CreatedAt = created
	return notification, nil
}

func recipientSet(actorID string, userIDs ...string) map[string]struct{} {
	recipients := map[string]struct{}{}
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" || userID == actorID {
			continue
		}
		recipients[userID] = struct{}{}
	}
	return recipients
}

func changeNew(data map[string]any, field string) string {
	changes, ok := data["changes"]
	if !ok {
		return ""
	}
	encoded, err := json.Marshal(changes)
	if err != nil {
		return ""
	}
	decoded := map[string]map[string]string{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return ""
	}
	return decoded[field]["new"]
}

func marshalData(data map[string]any) (string, error) {
	encoded, err := json.Marshal(nonNilMap(data))
	if err != nil {
		return "", fmt.Errorf("encode notification data: %w", err)
	}
	return string(encoded), nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func normalizeList(limit int, offset int) (int, int, error) {
	if limit < 0 || offset < 0 {
		return 0, 0, fmt.Errorf("%w: limit and offset must be non-negative", ErrValidation)
	}
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		return 0, 0, fmt.Errorf("%w: limit must be 200 or fewer", ErrValidation)
	}
	return limit, offset, nil
}

func joinAnd(parts []string) string {
	result := ""
	for index, part := range parts {
		if index > 0 {
			result += " AND "
		}
		result += part
	}
	return result
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate notification id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse notification time: %w", err)
	}
	return parsed, nil
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	parsed, err := parseTime(value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
