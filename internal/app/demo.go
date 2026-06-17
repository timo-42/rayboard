package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/notifications"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
	"github.com/timo-42/rayboard/internal/config"
)

func runDemoSeed(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "seed" {
		fmt.Fprintln(stderr, "usage: rayboard demo seed --backend-url http://host:port --admin-user admin --admin-password <password> --fresh-reset")
		return 2
	}

	flags := flag.NewFlagSet("demo seed", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var backendURL string
	var adminUser string
	var adminPassword string
	var freshReset bool
	flags.StringVar(&backendURL, "backend-url", config.DefaultBackendURL, "backend API base URL")
	flags.StringVar(&adminUser, "admin-user", "admin", "admin username")
	flags.StringVar(&adminPassword, "admin-password", "", "admin password")
	flags.BoolVar(&freshReset, "fresh-reset", false, "confirm demo data should be seeded")

	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if backendURL == "" || adminUser == "" || adminPassword == "" || !freshReset {
		fmt.Fprintln(stderr, "demo seed requires --backend-url, --admin-user, --admin-password, and --fresh-reset")
		return 2
	}

	seeder, err := newDemoSeeder(backendURL, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "demo seed: %v\n", err)
		return 2
	}
	if err := seeder.seed(ctx, adminUser, adminPassword); err != nil {
		fmt.Fprintf(stderr, "demo seed: %v\n", err)
		return 1
	}
	return 0
}

type demoSeeder struct {
	baseURL *url.URL
	client  *http.Client
	stdout  io.Writer
	csrf    string
	suffix  string
}

type demoLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type demoGroup struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	} `json:"spec"`
}

func (group demoGroup) id() string {
	return group.Metadata.ID
}

type demoCreatedUser struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"spec"`
	Status struct {
		Password string `json:"password"`
	} `json:"status"`
}

func (user demoCreatedUser) id() string {
	return user.Metadata.ID
}

type demoRoleBinding struct {
	RoleName    authz.RoleName          `json:"role_name"`
	SubjectType authz.BindingTargetKind `json:"subject_type"`
	SubjectID   string                  `json:"subject_id"`
	Scope       authz.ScopeKind         `json:"scope"`
	ProjectID   string                  `json:"project_id,omitempty"`
}

type demoProjectResource struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		LeadUserID  string `json:"lead_user_id"`
	} `json:"spec"`
}

type demoTicketResource struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Title    string `json:"title"`
		Priority string `json:"priority"`
		Type     string `json:"type"`
		Status   string `json:"status"`
	} `json:"spec"`
	Status struct {
		Key string `json:"key"`
	} `json:"status"`
}

func (ticket demoTicketResource) id() string {
	return ticket.Metadata.ID
}

type demoIDResource struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
}

func (resource demoIDResource) id() string {
	return resource.Metadata.ID
}

type demoCreatedWebhookResource struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}

func (resource demoCreatedWebhookResource) id() string {
	return resource.Metadata.ID
}

func newDemoSeeder(rawBackendURL string, stdout io.Writer) (*demoSeeder, error) {
	parsed, err := url.Parse(strings.TrimRight(rawBackendURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse backend URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("backend URL must include scheme and host")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	suffix, err := randomDemoSuffix()
	if err != nil {
		return nil, err
	}
	return &demoSeeder{
		baseURL: parsed,
		client: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
		stdout: stdout,
		suffix: suffix,
	}, nil
}

func (s *demoSeeder) seed(ctx context.Context, adminUser string, adminPassword string) error {
	if err := s.login(ctx, adminUser, adminPassword); err != nil {
		return err
	}

	fmt.Fprintf(s.stdout, "demo seed suffix: %s\n", s.suffix)
	fmt.Fprintln(s.stdout, "demo admin login succeeded")

	users, err := s.createUsers(ctx)
	if err != nil {
		return err
	}
	groups, err := s.createGroups(ctx, users)
	if err != nil {
		return err
	}
	project, err := s.createProject(ctx, users["lead"].id())
	if err != nil {
		return err
	}
	if err := s.bindProjectAccess(ctx, groups["engineers"].id(), groups["stakeholders"].id(), users["lead"].id(), project.ID); err != nil {
		return err
	}
	assets, err := s.createProjectPlanning(ctx, project.ID, users)
	if err != nil {
		return err
	}
	if err := s.createTicketHook(ctx, project.ID); err != nil {
		return err
	}
	if err := s.createTicketCreatePage(ctx, project.ID, users, assets); err != nil {
		return err
	}
	tickets, err := s.createTickets(ctx, project.ID, assets)
	if err != nil {
		return err
	}
	if err := s.createTicketActivityExamples(ctx, tickets); err != nil {
		return err
	}
	if err := s.createSavedViews(ctx, project.ID); err != nil {
		return err
	}
	if err := s.createCronJob(ctx, project.ID, users["lead"].id()); err != nil {
		return err
	}
	if err := s.createIntegrationExamples(ctx, project.ID, users["lead"].id()); err != nil {
		return err
	}

	fmt.Fprintf(s.stdout, "demo project: key=%s id=%s\n", project.Key, project.ID)
	fmt.Fprintln(s.stdout, "demo seed completed")
	return nil
}

func (s *demoSeeder) login(ctx context.Context, username string, password string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/login", map[string]demoLoginRequest{
		"spec": {Username: username, Password: password},
	}, nil); err != nil {
		return err
	}
	for _, cookie := range s.client.Jar.Cookies(s.baseURL) {
		if cookie.Name == "rayboard_csrf" {
			s.csrf = cookie.Value
			return nil
		}
	}
	return fmt.Errorf("login succeeded but CSRF cookie was not set")
}

func (s *demoSeeder) createUsers(ctx context.Context) (map[string]demoCreatedUser, error) {
	inputs := map[string]map[string]any{
		"lead": {
			"username":     "demo_lead_" + strings.ToLower(s.suffix),
			"display_name": "Demo Lead " + s.suffix,
		},
		"engineer": {
			"username":     "demo_engineer_" + strings.ToLower(s.suffix),
			"display_name": "Demo Engineer " + s.suffix,
		},
		"reporter": {
			"username":     "demo_reporter_" + strings.ToLower(s.suffix),
			"display_name": "Demo Reporter " + s.suffix,
		},
	}

	users := make(map[string]demoCreatedUser, len(inputs))
	for key, input := range inputs {
		var created demoCreatedUser
		if err := s.apiJSON(ctx, http.MethodPost, "/api/users", map[string]any{"spec": input}, &created); err != nil {
			return nil, fmt.Errorf("create demo user %s: %w", key, err)
		}
		users[key] = created
		fmt.Fprintf(s.stdout, "demo user: role=%s username=%s password=%s\n", key, created.Spec.Username, created.Status.Password)
	}
	return users, nil
}

func (s *demoSeeder) createGroups(ctx context.Context, users map[string]demoCreatedUser) (map[string]demoGroup, error) {
	inputs := map[string]map[string]any{
		"engineers": {
			"name":         "demo_engineers_" + strings.ToLower(s.suffix),
			"display_name": "Demo Engineers " + s.suffix,
		},
		"stakeholders": {
			"name":         "demo_stakeholders_" + strings.ToLower(s.suffix),
			"display_name": "Demo Stakeholders " + s.suffix,
		},
	}

	groups := make(map[string]demoGroup, len(inputs))
	for key, input := range inputs {
		var group demoGroup
		if err := s.apiJSON(ctx, http.MethodPost, "/api/groups", map[string]any{"spec": input}, &group); err != nil {
			return nil, fmt.Errorf("create demo group %s: %w", key, err)
		}
		groups[key] = group
	}

	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["engineers"].id()+"/members/"+users["lead"].id(), nil, nil); err != nil {
		return nil, fmt.Errorf("add lead to engineers: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["engineers"].id()+"/members/"+users["engineer"].id(), nil, nil); err != nil {
		return nil, fmt.Errorf("add engineer to engineers: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["stakeholders"].id()+"/members/"+users["reporter"].id(), nil, nil); err != nil {
		return nil, fmt.Errorf("add reporter to stakeholders: %w", err)
	}
	return groups, nil
}

func (s *demoSeeder) createProject(ctx context.Context, leadUserID string) (tracker.Project, error) {
	input := tracker.CreateProjectInput{
		Key:         "DEMO" + s.suffix,
		Name:        "Demo Project " + s.suffix,
		Description: "Seeded project for demos",
		LeadUserID:  leadUserID,
	}
	var project demoProjectResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects", map[string]tracker.CreateProjectInput{"spec": input}, &project); err != nil {
		return tracker.Project{}, fmt.Errorf("create demo project: %w", err)
	}
	return tracker.Project{
		ID:          project.Metadata.ID,
		Key:         project.Spec.Key,
		Name:        project.Spec.Name,
		Description: project.Spec.Description,
		LeadUserID:  project.Spec.LeadUserID,
	}, nil
}

func (s *demoSeeder) bindProjectAccess(ctx context.Context, engineerGroupID string, stakeholderGroupID string, leadUserID string, projectID string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/role-bindings", map[string]demoRoleBinding{
		"spec": {
			RoleName:    authz.RoleProjectOwner,
			SubjectType: authz.BindingTargetUser,
			SubjectID:   leadUserID,
			Scope:       authz.ScopeKindProject,
			ProjectID:   projectID,
		},
	}, nil); err != nil {
		return fmt.Errorf("bind demo project owner: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/role-bindings", map[string]demoRoleBinding{
		"spec": {
			RoleName:    authz.RoleProjectMember,
			SubjectType: authz.BindingTargetGroup,
			SubjectID:   engineerGroupID,
			Scope:       authz.ScopeKindProject,
			ProjectID:   projectID,
		},
	}, nil); err != nil {
		return fmt.Errorf("bind demo project members: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/role-bindings", map[string]demoRoleBinding{
		"spec": {
			RoleName:    authz.RoleProjectViewer,
			SubjectType: authz.BindingTargetGroup,
			SubjectID:   stakeholderGroupID,
			Scope:       authz.ScopeKindProject,
			ProjectID:   projectID,
		},
	}, nil); err != nil {
		return fmt.Errorf("bind demo project viewers: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo role binding: project owner/member/viewer")
	return nil
}

type demoPlanningAssets struct {
	ComponentID string
	VersionID   string
	SprintID    string
}

func (s *demoSeeder) createProjectPlanning(ctx context.Context, projectID string, users map[string]demoCreatedUser) (demoPlanningAssets, error) {
	statuses := []tracker.ProjectStatusInput{
		{Slug: "todo", Name: "To Do"},
		{Slug: "in_progress", Name: "In Progress"},
		{Slug: "review", Name: "Review"},
		{Slug: "done", Name: "Done"},
	}
	if err := s.apiJSON(ctx, http.MethodPut, "/api/projects/"+projectID+"/statuses", map[string]any{
		"spec": map[string]any{"statuses": statuses},
	}, nil); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("replace demo workflow statuses: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo workflow: todo,in_progress,review,done")

	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/boards", map[string]any{
		"spec": map[string]any{
			"name":         "Delivery Board",
			"description":  "Seeded board for demos",
			"status_slugs": []string{"todo", "in_progress", "review", "done"},
		},
	}, nil); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo board: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo board: Delivery Board")

	var component demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/components", map[string]any{
		"spec": map[string]any{
			"name":                "Platform",
			"description":         "Backend and automation foundation",
			"owner_user_id":       users["lead"].id(),
			"default_assignee_id": users["engineer"].id(),
		},
	}, &component); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo component: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo component: Platform")

	var version demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/versions", map[string]any{
		"spec": map[string]any{
			"name":        "2026.7",
			"description": "Demo release target",
			"target_date": "2026-07-31",
		},
	}, &version); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo version: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo version: 2026.7")

	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/custom-fields", map[string]any{
		"spec": map[string]any{
			"key":        "severity",
			"name":       "Severity",
			"field_type": "single_select",
			"required":   true,
			"options":    []string{"Low", "Medium", "High"},
		},
	}, nil); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo custom field: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo custom field: severity")

	var sprint demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/sprints", map[string]any{
		"spec": map[string]any{
			"name":       "Sprint Demo",
			"goal":       "Show planning and automation surfaces",
			"start_date": "2026-06-17",
			"end_date":   "2026-07-01",
		},
	}, &sprint); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo sprint: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo sprint: Sprint Demo")

	return demoPlanningAssets{
		ComponentID: component.id(),
		VersionID:   version.id(),
		SprintID:    sprint.id(),
	}, nil
}

func (s *demoSeeder) createTicketHook(ctx context.Context, projectID string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/ticket-hooks", map[string]any{
		"spec": map[string]any{
			"name":     "demo-label",
			"event":    tracker.HookEventTicketCreate,
			"phase":    tracker.HookPhaseBefore,
			"enabled":  true,
			"position": 10,
			"engine": map[string]any{
				"type": tracker.HookEngineLua,
				"script": `
ticket.labels = ticket.labels or {}
ticket.labels[#ticket.labels + 1] = "demo-hook"
return { ticket = ticket }
`,
			},
		},
	}, nil); err != nil {
		return fmt.Errorf("create demo ticket hook: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo ticket hook: demo-label")
	return nil
}

func (s *demoSeeder) createTicketCreatePage(ctx context.Context, projectID string, users map[string]demoCreatedUser, assets demoPlanningAssets) error {
	slug := "bug-intake-" + strings.ToLower(s.suffix)
	var page demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/ticket-create-pages", map[string]any{
		"spec": map[string]any{
			"name":          "Bug Intake " + s.suffix,
			"slug":          slug,
			"description":   "Seeded intake page for customer-reported issues",
			"enabled":       true,
			"target_type":   "bug",
			"target_status": "todo",
			"owner_user_id": users["lead"].id(),
			"field_layout": []map[string]any{
				{"key": "title", "label": "Summary", "type": "text", "required": true},
				{"key": "description", "label": "Details", "type": "textarea", "required": true},
				{"key": "custom_fields.severity", "label": "Severity", "type": "single_select", "required": true},
			},
			"defaults": map[string]any{
				"priority":     "Medium",
				"labels":       []string{"demo", "intake"},
				"component_id": assets.ComponentID,
				"version_id":   assets.VersionID,
				"sprint_id":    assets.SprintID,
				"custom_fields": map[string]any{
					"severity": "Medium",
				},
			},
		},
	}, &page); err != nil {
		return fmt.Errorf("create demo ticket create page: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo ticket create page: slug=%s id=%s\n", slug, page.id())

	var ticket demoTicketResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/ticket-create-pages/"+slug+"/submit", map[string]any{
		"spec": map[string]any{
			"ticket": map[string]any{
				"title":       "Customer cannot complete SSO login",
				"description": "Submitted through the seeded custom create page.",
				"labels":      []string{"demo", "intake-submission"},
				"custom_fields": map[string]any{
					"severity": "High",
				},
			},
		},
	}, &ticket); err != nil {
		return fmt.Errorf("submit demo ticket create page: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo ticket create page submission: key=%s title=%q\n", ticket.Status.Key, ticket.Spec.Title)
	return nil
}

func (s *demoSeeder) createTickets(ctx context.Context, projectID string, assets demoPlanningAssets) ([]demoTicketResource, error) {
	epic, err := s.createDemoTicket(ctx, projectID, tracker.CreateTicketInput{
		Title:       "Launch customer onboarding",
		Description: "Epic for the seeded onboarding workflow.",
		Priority:    "High",
		Type:        "Epic",
		Status:      "todo",
		ComponentID: assets.ComponentID,
		VersionID:   assets.VersionID,
		StartDate:   "2026-06-17",
		DueDate:     "2026-07-31",
		Labels:      []string{"demo", "roadmap"},
		CustomFields: map[string]any{
			"severity": "High",
		},
	})
	if err != nil {
		return nil, err
	}

	tickets := []tracker.CreateTicketInput{
		{
			Title:          "Prepare customer onboarding board",
			Description:    "Create the initial backlog and workflow states.",
			Priority:       "High",
			Type:           "Task",
			ParentTicketID: epic.id(),
			SprintID:       assets.SprintID,
			ComponentID:    assets.ComponentID,
			VersionID:      assets.VersionID,
			Labels:         []string{"demo", "planning"},
			CustomFields:   map[string]any{"severity": "High"},
		},
		{
			Title:          "Investigate login audit trail",
			Description:    "Capture activity that helps explain auth changes.",
			Priority:       "Medium",
			Type:           "Bug",
			ParentTicketID: epic.id(),
			SprintID:       assets.SprintID,
			ComponentID:    assets.ComponentID,
			VersionID:      assets.VersionID,
			Labels:         []string{"demo", "auth"},
			CustomFields:   map[string]any{"severity": "Medium"},
		},
		{
			Title:          "Draft roadmap for automation hooks",
			Description:    "Outline Lua and AI hook milestones.",
			Priority:       "Low",
			Type:           "Story",
			ParentTicketID: epic.id(),
			SprintID:       assets.SprintID,
			ComponentID:    assets.ComponentID,
			VersionID:      assets.VersionID,
			Labels:         []string{"demo", "automation"},
			CustomFields:   map[string]any{"severity": "Low"},
		},
	}
	createdTickets := []demoTicketResource{epic}
	for index, ticket := range tickets {
		created, err := s.createDemoTicket(ctx, projectID, ticket)
		if err != nil {
			return nil, fmt.Errorf("create demo ticket %d: %w", index+1, err)
		}
		createdTickets = append(createdTickets, created)
		if index == 1 {
			status := "in_progress"
			update := tracker.UpdateTicketInput{Status: &status}
			if err := s.apiJSON(ctx, http.MethodPatch, "/api/tickets/"+created.Metadata.ID, map[string]tracker.UpdateTicketInput{"spec": update}, nil); err != nil {
				return nil, fmt.Errorf("update demo ticket %s: %w", created.Status.Key, err)
			}
		}
	}
	reordered := make([]string, 0, len(createdTickets))
	for index := len(createdTickets) - 1; index >= 0; index-- {
		reordered = append(reordered, createdTickets[index].id())
	}
	if err := s.apiJSON(ctx, http.MethodPatch, "/api/projects/"+projectID+"/backlog", map[string]any{
		"spec": map[string]any{"ticket_ids": reordered},
	}, nil); err != nil {
		return nil, fmt.Errorf("reorder demo backlog: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo backlog: reordered")
	return createdTickets, nil
}

func (s *demoSeeder) createTicketActivityExamples(ctx context.Context, tickets []demoTicketResource) error {
	if len(tickets) == 0 {
		return nil
	}
	ticket := tickets[0]
	if err := s.apiJSON(ctx, http.MethodPost, "/api/tickets/"+ticket.id()+"/comments", map[string]any{
		"spec": map[string]any{
			"body": "Demo comment: link the onboarding epic to the release notes and customer rollout plan.",
		},
	}, nil); err != nil {
		return fmt.Errorf("create demo comment: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo comment: ticket=%s\n", ticket.Status.Key)

	if err := s.uploadAttachment(ctx, ticket.id(), "demo-release-notes.txt", "text/plain", []byte("Demo release notes for Rayboard onboarding.\nSearchable attachment content: rollout checklist.\n")); err != nil {
		return fmt.Errorf("upload demo attachment: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo attachment: ticket=%s file=demo-release-notes.txt\n", ticket.Status.Key)
	return nil
}

func (s *demoSeeder) createSavedViews(ctx context.Context, projectID string) error {
	var view demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/saved-views", map[string]any{
		"spec": map[string]any{
			"project_id":   projectID,
			"scope_type":   search.SavedViewScopeProject,
			"name":         "Pinned demo backlog",
			"query":        map[string]any{"filter": `labels == "demo"`, "text": "onboarding"},
			"sort":         []map[string]any{{"field": "priority", "direction": search.SortDirectionDesc}},
			"columns":      []string{"key", "title", "status", "priority", "assignee_id"},
			"display_mode": search.SavedViewDisplayBacklog,
			"group_by":     "status",
			"pinned":       true,
		},
	}, &view); err != nil {
		return fmt.Errorf("create demo saved view: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo saved view: name=%q id=%s\n", "Pinned demo backlog", view.id())

	if err := s.apiJSON(ctx, http.MethodPost, "/api/search", map[string]any{
		"spec": map[string]any{
			"project_id": projectID,
			"filter":     `labels == "demo"`,
			"text":       "rollout",
			"limit":      10,
		},
	}, nil); err != nil {
		return fmt.Errorf("run demo search example: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo search: CEL labels filter with FTS text")
	return nil
}

func (s *demoSeeder) createCronJob(ctx context.Context, projectID string, ownerUserID string) error {
	var job demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/cron-jobs", map[string]any{
		"spec": map[string]any{
			"owner_user_id": ownerUserID,
			"project_id":    projectID,
			"name":          "Demo stale ticket scan",
			"schedule":      "0 9 * * 1",
			"timezone":      "UTC",
			"enabled":       false,
			"engine": map[string]any{
				"type": "lua",
				"script": `
local results = rayboard.search({ filter = 'labels == "demo"', text = 'onboarding', limit = 5 })
return { checked = true, ticket_count = #(results.items or {}) }
`,
			},
		},
	}, &job); err != nil {
		return fmt.Errorf("create demo cron job: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo cron job: name=%q id=%s enabled=false\n", "Demo stale ticket scan", job.id())
	return nil
}

func (s *demoSeeder) createIntegrationExamples(ctx context.Context, projectID string, actorUserID string) error {
	if err := s.createWebhookExamples(ctx, projectID, actorUserID); err != nil {
		return err
	}
	if err := s.createNotificationExamples(ctx, projectID, actorUserID); err != nil {
		return err
	}
	return nil
}

func (s *demoSeeder) createWebhookExamples(ctx context.Context, projectID string, actorUserID string) error {
	var incoming demoCreatedWebhookResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/webhooks", map[string]any{
		"spec": map[string]any{
			"name":          "Demo incoming triage",
			"direction":     webhooks.DirectionIncoming,
			"enabled":       true,
			"actor_user_id": actorUserID,
			"engine": map[string]any{
				"type": webhooks.EngineTypeLua,
				"script": `
rayboard.log("demo incoming webhook accepted")
return { ok = true, actions = {} }
`,
			},
		},
	}, &incoming); err != nil {
		return fmt.Errorf("create demo incoming webhook: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo incoming webhook: id=%s token=%s\n", incoming.id(), incoming.Status.Token)

	var outgoing demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/webhooks", map[string]any{
		"spec": map[string]any{
			"name":          "Demo outgoing ticket update",
			"direction":     webhooks.DirectionOutgoing,
			"enabled":       false,
			"actor_user_id": actorUserID,
			"event_types":   []string{"ticket.updated"},
			"engine": map[string]any{
				"type": webhooks.EngineTypeLua,
				"script": `
return {
  method = "POST",
  path = "/demo/events",
  headers = { ["X-Rayboard-Demo"] = "true" },
  body = { event_type = event.type, subject_id = event.subject_id }
}
`,
			},
		},
	}, &outgoing); err != nil {
		return fmt.Errorf("create demo outgoing webhook: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo outgoing webhook: id=%s enabled=false\n", outgoing.id())
	return nil
}

func (s *demoSeeder) createNotificationExamples(ctx context.Context, projectID string, actorUserID string) error {
	var destination demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/notification-destinations", map[string]any{
		"spec": map[string]any{
			"name":         "Demo logger destination",
			"shoutrrr_url": "logger://",
			"enabled":      true,
		},
	}, &destination); err != nil {
		return fmt.Errorf("create demo notification destination: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo notification destination: id=%s type=logger\n", destination.id())

	disabled := false
	var policy demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/notification-policies", map[string]any{
		"spec": map[string]any{
			"name":            "Demo comment notifications",
			"event_types":     []string{"comment_added"},
			"destination_ids": []string{destination.id()},
			"enabled":         disabled,
		},
	}, &policy); err != nil {
		return fmt.Errorf("create demo notification policy: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo notification policy: id=%s enabled=false\n", policy.id())

	var hook demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/notification-hooks", map[string]any{
		"spec": map[string]any{
			"name":          "Demo notification annotator",
			"actor_user_id": actorUserID,
			"event_types":   []string{"comment_added"},
			"enabled":       disabled,
			"engine": map[string]any{
				"type": notifications.HookEngineLua,
				"script": `
notification.plan.message = "[demo] " .. notification.plan.message
return { message = notification.plan.message, payload = notification.plan.payload }
`,
			},
		},
	}, &hook); err != nil {
		return fmt.Errorf("create demo notification hook: %w", err)
	}
	fmt.Fprintf(s.stdout, "demo notification hook: id=%s enabled=false\n", hook.id())
	return nil
}

func (s *demoSeeder) createDemoTicket(ctx context.Context, projectID string, ticket tracker.CreateTicketInput) (demoTicketResource, error) {
	var created demoTicketResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/tickets", map[string]tracker.CreateTicketInput{"spec": ticket}, &created); err != nil {
		return demoTicketResource{}, err
	}
	fmt.Fprintf(s.stdout, "demo ticket: key=%s title=%q\n", created.Status.Key, created.Spec.Title)
	return created, nil
}

func (s *demoSeeder) apiJSON(ctx context.Context, method string, path string, input any, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	target := s.baseURL.ResolveReference(&url.URL{Path: path})
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if s.csrf != "" && mutatingMethod(method) {
		request.Header.Set("X-CSRF-Token", s.csrf)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("%s %s returned %d: %s", method, path, response.StatusCode, strings.TrimSpace(string(data)))
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (s *demoSeeder) uploadAttachment(ctx context.Context, ticketID string, fileName string, contentType string, data []byte) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeQuotes(fileName)))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("create multipart file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("write multipart file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart body: %w", err)
	}

	target := s.baseURL.ResolveReference(&url.URL{Path: "/api/tickets/" + ticketID + "/attachments"})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), &body)
	if err != nil {
		return fmt.Errorf("build attachment request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", s.csrf)

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("POST attachment: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseData, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("POST attachment returned %d: %s", response.StatusCode, strings.TrimSpace(string(responseData)))
	}
	return nil
}

func escapeQuotes(value string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(value)
}

func randomDemoSuffix() (string, error) {
	var raw [3]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate demo suffix: %w", err)
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(raw[:]), "="), nil
}

func mutatingMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
