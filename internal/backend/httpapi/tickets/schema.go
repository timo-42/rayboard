package tickets

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type TicketIDInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
}

type UpdateTicketInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
	Body     shared.ResourceInput[TicketUpdateSpec]
}

type AssignSprintInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
	Body     shared.ResourceInput[AssignSprintSpec]
}

type CreateTicketLinkInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
	Body     shared.ResourceInput[TicketLinkCreateSpec]
}

type DeleteTicketLinkInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
	LinkID   string `path:"link_id"`
}

type AssignSprintSpec struct {
	SprintID string `json:"sprint_id,omitempty"`
}

type TicketLinkCreateSpec struct {
	TargetTicketID string `json:"target_ticket_id,omitempty"`
	LinkType       string `json:"link_type,omitempty"`
}

type TicketOutput struct {
	Body TicketResource
}

type ActivityOutput = shared.ListOutput[ActivityResource]
type TicketWatchersOutput = shared.ListOutput[TicketWatcherResource]
type TicketLinksOutput = shared.ListOutput[TicketLinkResource]
type CreateTicketLinkOutput = shared.CreatedOutput[TicketLinkResource]

type TicketMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TicketSpec struct {
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Status         string         `json:"status,omitempty"`
	Priority       string         `json:"priority,omitempty"`
	Type           string         `json:"type,omitempty"`
	AssigneeID     string         `json:"assignee_id,omitempty"`
	ParentTicketID string         `json:"parent_ticket_id,omitempty"`
	SprintID       string         `json:"sprint_id,omitempty"`
	ComponentID    string         `json:"component_id,omitempty"`
	VersionID      string         `json:"version_id,omitempty"`
	Rank           string         `json:"rank,omitempty"`
	StartDate      string         `json:"start_date,omitempty"`
	DueDate        string         `json:"due_date,omitempty"`
	StoryPoints    *float64       `json:"story_points,omitempty"`
	Labels         []string       `json:"labels,omitempty"`
	CustomFields   map[string]any `json:"custom_fields,omitempty"`
}

type TicketUpdateSpec struct {
	Title          *string         `json:"title,omitempty"`
	Description    *string         `json:"description,omitempty"`
	Status         *string         `json:"status,omitempty"`
	Priority       *string         `json:"priority,omitempty"`
	Type           *string         `json:"type,omitempty"`
	AssigneeID     *string         `json:"assignee_id,omitempty"`
	ParentTicketID *string         `json:"parent_ticket_id,omitempty"`
	SprintID       *string         `json:"sprint_id,omitempty"`
	ComponentID    *string         `json:"component_id,omitempty"`
	VersionID      *string         `json:"version_id,omitempty"`
	Rank           *string         `json:"rank,omitempty"`
	StartDate      *string         `json:"start_date,omitempty"`
	DueDate        *string         `json:"due_date,omitempty"`
	StoryPoints    OptionalFloat   `json:"story_points,omitempty"`
	Labels         *[]string       `json:"labels,omitempty"`
	CustomFields   *map[string]any `json:"custom_fields,omitempty"`
}

type OptionalFloat struct {
	Set   bool
	Value *float64
}

func (value *OptionalFloat) UnmarshalJSON(data []byte) error {
	value.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		value.Value = nil
		return nil
	}
	var decoded float64
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

type TicketStatus struct {
	Key          string     `json:"key"`
	ReporterID   string     `json:"reporter_id,omitempty"`
	WatcherCount int        `json:"watcher_count"`
	Watching     bool       `json:"watching"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type TicketResource = shared.Resource[TicketMetadata, TicketSpec, TicketStatus]

type ActivityMetadata struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ActivitySpec struct {
	ActivityType string         `json:"activity_type"`
	Data         map[string]any `json:"data,omitempty"`
}

type ActivityStatus struct {
	ActorID string `json:"actor_id,omitempty"`
}

type ActivityResource = shared.Resource[ActivityMetadata, ActivitySpec, ActivityStatus]

type TicketWatcherMetadata struct {
	TicketID  string    `json:"ticket_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketWatcherSpec struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type TicketWatcherStatus struct{}

type TicketWatcherResource = shared.Resource[TicketWatcherMetadata, TicketWatcherSpec, TicketWatcherStatus]

type TicketLinkMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketLinkSpec struct {
	LinkType string         `json:"link_type"`
	Source   TicketResource `json:"source"`
	Target   TicketResource `json:"target"`
}

type TicketLinkStatus struct {
	CreatedBy string `json:"created_by,omitempty"`
}

type TicketLinkResource = shared.Resource[TicketLinkMetadata, TicketLinkSpec, TicketLinkStatus]

func (spec TicketSpec) ToCreateInput(projectID string) tracker.CreateTicketInput {
	return tracker.CreateTicketInput{
		ProjectID:      projectID,
		Title:          spec.Title,
		Description:    spec.Description,
		Status:         spec.Status,
		Priority:       spec.Priority,
		Type:           spec.Type,
		AssigneeID:     spec.AssigneeID,
		ParentTicketID: spec.ParentTicketID,
		SprintID:       spec.SprintID,
		ComponentID:    spec.ComponentID,
		VersionID:      spec.VersionID,
		Rank:           spec.Rank,
		StartDate:      spec.StartDate,
		DueDate:        spec.DueDate,
		StoryPoints:    spec.StoryPoints,
		Labels:         spec.Labels,
		CustomFields:   spec.CustomFields,
	}
}

func (spec TicketUpdateSpec) ToUpdateInput() tracker.UpdateTicketInput {
	return tracker.UpdateTicketInput{
		Title:          spec.Title,
		Description:    spec.Description,
		Status:         spec.Status,
		Priority:       spec.Priority,
		Type:           spec.Type,
		AssigneeID:     spec.AssigneeID,
		ParentTicketID: spec.ParentTicketID,
		SprintID:       spec.SprintID,
		ComponentID:    spec.ComponentID,
		VersionID:      spec.VersionID,
		Rank:           spec.Rank,
		StartDate:      spec.StartDate,
		DueDate:        spec.DueDate,
		StoryPoints:    spec.StoryPoints.Value,
		StoryPointsSet: spec.StoryPoints.Set,
		Labels:         spec.Labels,
		CustomFields:   spec.CustomFields,
	}
}

func (spec TicketLinkCreateSpec) ToCreateInput() tracker.CreateTicketLinkInput {
	return tracker.CreateTicketLinkInput{
		TargetTicketID: spec.TargetTicketID,
		LinkType:       spec.LinkType,
	}
}

func ResourceFromTracker(ticket tracker.Ticket) TicketResource {
	return TicketResource{
		Metadata: TicketMetadata{
			ID:        ticket.ID,
			ProjectID: ticket.ProjectID,
			CreatedAt: ticket.CreatedAt,
			UpdatedAt: ticket.UpdatedAt,
		},
		Spec: TicketSpec{
			Title:          ticket.Title,
			Description:    ticket.Description,
			Status:         ticket.Status,
			Priority:       ticket.Priority,
			Type:           ticket.Type,
			AssigneeID:     ticket.AssigneeID,
			ParentTicketID: ticket.ParentTicketID,
			SprintID:       ticket.SprintID,
			ComponentID:    ticket.ComponentID,
			VersionID:      ticket.VersionID,
			Rank:           ticket.Rank,
			StartDate:      ticket.StartDate,
			DueDate:        ticket.DueDate,
			StoryPoints:    ticket.StoryPoints,
			Labels:         ticket.Labels,
			CustomFields:   ticket.CustomFields,
		},
		Status: TicketStatus{
			Key:          ticket.Key,
			ReporterID:   ticket.ReporterID,
			WatcherCount: ticket.WatcherCount,
			Watching:     ticket.Watching,
			DeletedAt:    ticket.DeletedAt,
		},
	}
}

func ResourcesFromTracker(tickets []tracker.Ticket) []TicketResource {
	resources := make([]TicketResource, 0, len(tickets))
	for _, ticket := range tickets {
		resources = append(resources, ResourceFromTracker(ticket))
	}
	return resources
}

func LinkResourceFromTracker(link tracker.TicketLink) TicketLinkResource {
	return TicketLinkResource{
		Metadata: TicketLinkMetadata{
			ID:        link.ID,
			ProjectID: link.ProjectID,
			CreatedAt: link.CreatedAt,
		},
		Spec: TicketLinkSpec{
			LinkType: link.LinkType,
			Source:   ResourceFromTracker(link.Source),
			Target:   ResourceFromTracker(link.Target),
		},
		Status: TicketLinkStatus{
			CreatedBy: link.CreatedBy,
		},
	}
}

func LinkResourcesFromTracker(links []tracker.TicketLink) []TicketLinkResource {
	resources := make([]TicketLinkResource, 0, len(links))
	for _, link := range links {
		resources = append(resources, LinkResourceFromTracker(link))
	}
	return resources
}

func WatcherResourceFromTracker(watcher tracker.TicketWatcher) TicketWatcherResource {
	return TicketWatcherResource{
		Metadata: TicketWatcherMetadata{
			TicketID:  watcher.TicketID,
			UserID:    watcher.UserID,
			CreatedAt: watcher.CreatedAt,
		},
		Spec: TicketWatcherSpec{
			Username:    watcher.Username,
			DisplayName: watcher.DisplayName,
		},
		Status: TicketWatcherStatus{},
	}
}

func WatcherResourcesFromTracker(watchers []tracker.TicketWatcher) []TicketWatcherResource {
	resources := make([]TicketWatcherResource, 0, len(watchers))
	for _, watcher := range watchers {
		resources = append(resources, WatcherResourceFromTracker(watcher))
	}
	return resources
}

func ActivityResourceFromTracker(activity tracker.TicketActivity) ActivityResource {
	return ActivityResource{
		Metadata: ActivityMetadata{
			ID:        activity.ID,
			TicketID:  activity.TicketID,
			CreatedAt: activity.CreatedAt,
		},
		Spec: ActivitySpec{
			ActivityType: activity.ActivityType,
			Data:         activity.Data,
		},
		Status: ActivityStatus{
			ActorID: activity.ActorID,
		},
	}
}

func ActivityResourcesFromTracker(activities []tracker.TicketActivity) []ActivityResource {
	resources := make([]ActivityResource, 0, len(activities))
	for _, activity := range activities {
		resources = append(resources, ActivityResourceFromTracker(activity))
	}
	return resources
}
