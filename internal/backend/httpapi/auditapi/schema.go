package auditapi

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type ListAuditLogInput struct {
	shared.AuthInput
	EventType   string `query:"event_type" doc:"Filter by audit event type."`
	ActorID     string `query:"actor_user_id" doc:"Filter by actor user ID."`
	SubjectType string `query:"subject_type" doc:"Filter by subject type."`
	SubjectID   string `query:"subject_id" doc:"Filter by subject ID."`
	Outcome     string `query:"outcome" enum:"success,failure" doc:"Filter by audit outcome."`
	Limit       int    `query:"limit" minimum:"1" maximum:"500" default:"100" doc:"Maximum number of entries to return."`
}

type AuditEntryMetadata struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
}

type AuditEntrySpec struct {
	EventType   string         `json:"event_type"`
	ActorID     string         `json:"actor_user_id,omitempty"`
	AuthKind    authz.AuthKind `json:"auth_kind,omitempty"`
	SubjectType string         `json:"subject_type"`
	SubjectID   string         `json:"subject_id,omitempty"`
	Outcome     string         `json:"outcome"`
	Payload     map[string]any `json:"payload"`
}

type AuditEntryStatus struct {
	SecurityEvent bool `json:"security_event"`
}

type AuditEntryResource = shared.Resource[AuditEntryMetadata, AuditEntrySpec, AuditEntryStatus]
type ListAuditLogOutput = shared.ListOutput[AuditEntryResource]

func auditEntryResource(entry audit.Entry) AuditEntryResource {
	return AuditEntryResource{
		Metadata: AuditEntryMetadata{
			ID:         entry.ID,
			OccurredAt: entry.OccurredAt,
		},
		Spec: AuditEntrySpec{
			EventType:   entry.EventType,
			ActorID:     entry.ActorID,
			AuthKind:    entry.AuthKind,
			SubjectType: entry.SubjectType,
			SubjectID:   entry.SubjectID,
			Outcome:     entry.Outcome,
			Payload:     entry.Payload,
		},
		Status: AuditEntryStatus{
			SecurityEvent: true,
		},
	}
}

func auditEntryResources(entries []audit.Entry) []AuditEntryResource {
	resources := make([]AuditEntryResource, 0, len(entries))
	for _, entry := range entries {
		resources = append(resources, auditEntryResource(entry))
	}
	return resources
}
