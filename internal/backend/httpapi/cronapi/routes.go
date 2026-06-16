package cronapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/cron-jobs", "Cron Jobs", "List cron jobs"), provider.listJobs)
	huma.Register(api, operation(http.MethodPost, "/api/cron-jobs", "Cron Jobs", "Create cron job", http.StatusCreated), provider.createJob)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/cron-jobs/{job_id}", "Cron Jobs", "Get cron job"), provider.getJob)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/cron-jobs/{job_id}", "Cron Jobs", "Update cron job"), provider.updateJob)
	huma.Register(api, operation(http.MethodDelete, "/api/cron-jobs/{job_id}", "Cron Jobs", "Delete cron job", http.StatusNoContent), provider.deleteJob)
	huma.Register(api, operation(http.MethodPost, "/api/cron-jobs/{job_id}/run", "Cron Jobs", "Run cron job now", http.StatusAccepted), provider.runJob)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/cron-jobs/{job_id}/runs", "Cron Jobs", "List cron job runs"), provider.listRuns)
}

func (provider Provider) listJobs(ctx context.Context, input *ListJobsInput) (*ListJobsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	jobs, err := provider.Cron.List(ctx, principal, cronjobs.ListInput{
		ProjectID: input.ProjectID,
		Limit:     input.Limit,
		Offset:    input.Offset,
	})
	if err != nil {
		return nil, shared.CronError(err)
	}
	return &ListJobsOutput{Body: shared.ItemList[cronjobs.Job]{Items: jobs}}, nil
}

func (provider Provider) createJob(ctx context.Context, input *CreateJobInput) (*CreateJobOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	job, err := provider.Cron.Create(ctx, principal, input.Body)
	if err != nil {
		return nil, shared.CronError(err)
	}
	return &CreateJobOutput{Body: job}, nil
}

func (provider Provider) getJob(ctx context.Context, input *JobIDInput) (*JobOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	job, err := provider.Cron.Get(ctx, principal, input.JobID)
	if err != nil {
		return nil, shared.CronError(err)
	}
	return &JobOutput{Body: job}, nil
}

func (provider Provider) updateJob(ctx context.Context, input *UpdateJobInput) (*JobOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	job, err := provider.Cron.Update(ctx, principal, input.JobID, input.Body)
	if err != nil {
		return nil, shared.CronError(err)
	}
	return &JobOutput{Body: job}, nil
}

func (provider Provider) deleteJob(ctx context.Context, input *JobIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Cron.Delete(ctx, principal, input.JobID); err != nil {
		return nil, shared.CronError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) runJob(ctx context.Context, input *JobIDInput) (*RunJobOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	run, err := provider.Cron.RunNow(ctx, principal, input.JobID)
	if err != nil && run.ID == "" {
		return nil, shared.CronError(err)
	}
	return &RunJobOutput{Body: run}, nil
}

func (provider Provider) listRuns(ctx context.Context, input *ListRunsInput) (*ListRunsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	runs, err := provider.Cron.ListRuns(ctx, principal, input.JobID, input.Limit, input.Offset)
	if err != nil {
		return nil, shared.CronError(err)
	}
	return &ListRunsOutput{Body: shared.ItemList[automation.Run]{Items: runs}}, nil
}

func operation(method string, path string, tag string, summary string, status int) huma.Operation {
	op := shared.Operation(method, path, tag, summary)
	op.DefaultStatus = status
	return op
}
