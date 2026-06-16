package cronapi

import (
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
	Body cronjobs.CreateInput
}

type JobIDInput struct {
	shared.AuthInput
	JobID string `path:"job_id" doc:"Cron job ID."`
}

type UpdateJobInput struct {
	shared.AuthInput
	JobID string `path:"job_id" doc:"Cron job ID."`
	Body  cronjobs.UpdateInput
}

type ListRunsInput struct {
	shared.AuthInput
	JobID  string `path:"job_id" doc:"Cron job ID."`
	Limit  int    `query:"limit" doc:"Maximum number of runs to return."`
	Offset int    `query:"offset" doc:"Number of runs to skip."`
}

type JobOutput struct {
	Body cronjobs.Job
}

type ListJobsOutput = shared.ListOutput[cronjobs.Job]
type CreateJobOutput = shared.CreatedOutput[cronjobs.Job]
type RunJobOutput = shared.AcceptedOutput[automation.Run]
type ListRunsOutput = shared.ListOutput[automation.Run]
