package projects

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects", "Projects", "List projects"), provider.listProjects)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects", "Projects", "Create project", http.StatusCreated), provider.createProject)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}", "Projects", "Get project"), provider.getProject)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/backlog", "Backlog", "List project backlog"), provider.listBacklog)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/projects/{project_id}/backlog", "Backlog", "Reorder project backlog"), provider.reorderBacklog)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/statuses", "Workflows", "List project workflow statuses"), provider.listStatuses)
	huma.Register(api, shared.Operation(http.MethodPut, "/api/projects/{project_id}/statuses", "Workflows", "Replace project workflow statuses"), provider.replaceStatuses)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/boards", "Boards", "List project boards"), provider.listBoards)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/boards", "Boards", "Create project board", http.StatusCreated), provider.createBoard)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/components", "Components", "List project components"), provider.listComponents)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/components", "Components", "Create project component", http.StatusCreated), provider.createComponent)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/versions", "Versions", "List project versions"), provider.listVersions)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/versions", "Versions", "Create project version", http.StatusCreated), provider.createVersion)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/custom-fields", "Custom Fields", "List project custom fields"), provider.listCustomFields)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/custom-fields", "Custom Fields", "Create project custom field", http.StatusCreated), provider.createCustomField)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/tickets", "Tickets", "List project tickets"), provider.listTickets)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/tickets", "Tickets", "Create ticket", http.StatusCreated), provider.createTicket)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/roadmap", "Roadmap", "List project roadmap"), provider.listRoadmap)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/sprints", "Sprints", "List project sprints"), provider.listSprints)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/sprints", "Sprints", "Create project sprint", http.StatusCreated), provider.createSprint)
}

func (provider Provider) listProjects(ctx context.Context, input *ListProjectsInput) (*ListProjectsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListProjects(ctx, principal, tracker.ListProjectsInput{
		IncludeArchived: input.IncludeArchived,
		Limit:           input.Limit,
		Offset:          input.Offset,
	})
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListProjectsOutput{Body: shared.ItemList[tracker.Project]{Items: items}}, nil
}

func (provider Provider) createProject(ctx context.Context, input *CreateProjectInput) (*CreateProjectOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	project, err := provider.Tracker.CreateProject(ctx, principal, input.Body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateProjectOutput{Body: project}, nil
}

func (provider Provider) getProject(ctx context.Context, input *ProjectIDInput) (*ProjectOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	project, err := provider.Tracker.GetProject(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ProjectOutput{Body: project}, nil
}

func (provider Provider) listBacklog(ctx context.Context, input *ProjectIDInput) (*ListTicketsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListBacklog(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListTicketsOutput{Body: shared.ItemList[tracker.Ticket]{Items: items}}, nil
}

func (provider Provider) reorderBacklog(ctx context.Context, input *ReorderBacklogInput) (*ListTicketsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ReorderBacklog(ctx, principal, input.ProjectID, input.Body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListTicketsOutput{Body: shared.ItemList[tracker.Ticket]{Items: items}}, nil
}

func (provider Provider) listStatuses(ctx context.Context, input *ProjectIDInput) (*ListStatusesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListProjectStatuses(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListStatusesOutput{Body: shared.ItemList[tracker.ProjectStatus]{Items: items}}, nil
}

func (provider Provider) replaceStatuses(ctx context.Context, input *ReplaceStatusesInput) (*ListStatusesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ReplaceProjectStatuses(ctx, principal, input.ProjectID, input.Body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListStatusesOutput{Body: shared.ItemList[tracker.ProjectStatus]{Items: items}}, nil
}

func (provider Provider) listBoards(ctx context.Context, input *ProjectIDInput) (*ListBoardsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListBoards(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListBoardsOutput{Body: shared.ItemList[tracker.Board]{Items: items}}, nil
}

func (provider Provider) createBoard(ctx context.Context, input *CreateBoardInput) (*CreateBoardOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	body := input.Body
	body.ProjectID = input.ProjectID
	board, err := provider.Tracker.CreateBoard(ctx, principal, body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateBoardOutput{Body: board}, nil
}

func (provider Provider) listComponents(ctx context.Context, input *ProjectIDInput) (*ListComponentsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListComponents(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListComponentsOutput{Body: shared.ItemList[tracker.Component]{Items: items}}, nil
}

func (provider Provider) createComponent(ctx context.Context, input *CreateComponentInput) (*CreateComponentOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	body := input.Body
	body.ProjectID = input.ProjectID
	component, err := provider.Tracker.CreateComponent(ctx, principal, body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateComponentOutput{Body: component}, nil
}

func (provider Provider) listVersions(ctx context.Context, input *ListVersionsInput) (*ListVersionsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListVersions(ctx, principal, input.ProjectID, input.Status)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListVersionsOutput{Body: shared.ItemList[tracker.Version]{Items: items}}, nil
}

func (provider Provider) createVersion(ctx context.Context, input *CreateVersionInput) (*CreateVersionOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	body := input.Body
	body.ProjectID = input.ProjectID
	version, err := provider.Tracker.CreateVersion(ctx, principal, body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateVersionOutput{Body: version}, nil
}

func (provider Provider) listCustomFields(ctx context.Context, input *ProjectIDInput) (*ListCustomFieldsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListCustomFields(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListCustomFieldsOutput{Body: shared.ItemList[tracker.CustomFieldDefinition]{Items: items}}, nil
}

func (provider Provider) createCustomField(ctx context.Context, input *CreateCustomFieldInput) (*CreateCustomFieldOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	body := input.Body
	body.ProjectID = input.ProjectID
	field, err := provider.Tracker.CreateCustomField(ctx, principal, body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateCustomFieldOutput{Body: field}, nil
}

func (provider Provider) listTickets(ctx context.Context, input *ListTicketsInput) (*ListTicketsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListTickets(ctx, principal, tracker.ListTicketsInput{
		ProjectID:   input.ProjectID,
		Status:      input.Status,
		AssigneeID:  input.AssigneeID,
		SprintID:    input.SprintID,
		ComponentID: input.ComponentID,
		VersionID:   input.VersionID,
		Label:       input.Label,
		Limit:       input.Limit,
		Offset:      input.Offset,
	})
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListTicketsOutput{Body: shared.ItemList[tracker.Ticket]{Items: items}}, nil
}

func (provider Provider) createTicket(ctx context.Context, input *CreateTicketInput) (*CreateTicketOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	body := input.Body
	body.ProjectID = input.ProjectID
	ticket, err := provider.Tracker.CreateTicket(ctx, principal, body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateTicketOutput{Body: ticket}, nil
}

func (provider Provider) listRoadmap(ctx context.Context, input *ProjectIDInput) (*ListRoadmapOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListRoadmap(ctx, principal, input.ProjectID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListRoadmapOutput{Body: shared.ItemList[tracker.RoadmapItem]{Items: items}}, nil
}

func (provider Provider) listSprints(ctx context.Context, input *ListSprintsInput) (*ListSprintsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListSprints(ctx, principal, input.ProjectID, input.State)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListSprintsOutput{Body: shared.ItemList[tracker.Sprint]{Items: items}}, nil
}

func (provider Provider) createSprint(ctx context.Context, input *CreateSprintInput) (*CreateSprintOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	body := input.Body
	body.ProjectID = input.ProjectID
	sprint, err := provider.Tracker.CreateSprint(ctx, principal, body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateSprintOutput{Body: sprint}, nil
}
