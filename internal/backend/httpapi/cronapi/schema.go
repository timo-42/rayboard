package cronapi

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type ListJobsInput struct {
	shared.AuthInput
	ProjectID string `query:"project_id" doc:"Filter cron jobs by project ID."`
	Limit     int    `query:"limit" doc:"Maximum number of cron jobs to return."`
	Offset    int    `query:"offset" doc:"Number of cron jobs to skip."`
}

type CreateJobInput struct {
	shared.AuthInput
	Body shared.ResourceInput[JobSpec]
}

type JobIDInput struct {
	shared.AuthInput
	JobID string `path:"job_id" doc:"Cron job ID."`
}

type UpdateJobInput struct {
	shared.AuthInput
	JobID string `path:"job_id" doc:"Cron job ID."`
	Body  shared.ResourceInput[UpdateJobSpec]
}

type ListRunsInput struct {
	shared.AuthInput
	JobID  string `path:"job_id" doc:"Cron job ID."`
	Limit  int    `query:"limit" doc:"Maximum number of runs to return."`
	Offset int    `query:"offset" doc:"Number of runs to skip."`
}

type JobOutput struct {
	Body JobResource
}

type JobMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobSpec struct {
	OwnerUserID string              `json:"owner_user_id,omitempty"`
	ProjectID   string              `json:"project_id,omitempty"`
	Name        string              `json:"name,omitempty"`
	Schedule    string              `json:"schedule,omitempty"`
	Timezone    string              `json:"timezone,omitempty"`
	Enabled     bool                `json:"enabled,omitempty"`
	Engine      cronjobs.EngineSpec `json:"engine"`
}

type UpdateJobSpec struct {
	OwnerUserID *string              `json:"owner_user_id,omitempty"`
	ProjectID   *string              `json:"project_id,omitempty"`
	Name        *string              `json:"name,omitempty"`
	Schedule    *string              `json:"schedule,omitempty"`
	Timezone    *string              `json:"timezone,omitempty"`
	Enabled     *bool                `json:"enabled,omitempty"`
	Engine      *cronjobs.EngineSpec `json:"engine,omitempty"`
}

type JobStatus struct {
	LastRunStatus string     `json:"last_run_status,omitempty"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time `json:"next_run_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
}

type JobResource = shared.Resource[JobMetadata, JobSpec, JobStatus]

type ListJobsOutput = shared.ListOutput[JobResource]
type CreateJobOutput = shared.CreatedOutput[JobResource]
type RunJobOutput = shared.AcceptedOutput[automation.Run]
type ListRunsOutput = shared.ListOutput[automation.Run]

func (spec JobSpec) createInput() cronjobs.CreateInput {
	return cronjobs.CreateInput{
		OwnerUserID: spec.OwnerUserID,
		ProjectID:   spec.ProjectID,
		Name:        spec.Name,
		Schedule:    spec.Schedule,
		Timezone:    spec.Timezone,
		Enabled:     spec.Enabled,
		Engine:      spec.Engine,
	}
}

func (spec UpdateJobSpec) updateInput() cronjobs.UpdateInput {
	return cronjobs.UpdateInput{
		OwnerUserID: spec.OwnerUserID,
		ProjectID:   spec.ProjectID,
		Name:        spec.Name,
		Schedule:    spec.Schedule,
		Timezone:    spec.Timezone,
		Enabled:     spec.Enabled,
		Engine:      spec.Engine,
	}
}

func jobResource(job cronjobs.Job) JobResource {
	return JobResource{
		Metadata: JobMetadata{
			ID:        job.ID,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
		},
		Spec: JobSpec{
			OwnerUserID: job.OwnerUserID,
			ProjectID:   job.ProjectID,
			Name:        job.Name,
			Schedule:    job.Schedule,
			Timezone:    job.Timezone,
			Enabled:     job.Enabled,
			Engine:      job.Engine,
		},
		Status: JobStatus{
			LastRunStatus: job.LastRunStatus,
			LastRunAt:     job.LastRunAt,
			NextRunAt:     job.NextRunAt,
			LastError:     job.LastError,
		},
	}
}

func jobResources(jobs []cronjobs.Job) []JobResource {
	resources := make([]JobResource, 0, len(jobs))
	for _, job := range jobs {
		resources = append(resources, jobResource(job))
	}
	return resources
}
