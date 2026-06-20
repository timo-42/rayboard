package versions

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type VersionIDInput struct {
	shared.AuthInput
	VersionID string `path:"version_id"`
}

type UpdateVersionInput struct {
	shared.AuthInput
	VersionID string `path:"version_id"`
	Body      shared.ResourceInput[UpdateVersionSpec]
}

type VersionOutput struct {
	Body VersionResource
}

type VersionReportOutput struct {
	Body VersionReportResource
}

type VersionMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VersionSpec struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TargetDate  string `json:"target_date,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
}

type UpdateVersionSpec struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
}

type VersionStatus struct {
	State string `json:"state,omitempty"`
}

type VersionResource = shared.Resource[VersionMetadata, VersionSpec, VersionStatus]

type VersionReportMetadata struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
}

type VersionReportSpec struct {
	Version VersionResource `json:"version"`
}

type VersionReportStatus struct {
	Scope        string                            `json:"scope"`
	SnapshotAt   *time.Time                        `json:"snapshot_at,omitempty"`
	Progress     tracker.VersionReportProgress     `json:"progress"`
	Analytics    tracker.SprintAnalytics           `json:"analytics"`
	ScopeChanges tracker.VersionReportScopeChanges `json:"scope_changes"`
	Tickets      []ticketapi.TicketResource        `json:"tickets"`
}

type VersionReportResource = shared.Resource[VersionReportMetadata, VersionReportSpec, VersionReportStatus]

func (spec VersionSpec) ToCreateInput(projectID string) tracker.CreateVersionInput {
	return tracker.CreateVersionInput{
		ProjectID:   projectID,
		Name:        spec.Name,
		Description: spec.Description,
		TargetDate:  spec.TargetDate,
		ReleaseDate: spec.ReleaseDate,
	}
}

func (spec UpdateVersionSpec) ToUpdateInput() tracker.UpdateVersionInput {
	return tracker.UpdateVersionInput{
		Name:        spec.Name,
		Description: spec.Description,
		Status:      spec.Status,
		TargetDate:  spec.TargetDate,
		ReleaseDate: spec.ReleaseDate,
	}
}

func ResourceFromTracker(version tracker.Version) VersionResource {
	return VersionResource{
		Metadata: VersionMetadata{
			ID:        version.ID,
			ProjectID: version.ProjectID,
			CreatedAt: version.CreatedAt,
			UpdatedAt: version.UpdatedAt,
		},
		Spec: VersionSpec{
			Name:        version.Name,
			Description: version.Description,
			TargetDate:  version.TargetDate,
			ReleaseDate: version.ReleaseDate,
		},
		Status: VersionStatus{
			State: version.Status,
		},
	}
}

func ReportResourceFromTracker(report tracker.VersionReport) VersionReportResource {
	return VersionReportResource{
		Metadata: VersionReportMetadata{
			ID:        report.Version.ID,
			ProjectID: report.Version.ProjectID,
		},
		Spec: VersionReportSpec{
			Version: ResourceFromTracker(report.Version),
		},
		Status: VersionReportStatus{
			Scope:        report.Scope,
			SnapshotAt:   report.SnapshotAt,
			Progress:     report.Progress,
			Analytics:    report.Analytics,
			ScopeChanges: report.ScopeChanges,
			Tickets:      ticketapi.ResourcesFromTracker(report.Tickets),
		},
	}
}

func ResourcesFromTracker(versions []tracker.Version) []VersionResource {
	resources := make([]VersionResource, 0, len(versions))
	for _, version := range versions {
		resources = append(resources, ResourceFromTracker(version))
	}
	return resources
}
